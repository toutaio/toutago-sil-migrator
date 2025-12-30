package sil

import (
	"fmt"
	"time"
)

// Config holds the configuration for Síl migrations and seeding.
type Config struct {
	// DatabaseURL is the connection string for the database
	// Examples:
	//   - PostgreSQL: "postgres://user:pass@localhost:5432/dbname"
	//   - MySQL: "mysql://user:pass@localhost:3306/dbname"
	//   - SQLite: "sqlite://./path/to/database.db"
	DatabaseURL string `yaml:"database_url" json:"database_url"`

	// MigrationsDir is the directory containing migration files
	MigrationsDir string `yaml:"migrations_dir" json:"migrations_dir"`

	// SeedersDir is the directory containing seeder files
	SeedersDir string `yaml:"seeders_dir" json:"seeders_dir"`

	// LockTimeout is the maximum time to wait for migration lock
	LockTimeout time.Duration `yaml:"lock_timeout" json:"lock_timeout"`

	// MigrationTimeout is the maximum time allowed for a single migration
	MigrationTimeout time.Duration `yaml:"migration_timeout" json:"migration_timeout"`

	// MaxConnections is the maximum number of database connections
	MaxConnections int `yaml:"max_connections" json:"max_connections"`

	// MaxIdleConnections is the maximum number of idle connections
	MaxIdleConnections int `yaml:"max_idle_connections" json:"max_idle_connections"`

	// ConnectionMaxLifetime is the maximum lifetime of a connection
	ConnectionMaxLifetime time.Duration `yaml:"connection_max_lifetime" json:"connection_max_lifetime"`

	// Environment is the current environment (development, test, production)
	Environment string `yaml:"environment" json:"environment"`

	// TableName is the name of the migrations table
	TableName string `yaml:"table_name" json:"table_name"`

	// SeedsTableName is the name of the seeds table
	SeedsTableName string `yaml:"seeds_table_name" json:"seeds_table_name"`

	// Verbose enables verbose logging
	Verbose bool `yaml:"verbose" json:"verbose"`
}

// MigrationRecord represents a migration that has been applied to the database.
type MigrationRecord struct {
	ID          int64
	Version     string
	Description string
	Batch       int
	ExecutedAt  time.Time
}

// MigrationStatus represents the status of a migration.
type MigrationStatus struct {
	Version     string
	Description string
	Applied     bool
	Batch       int
	ExecutedAt  *time.Time
}

// SeedRecord represents a seeder that has been executed.
type SeedRecord struct {
	ID          int64
	Name        string
	Environment string
	ExecutedAt  time.Time
}

// SeedStatus represents the status of a seeder.
type SeedStatus struct {
	Name        string
	Executed    bool
	Skipped     bool
	Reason      string
	ExecutedAt  *time.Time
	Environment string
}

// DefaultConfig returns a configuration with sensible defaults.
func DefaultConfig() *Config {
	return &Config{
		MigrationsDir:         "./migrations",
		SeedersDir:            "./seeders",
		LockTimeout:           5 * time.Minute,
		MigrationTimeout:      30 * time.Minute,
		MaxConnections:        10,
		MaxIdleConnections:    5,
		ConnectionMaxLifetime: time.Hour,
		Environment:           "development",
		TableName:             "sil_migrations",
		SeedsTableName:        "sil_seeds",
		Verbose:               false,
	}
}

// Validate validates the configuration and returns an error if invalid.
func (c *Config) Validate() error {
	if c.DatabaseURL == "" {
		return ErrInvalidConfiguration("database_url is required")
	}

	if c.MigrationsDir == "" {
		return ErrInvalidConfiguration("migrations_dir is required")
	}

	if c.LockTimeout <= 0 {
		return ErrInvalidConfiguration("lock_timeout must be positive")
	}

	if c.MigrationTimeout <= 0 {
		return ErrInvalidConfiguration("migration_timeout must be positive")
	}

	if c.MaxConnections <= 0 {
		return ErrInvalidConfiguration("max_connections must be positive")
	}

	if c.TableName == "" {
		return ErrInvalidConfiguration("table_name is required")
	}

	return nil
}

// Merge merges another configuration into this one, with the other config taking precedence.
func (c *Config) Merge(other *Config) *Config {
	result := *c

	if other.DatabaseURL != "" {
		result.DatabaseURL = other.DatabaseURL
	}
	if other.MigrationsDir != "" {
		result.MigrationsDir = other.MigrationsDir
	}
	if other.SeedersDir != "" {
		result.SeedersDir = other.SeedersDir
	}
	if other.LockTimeout > 0 {
		result.LockTimeout = other.LockTimeout
	}
	if other.MigrationTimeout > 0 {
		result.MigrationTimeout = other.MigrationTimeout
	}
	if other.MaxConnections > 0 {
		result.MaxConnections = other.MaxConnections
	}
	if other.MaxIdleConnections > 0 {
		result.MaxIdleConnections = other.MaxIdleConnections
	}
	if other.ConnectionMaxLifetime > 0 {
		result.ConnectionMaxLifetime = other.ConnectionMaxLifetime
	}
	if other.Environment != "" {
		result.Environment = other.Environment
	}
	if other.TableName != "" {
		result.TableName = other.TableName
	}
	if other.SeedsTableName != "" {
		result.SeedsTableName = other.SeedsTableName
	}
	if other.Verbose {
		result.Verbose = other.Verbose
	}

	return &result
}

// String returns a string representation of the config (with sensitive data masked).
func (c *Config) String() string {
	return fmt.Sprintf("Config{MigrationsDir: %s, Environment: %s, TableName: %s}",
		c.MigrationsDir, c.Environment, c.TableName)
}
