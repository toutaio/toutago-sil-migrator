// Package sil provides a standalone database migration and seeding tool for Go applications.
//
// Síl (Old Irish: "seed" or "lineage") enables robust schema evolution management and
// reliable data seeding across all environments with production-ready distributed locking.
//
// # Features
//
//   - Multi-Database Support - PostgreSQL, MySQL, SQLite with automatic detection
//   - Distributed Locking - Database-appropriate locks (advisory, named, file-based)
//   - Transaction Safety - Automatic rollback on migration failures
//   - Batch Tracking - Precise rollback control by migration batches
//   - Type-Safe Migrations - Write migrations in Go
//   - Standalone - Zero required dependencies (optional datamapper integration)
//   - Beautiful CLI - Colored output with progress tracking
//   - Comprehensive Logging - Track every migration step
//
// # Quick Start
//
// Create a new migration:
//
//	sil create create_users_table
//
// Edit the generated migration file:
//
//	func (m *Migration_20241230_CreateUsersTable) Up(adapter sil.DatabaseAdapter) error {
//	    return adapter.Exec(context.Background(), `
//	        CREATE TABLE users (
//	            id SERIAL PRIMARY KEY,
//	            name VARCHAR(255) NOT NULL,
//	            email VARCHAR(255) UNIQUE NOT NULL
//	        )
//	    `)
//	}
//
//	func (m *Migration_20241230_CreateUsersTable) Down(adapter sil.DatabaseAdapter) error {
//	    return adapter.Exec(context.Background(), "DROP TABLE users")
//	}
//
// Run migrations:
//
//	export DATABASE_URL="postgres://user:pass@localhost/mydb"
//	sil migrate
//
// # Commands
//
//   - sil init - Initialize project structure
//   - sil create <name> - Create new migration
//   - sil migrate - Run pending migrations
//   - sil rollback - Rollback last batch
//   - sil status - Show migration status
//   - sil reset - Rollback all migrations
//
// # Database Adapters
//
// Built-in adapters for:
//   - PostgreSQL (uses advisory locks)
//   - MySQL (uses named locks)
//   - SQLite (uses file locks)
//
// Custom adapters can be implemented using the DatabaseAdapter interface.
//
// # Migration Files
//
// Migrations are Go source files with timestamp-based naming:
//
//	20241230100000_create_users_table.go
//
// Each migration implements Up() and Down() methods for schema changes.
//
// # Thread Safety
//
// Distributed locking ensures only one migration process runs at a time,
// safe for deployment across multiple servers.
//
// # Version
//
// This is version 0.3.0 - Phase 3 complete with 73.9% test coverage.
// Requires Go 1.22 or higher.
package sil

const (
	// Version is the current version of Síl
	Version = "0.3.0"
)
