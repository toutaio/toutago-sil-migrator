package sil

import (
	"context"
	"database/sql"
)

// Migration represents a database migration with Up and Down operations.
// Each migration must have a unique version (timestamp) and description.
type Migration interface {
	// Version returns the migration version (format: YYYYMMDDHHMMSS)
	Version() string

	// Description returns a human-readable description of the migration
	Description() string

	// Up applies the migration changes to the database
	Up(adapter DatabaseAdapter) error

	// Down reverts the migration changes
	Down(adapter DatabaseAdapter) error
}

// DatabaseAdapter defines the interface for database-specific operations.
// This abstraction allows Síl to work with different databases through
// pluggable adapters.
type DatabaseAdapter interface {
	// Connect establishes a connection to the database
	Connect(ctx context.Context, config *Config) error

	// Close closes the database connection
	Close() error

	// Exec executes a SQL statement without returning rows
	Exec(ctx context.Context, query string, args ...interface{}) error

	// Query executes a SQL query and returns rows
	Query(ctx context.Context, query string, args ...interface{}) (Rows, error)

	// BeginTx starts a new transaction
	BeginTx(ctx context.Context) (Transaction, error)

	// CreateMigrationsTable creates the migrations tracking table if it doesn't exist
	CreateMigrationsTable(ctx context.Context) error

	// GetAppliedMigrations returns all migrations that have been applied
	GetAppliedMigrations(ctx context.Context) ([]MigrationRecord, error)

	// RecordMigration records a migration as applied
	RecordMigration(ctx context.Context, version, description string, batch int) error

	// RemoveMigration removes a migration record (for rollback)
	RemoveMigration(ctx context.Context, version string) error

	// Lock acquires a migration lock to prevent concurrent migrations
	Lock(ctx context.Context) (Lock, error)

	// GetLastBatch returns the last batch number used
	GetLastBatch(ctx context.Context) (int, error)
}

// Migrator coordinates the execution of database migrations.
type Migrator interface {
	// Migrate runs all pending migrations
	Migrate(ctx context.Context) error

	// MigrateUp runs the next N pending migrations
	MigrateUp(ctx context.Context, steps int) error

	// MigrateDown rolls back the last N migrations
	MigrateDown(ctx context.Context, steps int) error

	// Rollback rolls back the last batch of migrations
	Rollback(ctx context.Context) error

	// Status returns the status of all migrations
	Status(ctx context.Context) ([]MigrationStatus, error)

	// Reset rolls back all migrations
	Reset(ctx context.Context) error

	// Refresh rolls back all migrations and re-runs them
	Refresh(ctx context.Context) error

	// SetBeforeMigrate sets the callback to run before each migration
	SetBeforeMigrate(callback MigrationCallback)

	// SetAfterMigrate sets the callback to run after each migration
	SetAfterMigrate(callback MigrationCallback)

	// SetOnError sets the callback to run when a migration fails
	SetOnError(callback MigrationErrorCallback)

	// SetLogger sets a custom logger
	SetLogger(logger Logger)
}

// Transaction represents a database transaction.
type Transaction interface {
	// Commit commits the transaction
	Commit() error

	// Rollback rolls back the transaction
	Rollback() error

	// Exec executes a SQL statement within the transaction
	Exec(ctx context.Context, query string, args ...interface{}) error

	// Query executes a SQL query within the transaction
	Query(ctx context.Context, query string, args ...interface{}) (Rows, error)
}

// Lock represents a migration lock to prevent concurrent execution.
type Lock interface {
	// Release releases the lock
	Release() error

	// IsLocked checks if the lock is currently held
	IsLocked() bool
}

// Rows represents the result of a database query.
type Rows interface {
	// Next prepares the next row for reading
	Next() bool

	// Scan copies the columns from the current row into the values
	Scan(dest ...interface{}) error

	// Close closes the rows iterator
	Close() error

	// Err returns any error that occurred during iteration
	Err() error
}

// Seeder represents a data seeder that can populate the database.
type Seeder interface {
	// Name returns the seeder name (must be unique)
	Name() string

	// Dependencies returns the names of seeders that must run before this one
	Dependencies() []string

	// Environments returns the environments where this seeder should run
	// (e.g., ["development", "test"])
	Environments() []string

	// Seed executes the seeding operation
	Seed(ctx context.Context, adapter DatabaseAdapter) error

	// ShouldRun checks if the seeder should run (for idempotency)
	ShouldRun(ctx context.Context, adapter DatabaseAdapter) (bool, error)
}

// SeedManager coordinates the execution of seeders.
type SeedManager interface {
	// Seed runs the specified seeders
	Seed(ctx context.Context, seeders ...string) error

	// SeedAll runs all seeders
	SeedAll(ctx context.Context) error

	// Status returns the status of all seeders
	Status(ctx context.Context) ([]SeedStatus, error)
}

// Logger defines the logging interface for Síl.
type Logger interface {
	// Info logs an informational message
	Info(msg string, args ...interface{})

	// Warn logs a warning message
	Warn(msg string, args ...interface{})

	// Error logs an error message
	Error(msg string, args ...interface{})

	// Debug logs a debug message
	Debug(msg string, args ...interface{})
}

// RowsAdapter wraps sql.Rows to implement the Rows interface.
type RowsAdapter struct {
	*sql.Rows
}

// Next prepares the next row for reading.
func (r *RowsAdapter) Next() bool {
	return r.Rows.Next()
}

// Scan copies the columns from the current row into the values.
func (r *RowsAdapter) Scan(dest ...interface{}) error {
	return r.Rows.Scan(dest...)
}

// Close closes the rows iterator.
func (r *RowsAdapter) Close() error {
	return r.Rows.Close()
}

// Err returns any error that occurred during iteration.
func (r *RowsAdapter) Err() error {
	return r.Rows.Err()
}

// NewRowsAdapter creates a new RowsAdapter from sql.Rows.
func NewRowsAdapter(rows *sql.Rows) Rows {
	return &RowsAdapter{Rows: rows}
}

// MigrationFunc is a function that performs a migration operation.
type MigrationFunc func(adapter DatabaseAdapter) error

// MigrationCallback is a callback function called before/after migrations.
type MigrationCallback func(migration Migration, direction string) error

// MigrationErrorCallback is a callback function called when a migration fails.
type MigrationErrorCallback func(migration Migration, err error)
