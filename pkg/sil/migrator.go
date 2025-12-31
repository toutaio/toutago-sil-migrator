package sil

import (
	"context"
	"fmt"
	"time"
)

// migrator implements the Migrator interface.
type migrator struct {
	adapter       DatabaseAdapter
	config        *Config
	loader        *Loader
	logger        Logger
	beforeMigrate MigrationCallback
	afterMigrate  MigrationCallback
	onError       MigrationErrorCallback
}

// NewMigrator creates a new migrator instance.
func NewMigrator(config *Config, adapter DatabaseAdapter) (Migrator, error) {
	if config == nil {
		return nil, ErrInvalidConfiguration("config is nil")
	}

	if adapter == nil {
		return nil, fmt.Errorf("adapter is nil")
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, err
	}

	logger := NewDefaultLogger(config.Verbose)
	if config.Verbose {
		logger = NewColorLogger(config.Verbose)
	}

	loader := NewLoader(config.MigrationsDir, logger)

	return &migrator{
		adapter: adapter,
		config:  config,
		loader:  loader,
		logger:  logger,
	}, nil
}

// Migrate runs all pending migrations.
func (m *migrator) Migrate(ctx context.Context) error {
	m.logger.Info("Starting migration...")

	// Acquire lock
	lock, err := m.adapter.Lock(ctx)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrLockAcquisitionFailed, err)
	}
	defer func() { _ = lock.Release() }()

	m.logger.Debug("Migration lock acquired")

	// Ensure migrations table exists
	if err := m.adapter.CreateMigrationsTable(ctx); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Get applied migrations
	applied, err := m.adapter.GetAppliedMigrations(ctx)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	// Load pending migrations
	pending, err := m.loader.LoadPending(applied)
	if err != nil {
		if err == ErrNoMigrationsFound {
			m.logger.Warn("No migrations found")
			return nil
		}
		return err
	}

	if len(pending) == 0 {
		m.logger.Info("No pending migrations")
		return nil
	}

	m.logger.Info("Found %d pending migrations", len(pending))

	// Get next batch number
	batch, err := m.adapter.GetLastBatch(ctx)
	if err != nil {
		return fmt.Errorf("failed to get last batch: %w", err)
	}
	batch++ // Increment for new batch

	// Run migrations
	for i, migration := range pending {
		m.logger.Info("[%d/%d] Migrating: %s - %s", i+1, len(pending), migration.Version(), migration.Description())

		if err := m.runMigration(ctx, migration, "up", batch); err != nil {
			return err
		}

		m.logger.Info("✓ Migrated: %s", migration.Version())
	}

	m.logger.Info("Migration complete! Ran %d migrations", len(pending))
	return nil
}

// MigrateUp runs the next N pending migrations.
func (m *migrator) MigrateUp(ctx context.Context, steps int) error {
	if steps <= 0 {
		return fmt.Errorf("steps must be positive, got %d", steps)
	}

	m.logger.Info("Migrating up %d steps...", steps)

	// Acquire lock
	lock, err := m.adapter.Lock(ctx)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrLockAcquisitionFailed, err)
	}
	defer func() { _ = lock.Release() }()

	// Ensure migrations table exists
	if err := m.adapter.CreateMigrationsTable(ctx); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Get applied migrations
	applied, err := m.adapter.GetAppliedMigrations(ctx)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	// Load pending migrations
	pending, err := m.loader.LoadPending(applied)
	if err != nil {
		return err
	}

	if len(pending) == 0 {
		m.logger.Info("No pending migrations")
		return nil
	}

	// Limit to requested steps
	if steps > len(pending) {
		steps = len(pending)
	}
	pending = pending[:steps]

	// Get next batch number
	batch, err := m.adapter.GetLastBatch(ctx)
	if err != nil {
		return fmt.Errorf("failed to get last batch: %w", err)
	}
	batch++

	// Run migrations
	for i, migration := range pending {
		m.logger.Info("[%d/%d] Migrating: %s - %s", i+1, len(pending), migration.Version(), migration.Description())

		if err := m.runMigration(ctx, migration, "up", batch); err != nil {
			return err
		}

		m.logger.Info("✓ Migrated: %s", migration.Version())
	}

	m.logger.Info("Migration complete! Ran %d migrations", len(pending))
	return nil
}

// MigrateDown rolls back the last N migrations.
func (m *migrator) MigrateDown(ctx context.Context, steps int) error {
	if steps <= 0 {
		return fmt.Errorf("steps must be positive, got %d", steps)
	}

	m.logger.Info("Rolling back %d migrations...", steps)

	// Acquire lock
	lock, err := m.adapter.Lock(ctx)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrLockAcquisitionFailed, err)
	}
	defer func() { _ = lock.Release() }()

	// Get applied migrations
	applied, err := m.adapter.GetAppliedMigrations(ctx)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	if len(applied) == 0 {
		m.logger.Info("No migrations to roll back")
		return nil
	}

	// Load all migrations
	all, err := m.loader.Load()
	if err != nil {
		return err
	}

	// Get applied migrations objects
	appliedMigrations := FilterAppliedMigrations(all, applied)

	// Sort in reverse order
	SortMigrationsDescending(appliedMigrations)

	// Limit to requested steps
	if steps > len(appliedMigrations) {
		steps = len(appliedMigrations)
	}
	toRollback := appliedMigrations[:steps]

	// Roll back migrations
	for i, migration := range toRollback {
		m.logger.Info("[%d/%d] Rolling back: %s - %s", i+1, len(toRollback), migration.Version(), migration.Description())

		if err := m.runMigration(ctx, migration, "down", 0); err != nil {
			return err
		}

		m.logger.Info("✓ Rolled back: %s", migration.Version())
	}

	m.logger.Info("Rollback complete! Rolled back %d migrations", len(toRollback))
	return nil
}

// Rollback rolls back the last batch of migrations.
func (m *migrator) Rollback(ctx context.Context) error {
	m.logger.Info("Rolling back last batch...")

	// Acquire lock
	lock, err := m.adapter.Lock(ctx)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrLockAcquisitionFailed, err)
	}
	defer func() { _ = lock.Release() }()

	// Get applied migrations
	applied, err := m.adapter.GetAppliedMigrations(ctx)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	if len(applied) == 0 {
		m.logger.Info("No migrations to roll back")
		return nil
	}

	// Get last batch
	lastBatch, err := m.adapter.GetLastBatch(ctx)
	if err != nil {
		return fmt.Errorf("failed to get last batch: %w", err)
	}

	// Load all migrations
	all, err := m.loader.Load()
	if err != nil {
		return err
	}

	// Get migrations from last batch
	batchMigrations := GetMigrationsByBatch(all, applied, lastBatch)

	// Sort in reverse order
	SortMigrationsDescending(batchMigrations)

	if len(batchMigrations) == 0 {
		m.logger.Info("No migrations in last batch")
		return nil
	}

	m.logger.Info("Rolling back batch %d (%d migrations)", lastBatch, len(batchMigrations))

	// Roll back migrations
	for i, migration := range batchMigrations {
		m.logger.Info("[%d/%d] Rolling back: %s - %s", i+1, len(batchMigrations), migration.Version(), migration.Description())

		if err := m.runMigration(ctx, migration, "down", 0); err != nil {
			return err
		}

		m.logger.Info("✓ Rolled back: %s", migration.Version())
	}

	m.logger.Info("Rollback complete! Rolled back %d migrations", len(batchMigrations))
	return nil
}

// Status returns the status of all migrations.
func (m *migrator) Status(ctx context.Context) ([]MigrationStatus, error) {
	m.logger.Debug("Getting migration status...")

	// Ensure migrations table exists
	if err := m.adapter.CreateMigrationsTable(ctx); err != nil {
		return nil, fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Get applied migrations
	applied, err := m.adapter.GetAppliedMigrations(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get applied migrations: %w", err)
	}

	// Load all migrations
	all, err := m.loader.Load()
	if err != nil {
		if err == ErrNoMigrationsFound {
			return []MigrationStatus{}, nil
		}
		return nil, err
	}

	// Get status
	statuses := GetMigrationStatus(all, applied)

	return statuses, nil
}

// Reset rolls back all migrations.
func (m *migrator) Reset(ctx context.Context) error {
	m.logger.Info("Resetting all migrations...")

	// Acquire lock
	lock, err := m.adapter.Lock(ctx)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrLockAcquisitionFailed, err)
	}
	defer func() { _ = lock.Release() }()

	// Get applied migrations
	applied, err := m.adapter.GetAppliedMigrations(ctx)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	if len(applied) == 0 {
		m.logger.Info("No migrations to reset")
		return nil
	}

	// Load all migrations
	all, err := m.loader.Load()
	if err != nil {
		return err
	}

	// Get applied migrations objects
	appliedMigrations := FilterAppliedMigrations(all, applied)

	// Sort in reverse order
	SortMigrationsDescending(appliedMigrations)

	m.logger.Info("Resetting %d migrations", len(appliedMigrations))

	// Roll back all migrations
	for i, migration := range appliedMigrations {
		m.logger.Info("[%d/%d] Rolling back: %s - %s", i+1, len(appliedMigrations), migration.Version(), migration.Description())

		if err := m.runMigration(ctx, migration, "down", 0); err != nil {
			return err
		}

		m.logger.Info("✓ Rolled back: %s", migration.Version())
	}

	m.logger.Info("Reset complete! Rolled back %d migrations", len(appliedMigrations))
	return nil
}

// Refresh rolls back all migrations and re-runs them.
func (m *migrator) Refresh(ctx context.Context) error {
	m.logger.Info("Refreshing migrations...")

	// Reset all migrations
	if err := m.Reset(ctx); err != nil {
		return fmt.Errorf("failed to reset migrations: %w", err)
	}

	// Run all migrations
	if err := m.Migrate(ctx); err != nil {
		return fmt.Errorf("failed to migrate: %w", err)
	}

	m.logger.Info("Refresh complete!")
	return nil
}

// runMigration runs a single migration with transaction support.
func (m *migrator) runMigration(ctx context.Context, migration Migration, direction string, batch int) error {
	// Create context with timeout
	migrationCtx, cancel := context.WithTimeout(ctx, m.config.MigrationTimeout)
	defer cancel()

	// Call before callback
	if m.beforeMigrate != nil {
		if err := m.beforeMigrate(migration, direction); err != nil {
			return fmt.Errorf("before migration callback failed: %w", err)
		}
	}

	// Begin transaction
	tx, err := m.adapter.BeginTx(migrationCtx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Track whether we need to rollback
	needsRollback := true
	defer func() {
		if needsRollback {
			if rbErr := tx.Rollback(); rbErr != nil {
				m.logger.Error("Failed to rollback transaction: %v", rbErr)
			}
		}
	}()

	start := time.Now()

	// Run migration with panic recovery
	var migrationErr error
	func() {
		defer func() {
			if r := recover(); r != nil {
				migrationErr = fmt.Errorf("migration panicked: %v", r)
			}
		}()

		if direction == "up" {
			migrationErr = migration.Up(m.adapter)
		} else {
			migrationErr = migration.Down(m.adapter)
		}
	}()

	elapsed := time.Since(start)

	if migrationErr != nil {
		m.logger.Error("Migration failed after %v: %v", elapsed, migrationErr)

		// Call error callback
		if m.onError != nil {
			m.onError(migration, migrationErr)
		}

		return WrapMigrationError(migration.Version(), direction, migrationErr)
	}

	// Record or remove migration
	if direction == "up" {
		if err := m.adapter.RecordMigration(migrationCtx, migration.Version(), migration.Description(), batch); err != nil {
			return fmt.Errorf("failed to record migration: %w", err)
		}
	} else {
		if err := m.adapter.RemoveMigration(migrationCtx, migration.Version()); err != nil {
			return fmt.Errorf("failed to remove migration record: %w", err)
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	needsRollback = false

	m.logger.Debug("Migration %s completed in %v", migration.Version(), elapsed)

	// Call after callback
	if m.afterMigrate != nil {
		if err := m.afterMigrate(migration, direction); err != nil {
			m.logger.Warn("After migration callback failed: %v", err)
		}
	}

	return nil
}

// SetBeforeMigrate sets the callback to run before each migration.
func (m *migrator) SetBeforeMigrate(callback MigrationCallback) {
	m.beforeMigrate = callback
}

// SetAfterMigrate sets the callback to run after each migration.
func (m *migrator) SetAfterMigrate(callback MigrationCallback) {
	m.afterMigrate = callback
}

// SetOnError sets the callback to run when a migration fails.
func (m *migrator) SetOnError(callback MigrationErrorCallback) {
	m.onError = callback
}

// SetLogger sets a custom logger.
func (m *migrator) SetLogger(logger Logger) {
	m.logger = logger
	m.loader.logger = logger
}
