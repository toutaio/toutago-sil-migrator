package sil

import (
	"context"
	"fmt"
	"time"
)

// seedManager implements the SeedManager interface.
type seedManager struct {
	adapter     DatabaseAdapter
	config      *Config
	logger      Logger
	environment string
}

// NewSeedManager creates a new seed manager.
func NewSeedManager(config *Config, adapter DatabaseAdapter) (SeedManager, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	if adapter == nil {
		return nil, fmt.Errorf("adapter cannot be nil")
	}

	return &seedManager{
		adapter:     adapter,
		config:      config,
		logger:      NewDefaultLogger(config.Verbose),
		environment: config.Environment,
	}, nil
}

// SetLogger sets a custom logger.
func (sm *seedManager) SetLogger(logger Logger) {
	sm.logger = logger
}

// Seed runs the specified seeders.
func (sm *seedManager) Seed(ctx context.Context, seeders ...string) error {
	sm.logger.Info("Starting seeding...")

	// Create seeds tracking table
	if err := sm.createSeedsTable(ctx); err != nil {
		return fmt.Errorf("failed to create seeds table: %w", err)
	}

	// Get seeders to run
	var toRun []Seeder
	if len(seeders) == 0 {
		// If no seeders specified, run all
		return sm.SeedAll(ctx)
	}

	// Get specific seeders
	for _, name := range seeders {
		seeder, exists := GetSeederByName(name)
		if !exists {
			return fmt.Errorf("seeder not found: %s", name)
		}
		toRun = append(toRun, seeder)
	}

	return sm.runSeeders(ctx, toRun)
}

// SeedAll runs all registered seeders.
func (sm *seedManager) SeedAll(ctx context.Context) error {
	sm.logger.Info("Running all seeders...")

	// Create seeds tracking table
	if err := sm.createSeedsTable(ctx); err != nil {
		return fmt.Errorf("failed to create seeds table: %w", err)
	}

	// Get all registered seeders
	seeders := GetRegisteredSeeders()
	if len(seeders) == 0 {
		sm.logger.Warn("No seeders found in registry")
		return nil
	}

	sm.logger.Info("Loaded %d seeders", len(seeders))

	return sm.runSeeders(ctx, seeders)
}

// runSeeders executes the given seeders in dependency order.
func (sm *seedManager) runSeeders(ctx context.Context, seeders []Seeder) error {
	// Filter by environment
	filtered := sm.filterByEnvironment(seeders)
	if len(filtered) == 0 {
		sm.logger.Info("No seeders to run for environment: %s", sm.environment)
		return nil
	}

	sm.logger.Info("Found %d seeders for environment: %s", len(filtered), sm.environment)

	// Sort by dependencies
	sorted, err := SortSeedersByDependencies(filtered)
	if err != nil {
		return fmt.Errorf("failed to resolve dependencies: %w", err)
	}

	// Check which have already been executed
	executed, err := sm.getExecutedSeeders(ctx)
	if err != nil {
		return fmt.Errorf("failed to get executed seeders: %w", err)
	}

	executedMap := make(map[string]bool)
	for _, record := range executed {
		executedMap[record.Name] = true
	}

	// Run seeders
	runCount := 0
	skipCount := 0

	for i, seeder := range sorted {
		name := seeder.Name()

		sm.logger.Info("[%d/%d] Seeding: %s", i+1, len(sorted), name)

		// Check if already executed
		if executedMap[name] {
			sm.logger.Info("✓ Skipped (already executed): %s", name)
			skipCount++
			continue
		}

		// Check if should run
		shouldRun, err := seeder.ShouldRun(ctx, sm.adapter)
		if err != nil {
			return fmt.Errorf("failed to check if seeder %s should run: %w", name, err)
		}

		if !shouldRun {
			sm.logger.Info("✓ Skipped (condition not met): %s", name)
			skipCount++
			continue
		}

		// Run seeder
		startTime := time.Now()
		if err := seeder.Seed(ctx, sm.adapter); err != nil {
			sm.logger.Error("✗ Seeder failed: %s - %v", name, err)
			return fmt.Errorf("seeder %s failed: %w", name, err)
		}

		// Record execution
		if err := sm.recordSeeder(ctx, name); err != nil {
			return fmt.Errorf("failed to record seeder %s: %w", name, err)
		}

		duration := time.Since(startTime)
		sm.logger.Info("✓ Seeded: %s (took %v)", name, duration)
		runCount++
	}

	sm.logger.Info("Seeding complete! Ran %d seeders, skipped %d", runCount, skipCount)
	return nil
}

// Status returns the status of all seeders.
func (sm *seedManager) Status(ctx context.Context) ([]SeedStatus, error) {
	// Create seeds table if it doesn't exist
	if err := sm.createSeedsTable(ctx); err != nil {
		return nil, fmt.Errorf("failed to create seeds table: %w", err)
	}

	// Get all registered seeders
	seeders := GetRegisteredSeeders()
	
	// Get executed seeders
	executed, err := sm.getExecutedSeeders(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get executed seeders: %w", err)
	}

	executedMap := make(map[string]SeedRecord)
	for _, record := range executed {
		executedMap[record.Name] = record
	}

	// Build status list
	var statuses []SeedStatus
	for _, seeder := range seeders {
		name := seeder.Name()
		record, executed := executedMap[name]

		// Check environment match
		envMatch := sm.matchesEnvironment(seeder)
		skipReason := ""
		if !envMatch {
			skipReason = fmt.Sprintf("environment mismatch (requires: %v, current: %s)", 
				seeder.Environments(), sm.environment)
		}

		var executedAt *time.Time
		if executed {
			executedAt = &record.ExecutedAt
		}

		status := SeedStatus{
			Name:        name,
			Executed:    executed,
			ExecutedAt:  executedAt,
			Environment: record.Environment,
			Skipped:     !envMatch,
			Reason:      skipReason,
		}

		statuses = append(statuses, status)
	}

	// Sort by dependency order
	sorted, err := SortSeedersByDependencies(seeders)
	if err != nil {
		// If we can't sort, just return unsorted
		return statuses, nil
	}

	// Reorder statuses to match sorted order
	statusMap := make(map[string]SeedStatus)
	for _, status := range statuses {
		statusMap[status.Name] = status
	}

	var sortedStatuses []SeedStatus
	for _, seeder := range sorted {
		sortedStatuses = append(sortedStatuses, statusMap[seeder.Name()])
	}

	return sortedStatuses, nil
}

// filterByEnvironment filters seeders that should run in the current environment.
func (sm *seedManager) filterByEnvironment(seeders []Seeder) []Seeder {
	var filtered []Seeder
	for _, seeder := range seeders {
		if sm.matchesEnvironment(seeder) {
			filtered = append(filtered, seeder)
		}
	}
	return filtered
}

// matchesEnvironment checks if a seeder should run in the current environment.
func (sm *seedManager) matchesEnvironment(seeder Seeder) bool {
	envs := seeder.Environments()
	
	// If no environments specified, run in all environments
	if len(envs) == 0 {
		return true
	}

	// Check if current environment matches
	for _, env := range envs {
		if env == sm.environment {
			return true
		}
	}

	return false
}

// createSeedsTable creates the seeds tracking table.
func (sm *seedManager) createSeedsTable(ctx context.Context) error {
	query := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) NOT NULL UNIQUE,
			environment VARCHAR(100),
			executed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`, sm.config.SeedsTableName)

	return sm.adapter.Exec(ctx, query)
}

// getExecutedSeeders returns all executed seeders.
func (sm *seedManager) getExecutedSeeders(ctx context.Context) ([]SeedRecord, error) {
	query := fmt.Sprintf(`
		SELECT id, name, environment, executed_at
		FROM %s
		ORDER BY id ASC
	`, sm.config.SeedsTableName)

	rows, err := sm.adapter.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []SeedRecord
	for rows.Next() {
		var record SeedRecord
		if err := rows.Scan(&record.ID, &record.Name, &record.Environment, &record.ExecutedAt); err != nil {
			return nil, err
		}
		records = append(records, record)
	}

	return records, rows.Err()
}

// recordSeeder records a seeder as executed.
func (sm *seedManager) recordSeeder(ctx context.Context, name string) error {
	query := fmt.Sprintf(`
		INSERT INTO %s (name, environment)
		VALUES ($1, $2)
	`, sm.config.SeedsTableName)

	return sm.adapter.Exec(ctx, query, name, sm.environment)
}
