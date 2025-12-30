// Package sil provides a standalone database migration and seeding tool for Go projects.
//
// Síl (Old Irish for "seed" or "lineage") enables version-based database migrations
// with support for multiple databases, transaction safety, and distributed locking.
//
// Core Features:
//   - Version-based sequential migrations
//   - Transaction-wrapped execution with auto-rollback
//   - Distributed migration locking
//   - Multiple database adapter support
//   - Data seeding with dependency management
//   - CLI and programmatic API
//
// Example Usage:
//
//	config := &sil.Config{
//	    DatabaseURL:   "postgres://user:pass@localhost/db",
//	    MigrationsDir: "./migrations",
//	}
//
//	adapter, _ := adapters.NewPostgresAdapter(config)
//	migrator, _ := sil.NewMigrator(config, adapter)
//
//	// Run all pending migrations
//	migrator.Migrate()
package sil

const (
	// Version is the current version of Síl
	Version = "0.1.0-dev"
)
