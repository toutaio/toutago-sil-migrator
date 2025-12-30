package sil

import (
	"errors"
	"fmt"
	"time"
)

var (
	// ErrMigrationNotFound is returned when a migration cannot be found
	ErrMigrationNotFound = errors.New("migration not found")

	// ErrMigrationFailed is returned when a migration fails to execute
	ErrMigrationFailed = errors.New("migration failed")

	// ErrLockAcquisitionFailed is returned when the migration lock cannot be acquired
	ErrLockAcquisitionFailed = errors.New("failed to acquire migration lock")

	// ErrLockTimeout is returned when waiting for the lock times out
	ErrLockTimeout = errors.New("timeout waiting for migration lock")

	// ErrInvalidMigrationVersion is returned when a migration version is invalid
	ErrInvalidMigrationVersion = errors.New("invalid migration version")

	// ErrDuplicateMigrationVersion is returned when a duplicate migration version is detected
	ErrDuplicateMigrationVersion = errors.New("duplicate migration version")

	// ErrNoMigrationsFound is returned when no migrations are found
	ErrNoMigrationsFound = errors.New("no migrations found")

	// ErrNoPendingMigrations is returned when no pending migrations exist
	ErrNoPendingMigrations = errors.New("no pending migrations")

	// ErrNoAppliedMigrations is returned when no applied migrations exist
	ErrNoAppliedMigrations = errors.New("no applied migrations")

	// ErrInvalidMigrationDirection is returned when an invalid migration direction is specified
	ErrInvalidMigrationDirection = errors.New("invalid migration direction")

	// ErrTransactionFailed is returned when a database transaction fails
	ErrTransactionFailed = errors.New("transaction failed")

	// ErrDatabaseConnectionFailed is returned when database connection fails
	ErrDatabaseConnectionFailed = errors.New("database connection failed")

	// ErrSeederNotFound is returned when a seeder cannot be found
	ErrSeederNotFound = errors.New("seeder not found")

	// ErrCircularDependency is returned when a circular dependency is detected in seeders
	ErrCircularDependency = errors.New("circular dependency detected")

	// ErrSeederFailed is returned when a seeder fails to execute
	ErrSeederFailed = errors.New("seeder failed")
)

// MigrationError wraps a migration error with additional context.
type MigrationError struct {
	Migration   string
	Operation   string // "up" or "down"
	Err         error
	StackTrace  string
	Recoverable bool
}

// Error returns the error message.
func (e *MigrationError) Error() string {
	return fmt.Sprintf("migration %s failed during %s: %v", e.Migration, e.Operation, e.Err)
}

// Unwrap returns the underlying error.
func (e *MigrationError) Unwrap() error {
	return e.Err
}

// ConfigurationError represents a configuration validation error.
type ConfigurationError struct {
	Field   string
	Message string
}

// Error returns the error message.
func (e *ConfigurationError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("configuration error: %s: %s", e.Field, e.Message)
	}
	return fmt.Sprintf("configuration error: %s", e.Message)
}

// ErrInvalidConfiguration creates a new configuration error.
func ErrInvalidConfiguration(message string) error {
	return &ConfigurationError{Message: message}
}

// ErrInvalidConfigurationField creates a new configuration error for a specific field.
func ErrInvalidConfigurationField(field, message string) error {
	return &ConfigurationError{Field: field, Message: message}
}

// LockError represents a lock-related error.
type LockError struct {
	Reason  string
	LockID  string
	Timeout time.Duration
}

// Error returns the error message.
func (e *LockError) Error() string {
	if e.Timeout > 0 {
		return fmt.Sprintf("lock error: %s (timeout: %v)", e.Reason, e.Timeout)
	}
	return fmt.Sprintf("lock error: %s", e.Reason)
}

// SeederError wraps a seeder error with additional context.
type SeederError struct {
	Seeder      string
	Err         error
	Recoverable bool
}

// Error returns the error message.
func (e *SeederError) Error() string {
	return fmt.Sprintf("seeder %s failed: %v", e.Seeder, e.Err)
}

// Unwrap returns the underlying error.
func (e *SeederError) Unwrap() error {
	return e.Err
}

// ValidationError represents a validation error.
type ValidationError struct {
	Field   string
	Message string
}

// Error returns the error message.
func (e *ValidationError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("validation error: %s: %s", e.Field, e.Message)
	}
	return fmt.Sprintf("validation error: %s", e.Message)
}

// WrapMigrationError wraps an error with migration context.
func WrapMigrationError(migration string, operation string, err error) error {
	return &MigrationError{
		Migration: migration,
		Operation: operation,
		Err:       err,
	}
}

// WrapSeederError wraps an error with seeder context.
func WrapSeederError(seeder string, err error) error {
	return &SeederError{
		Seeder: seeder,
		Err:    err,
	}
}

// IsMigrationError checks if an error is a MigrationError.
func IsMigrationError(err error) bool {
	var migrationErr *MigrationError
	return errors.As(err, &migrationErr)
}

// IsSeederError checks if an error is a SeederError.
func IsSeederError(err error) bool {
	var seederErr *SeederError
	return errors.As(err, &seederErr)
}

// IsConfigurationError checks if an error is a ConfigurationError.
func IsConfigurationError(err error) bool {
	var configErr *ConfigurationError
	return errors.As(err, &configErr)
}

// IsLockError checks if an error is a LockError.
func IsLockError(err error) bool {
	var lockErr *LockError
	return errors.As(err, &lockErr)
}
