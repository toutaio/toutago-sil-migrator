package sil

import (
	"fmt"
	"regexp"
	"sort"
	"time"
)

// BaseMigration provides a base implementation of the Migration interface.
type BaseMigration struct {
	version     string
	description string
	upFunc      MigrationFunc
	downFunc    MigrationFunc
}

// NewBaseMigration creates a new base migration.
func NewBaseMigration(version, description string, upFunc, downFunc MigrationFunc) *BaseMigration {
	return &BaseMigration{
		version:     version,
		description: description,
		upFunc:      upFunc,
		downFunc:    downFunc,
	}
}

// Version returns the migration version.
func (m *BaseMigration) Version() string {
	return m.version
}

// Description returns the migration description.
func (m *BaseMigration) Description() string {
	return m.description
}

// Up applies the migration.
func (m *BaseMigration) Up(adapter DatabaseAdapter) error {
	if m.upFunc == nil {
		return fmt.Errorf("migration %s has no Up function", m.version)
	}
	return m.upFunc(adapter)
}

// Down reverts the migration.
func (m *BaseMigration) Down(adapter DatabaseAdapter) error {
	if m.downFunc == nil {
		return fmt.Errorf("migration %s has no Down function", m.version)
	}
	return m.downFunc(adapter)
}

// migrationRegistry holds all registered migrations.
var migrationRegistry = make(map[string]Migration)

// RegisterMigration registers a migration.
func RegisterMigration(migration Migration) {
	version := migration.Version()
	if _, exists := migrationRegistry[version]; exists {
		panic(fmt.Sprintf("migration %s already registered", version))
	}
	migrationRegistry[version] = migration
}

// GetRegisteredMigrations returns all registered migrations sorted by version.
func GetRegisteredMigrations() []Migration {
	migrations := make([]Migration, 0, len(migrationRegistry))
	for _, migration := range migrationRegistry {
		migrations = append(migrations, migration)
	}

	// Sort by version
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version() < migrations[j].Version()
	})

	return migrations
}

// ClearRegisteredMigrations clears all registered migrations (useful for testing).
func ClearRegisteredMigrations() {
	migrationRegistry = make(map[string]Migration)
}

// ValidateVersion validates a migration version format.
// Expected format: YYYYMMDDHHMMSS
func ValidateVersion(version string) error {
	// Version must be exactly 14 digits
	re := regexp.MustCompile(`^\d{14}$`)
	if !re.MatchString(version) {
		return fmt.Errorf("%w: %s (expected format: YYYYMMDDHHMMSS)", ErrInvalidMigrationVersion, version)
	}

	// Parse as timestamp to ensure it's a valid date/time
	_, err := time.Parse("20060102150405", version)
	if err != nil {
		return fmt.Errorf("%w: %s is not a valid timestamp", ErrInvalidMigrationVersion, version)
	}

	return nil
}

// GenerateVersion generates a new migration version based on current time.
func GenerateVersion() string {
	return time.Now().Format("20060102150405")
}

// SortMigrations sorts migrations by version in ascending order.
func SortMigrations(migrations []Migration) {
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version() < migrations[j].Version()
	})
}

// SortMigrationsDescending sorts migrations by version in descending order.
func SortMigrationsDescending(migrations []Migration) {
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version() > migrations[j].Version()
	})
}

// FilterPendingMigrations returns migrations that have not been applied.
func FilterPendingMigrations(all []Migration, applied []MigrationRecord) []Migration {
	appliedMap := make(map[string]bool)
	for _, record := range applied {
		appliedMap[record.Version] = true
	}

	pending := make([]Migration, 0)
	for _, migration := range all {
		if !appliedMap[migration.Version()] {
			pending = append(pending, migration)
		}
	}

	return pending
}

// FilterAppliedMigrations returns migrations that have been applied.
func FilterAppliedMigrations(all []Migration, applied []MigrationRecord) []Migration {
	appliedMap := make(map[string]MigrationRecord)
	for _, record := range applied {
		appliedMap[record.Version] = record
	}

	appliedMigrations := make([]Migration, 0)
	for _, migration := range all {
		if _, exists := appliedMap[migration.Version()]; exists {
			appliedMigrations = append(appliedMigrations, migration)
		}
	}

	return appliedMigrations
}

// GetMigrationsByBatch returns migrations from a specific batch.
func GetMigrationsByBatch(migrations []Migration, applied []MigrationRecord, batch int) []Migration {
	batchVersions := make(map[string]bool)
	for _, record := range applied {
		if record.Batch == batch {
			batchVersions[record.Version] = true
		}
	}

	batchMigrations := make([]Migration, 0)
	for _, migration := range migrations {
		if batchVersions[migration.Version()] {
			batchMigrations = append(batchMigrations, migration)
		}
	}

	return batchMigrations
}

// GetMigrationStatus returns the status of all migrations.
func GetMigrationStatus(all []Migration, applied []MigrationRecord) []MigrationStatus {
	appliedMap := make(map[string]MigrationRecord)
	for _, record := range applied {
		appliedMap[record.Version] = record
	}

	statuses := make([]MigrationStatus, 0, len(all))
	for _, migration := range all {
		status := MigrationStatus{
			Version:     migration.Version(),
			Description: migration.Description(),
			Applied:     false,
		}

		if record, exists := appliedMap[migration.Version()]; exists {
			status.Applied = true
			status.Batch = record.Batch
			status.ExecutedAt = &record.ExecutedAt
		}

		statuses = append(statuses, status)
	}

	return statuses
}
