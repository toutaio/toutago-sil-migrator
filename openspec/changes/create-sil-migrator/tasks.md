# Tasks: Síl Migration and Seeding System

## Phase 1: Foundation (Weeks 1-3)

### 1.1 Project Setup
- [ ] 1.1.1 Create repository at `/home/nestor/Proyects/toutago-sil-migrator`
- [ ] 1.1.2 Initialize Go module with `go mod init github.com/toutaio/toutago-sil-migrator`
- [ ] 1.1.3 Create directory structure (cmd/, pkg/, examples/, tests/)
- [ ] 1.1.4 Setup GitHub repository at `https://github.com/toutaio/toutago-sil-migrator`
- [ ] 1.1.5 Create README.md with project overview and quick start
- [ ] 1.1.6 Create LICENSE file (MIT or Apache 2.0)
- [ ] 1.1.7 Create .gitignore for Go projects
- [ ] 1.1.8 Setup CI/CD workflow (GitHub Actions)

### 1.2 Core Interfaces
- [ ] 1.2.1 Define `Migration` interface in `pkg/sil/interfaces.go`
- [ ] 1.2.2 Define `DatabaseAdapter` interface
- [ ] 1.2.3 Define `Migrator` interface
- [ ] 1.2.4 Define `Transaction` interface
- [ ] 1.2.5 Define `Lock` interface
- [ ] 1.2.6 Define `Config` struct
- [ ] 1.2.7 Define `MigrationRecord` and `MigrationStatus` types
- [ ] 1.2.8 Add comprehensive interface documentation

### 1.3 Migration Types
- [ ] 1.3.1 Create `BaseMigration` struct implementing `Migration` interface
- [ ] 1.3.2 Implement version parsing and validation
- [ ] 1.3.3 Create migration file loader
- [ ] 1.3.4 Implement migration sorting by version
- [ ] 1.3.5 Add migration validation (duplicate versions, format)
- [ ] 1.3.6 Create migration template generator

### 1.4 Core Migration Engine
- [ ] 1.4.1 Implement `Migrator` struct in `pkg/sil/migrator.go`
- [ ] 1.4.2 Implement `Migrate()` method (run all pending)
- [ ] 1.4.3 Implement `MigrateUp(steps int)` method
- [ ] 1.4.4 Implement `MigrateDown(steps int)` method
- [ ] 1.4.5 Implement `Rollback()` method (rollback last batch)
- [ ] 1.4.6 Implement `Status()` method (show migration status)
- [ ] 1.4.7 Implement `Reset()` method (rollback all)
- [ ] 1.4.8 Add migration execution logging
- [ ] 1.4.9 Implement error handling and recovery

### 1.5 PostgreSQL Adapter
- [ ] 1.5.1 Create `PostgresAdapter` struct in `pkg/sil/adapters/postgres.go`
- [ ] 1.5.2 Implement `Connect()` method
- [ ] 1.5.3 Implement `Close()` method
- [ ] 1.5.4 Implement `Exec()` and `Query()` methods
- [ ] 1.5.5 Implement `BeginTx()` for transaction support
- [ ] 1.5.6 Implement `CreateMigrationsTable()` method
- [ ] 1.5.7 Implement `GetAppliedMigrations()` method
- [ ] 1.5.8 Implement `RecordMigration()` method
- [ ] 1.5.9 Implement `RemoveMigration()` method
- [ ] 1.5.10 Add PostgreSQL-specific connection pooling
- [ ] 1.5.11 Add connection retry logic

### 1.6 Migration Locking
- [ ] 1.6.1 Create `Lock` implementation in `pkg/sil/lock.go`
- [ ] 1.6.2 Implement PostgreSQL advisory lock (`pg_advisory_lock()`)
- [ ] 1.6.3 Add lock timeout configuration
- [ ] 1.6.4 Implement lock acquisition with retry
- [ ] 1.6.5 Implement automatic lock release
- [ ] 1.6.6 Add manual unlock functionality
- [ ] 1.6.7 Implement lock status checking
- [ ] 1.6.8 Add lock monitoring and logging

### 1.7 Transaction Support
- [ ] 1.7.1 Create `Transaction` interface wrapper
- [ ] 1.7.2 Implement transaction begin/commit/rollback
- [ ] 1.7.3 Add automatic rollback on panic
- [ ] 1.7.4 Implement nested transaction detection
- [ ] 1.7.5 Add transaction timeout support
- [ ] 1.7.6 Implement savepoint support for partial rollback

### 1.8 Configuration
- [ ] 1.8.1 Create `Config` struct in `pkg/sil/config.go`
- [ ] 1.8.2 Add database connection configuration
- [ ] 1.8.3 Add migration directory configuration
- [ ] 1.8.4 Add lock timeout configuration
- [ ] 1.8.5 Implement configuration file loading (YAML/JSON)
- [ ] 1.8.6 Add environment variable support
- [ ] 1.8.7 Implement configuration validation
- [ ] 1.8.8 Add configuration defaults

### 1.9 CLI Foundation
- [ ] 1.9.1 Create `cmd/sil/main.go` with Cobra setup
- [ ] 1.9.2 Implement `sil init` command (create migrations directory)
- [ ] 1.9.3 Implement `sil create <name>` command (generate migration)
- [ ] 1.9.4 Implement `sil migrate` command (run pending migrations)
- [ ] 1.9.5 Implement `sil rollback` command (rollback last batch)
- [ ] 1.9.6 Implement `sil status` command (show migration status)
- [ ] 1.9.7 Implement `sil reset` command (rollback all)
- [ ] 1.9.8 Add global flags (--config, --verbose, --dry-run)
- [ ] 1.9.9 Implement colored output for better UX
- [ ] 1.9.10 Add progress indicators for long operations

### 1.10 Testing - Phase 1
- [ ] 1.10.1 Create mock `DatabaseAdapter` for testing
- [ ] 1.10.2 Write unit tests for migration loading
- [ ] 1.10.3 Write unit tests for migration sorting
- [ ] 1.10.4 Write unit tests for migration validation
- [ ] 1.10.5 Write unit tests for migrator engine
- [ ] 1.10.6 Write integration tests with real PostgreSQL (testcontainers)
- [ ] 1.10.7 Write tests for transaction rollback scenarios
- [ ] 1.10.8 Write tests for lock acquisition/release
- [ ] 1.10.9 Write tests for concurrent migration attempts
- [ ] 1.10.10 Achieve 80%+ code coverage

### 1.11 Documentation - Phase 1
- [ ] 1.11.1 Write comprehensive README with quick start
- [ ] 1.11.2 Document PostgreSQL adapter usage
- [ ] 1.11.3 Create migration writing guide
- [ ] 1.11.4 Document CLI commands and flags
- [ ] 1.11.5 Create example migrations
- [ ] 1.11.6 Write troubleshooting guide
- [ ] 1.11.7 Add GoDoc comments to all public APIs
- [ ] 1.11.8 Create CONTRIBUTING.md

## Phase 2: Multi-Database Support (Weeks 4-5)

### 2.1 MySQL Adapter
- [ ] 2.1.1 Create `MySQLAdapter` struct in `pkg/sil/adapters/mysql.go`
- [ ] 2.1.2 Implement all `DatabaseAdapter` methods for MySQL
- [ ] 2.1.3 Implement MySQL locking with `GET_LOCK()`
- [ ] 2.1.4 Handle MySQL-specific transaction limitations (DDL)
- [ ] 2.1.5 Add MySQL connection configuration
- [ ] 2.1.6 Write MySQL-specific integration tests
- [ ] 2.1.7 Document MySQL-specific behavior and limitations

### 2.2 SQLite Adapter
- [ ] 2.2.1 Create `SQLiteAdapter` struct in `pkg/sil/adapters/sqlite.go`
- [ ] 2.2.2 Implement all `DatabaseAdapter` methods for SQLite
- [ ] 2.2.3 Implement file-based locking for SQLite
- [ ] 2.2.4 Handle SQLite transaction limitations
- [ ] 2.2.5 Add SQLite connection configuration
- [ ] 2.2.6 Write SQLite-specific integration tests
- [ ] 2.2.7 Document SQLite-specific behavior and limitations

### 2.3 Adapter Testing Framework
- [ ] 2.3.1 Create adapter compliance test suite
- [ ] 2.3.2 Define standard test scenarios for all adapters
- [ ] 2.3.3 Implement automated adapter testing
- [ ] 2.3.4 Add performance benchmarks for adapters
- [ ] 2.3.5 Create adapter comparison documentation

### 2.4 Adapter Factory
- [ ] 2.4.1 Create adapter factory pattern in `pkg/sil/adapters/factory.go`
- [ ] 2.4.2 Implement adapter registration system
- [ ] 2.4.3 Add adapter auto-detection from connection string
- [ ] 2.4.4 Support custom adapter registration
- [ ] 2.4.5 Add adapter validation

### 2.5 Testing - Phase 2
- [ ] 2.5.1 Write cross-database migration tests
- [ ] 2.5.2 Test adapter switching scenarios
- [ ] 2.5.3 Test multi-database applications
- [ ] 2.5.4 Write compatibility tests for all adapters
- [ ] 2.5.5 Achieve 85%+ code coverage

### 2.6 Documentation - Phase 2
- [ ] 2.6.1 Document all supported databases
- [ ] 2.6.2 Create database-specific migration guides
- [ ] 2.6.3 Document adapter selection guide
- [ ] 2.6.4 Create multi-database examples
- [ ] 2.6.5 Document custom adapter creation

## Phase 3: Seeding System (Weeks 6-7)

### 3.1 Seeder Interfaces
- [ ] 3.1.1 Define `Seeder` interface in `pkg/sil/interfaces.go`
- [ ] 3.1.2 Define `SeedManager` interface
- [ ] 3.1.3 Create `BaseSeeder` struct
- [ ] 3.1.4 Define `SeedRecord` and `SeedStatus` types

### 3.2 Seeder Engine
- [ ] 3.2.1 Implement `SeedManager` struct in `pkg/sil/seeder.go`
- [ ] 3.2.2 Implement seeder file loader
- [ ] 3.2.3 Implement dependency graph builder
- [ ] 3.2.4 Implement topological sort for execution order
- [ ] 3.2.5 Implement `Seed(seeders ...string)` method
- [ ] 3.2.6 Implement `SeedAll()` method
- [ ] 3.2.7 Implement `Status()` method
- [ ] 3.2.8 Add seeder execution tracking
- [ ] 3.2.9 Implement idempotency checking

### 3.3 Seeder Templates
- [ ] 3.3.1 Create seeder template generator
- [ ] 3.3.2 Add common seeder patterns (users, roles, etc.)
- [ ] 3.3.3 Create factory-based seeder helpers
- [ ] 3.3.4 Add faker integration for test data

### 3.4 Environment-Specific Seeds
- [ ] 3.4.1 Add environment detection (dev, test, staging, prod)
- [ ] 3.4.2 Implement environment-specific seed filtering
- [ ] 3.4.3 Add seed tagging system (e.g., @dev, @test)
- [ ] 3.4.4 Create environment-based seed examples

### 3.5 Seed CLI Commands
- [ ] 3.5.1 Implement `sil seed:create <name>` command
- [ ] 3.5.2 Implement `sil seed:run [seeders...]` command
- [ ] 3.5.3 Implement `sil seed:run --all` command
- [ ] 3.5.4 Implement `sil seed:status` command
- [ ] 3.5.5 Add `--env` flag for environment-specific seeding
- [ ] 3.5.6 Add `--force` flag to re-run seeders

### 3.6 Testing - Phase 3
- [ ] 3.6.1 Write unit tests for dependency resolution
- [ ] 3.6.2 Write tests for circular dependency detection
- [ ] 3.6.3 Write tests for idempotency checking
- [ ] 3.6.4 Write integration tests for seeders
- [ ] 3.6.5 Test environment-specific seeding
- [ ] 3.6.6 Achieve 90%+ code coverage

### 3.7 Documentation - Phase 3
- [ ] 3.7.1 Write seeder creation guide
- [ ] 3.7.2 Document dependency management
- [ ] 3.7.3 Document idempotency patterns
- [ ] 3.7.4 Create seeder best practices guide
- [ ] 3.7.5 Add seeder examples for common scenarios
- [ ] 3.7.6 Document environment-specific seeding

## Phase 4: Advanced Features (Weeks 8-9)

### 4.1 Dry-Run Mode
- [ ] 4.1.1 Implement dry-run flag for migrations
- [ ] 4.1.2 Add SQL preview generation
- [ ] 4.1.3 Implement migration plan output
- [ ] 4.1.4 Add impact estimation (affected rows, duration)
- [ ] 4.1.5 Create dry-run examples and documentation

### 4.2 Migration Squashing
- [ ] 4.2.1 Implement `sil squash` command
- [ ] 4.2.2 Add migration range selection
- [ ] 4.2.3 Implement squashed migration generation
- [ ] 4.2.4 Add safety checks for squashing
- [ ] 4.2.5 Document squashing best practices

### 4.3 Advanced Locking Features
- [ ] 4.3.1 Implement `sil lock:status` command
- [ ] 4.3.2 Implement `sil lock:release` command (emergency unlock)
- [ ] 4.3.3 Add lock monitoring and alerting
- [ ] 4.3.4 Implement lock acquisition history

### 4.4 Migration Helpers
- [ ] 4.4.1 Create schema builder helpers (create_table, add_column, etc.)
- [ ] 4.4.2 Add common migration patterns (add_timestamps, add_soft_deletes)
- [ ] 4.4.3 Implement migration batching helpers for large data
- [ ] 4.4.4 Add progress reporting for long-running migrations

### 4.5 Performance Optimization
- [ ] 4.5.1 Optimize migration file loading
- [ ] 4.5.2 Add connection pooling optimization
- [ ] 4.5.3 Implement parallel seeding (where safe)
- [ ] 4.5.4 Add query optimization for migration tracking
- [ ] 4.5.5 Run performance benchmarks

### 4.6 Error Handling & Recovery
- [ ] 4.6.1 Implement detailed error reporting
- [ ] 4.6.2 Add migration failure recovery suggestions
- [ ] 4.6.3 Implement partial migration rollback
- [ ] 4.6.4 Add migration checksum validation
- [ ] 4.6.5 Create error handling guide

### 4.7 Programmatic API
- [ ] 4.7.1 Create `NewMigrator()` factory function
- [ ] 4.7.2 Add programmatic migration execution API
- [ ] 4.7.3 Implement event callbacks (before/after migration)
- [ ] 4.7.4 Add custom logger interface
- [ ] 4.7.5 Create programmatic API examples
- [ ] 4.7.6 Document library usage patterns

### 4.8 Testing - Phase 4
- [ ] 4.8.1 Write E2E tests for all features
- [ ] 4.8.2 Write tests for dry-run mode
- [ ] 4.8.3 Write tests for migration squashing
- [ ] 4.8.4 Write stress tests for large migrations
- [ ] 4.8.5 Write tests for programmatic API
- [ ] 4.8.6 Achieve 90%+ code coverage

### 4.9 Documentation - Phase 4
- [ ] 4.9.1 Create comprehensive API documentation
- [ ] 4.9.2 Write advanced usage guide
- [ ] 4.9.3 Document all CLI commands and flags
- [ ] 4.9.4 Create migration cookbook (common patterns)
- [ ] 4.9.5 Write performance tuning guide
- [ ] 4.9.6 Create troubleshooting guide
- [ ] 4.9.7 Add video tutorials (optional)

### 4.10 Examples & Integration
- [ ] 4.10.1 Create basic usage example in `examples/basic/`
- [ ] 4.10.2 Create multi-database example in `examples/multi-db/`
- [ ] 4.10.3 Create Toutā integration example in `examples/with-touta/`
- [ ] 4.10.4 Create Docker example with migrations
- [ ] 4.10.5 Create CI/CD integration example

### 4.11 Release Preparation
- [ ] 4.11.1 Review all documentation for completeness
- [ ] 4.11.2 Create CHANGELOG.md
- [ ] 4.11.3 Tag v0.1.0-alpha release
- [ ] 4.11.4 Publish to GitHub
- [ ] 4.11.5 Create announcement blog post
- [ ] 4.11.6 Submit to awesome-go lists

## Toutā Integration (Post-Release)

### 5.1 Optional Toutā CLI Integration
- [ ] 5.1.1 Create `touta migrate` command wrapper
- [ ] 5.1.2 Create `touta seed` command wrapper
- [ ] 5.1.3 Add Síl to Toutā dependencies
- [ ] 5.1.4 Document Toutā integration in both projects
- [ ] 5.1.5 Create Toutā project template with migrations

### 5.2 Storage Adapter Integration
- [ ] 5.2.1 Update storage adapter documentation to reference Síl
- [ ] 5.2.2 Create example storage adapter with Síl migrations
- [ ] 5.2.3 Add migration support to nemeton templates

### 5.3 Ritual Templates
- [ ] 5.3.1 Add migrations to blog ritual template
- [ ] 5.3.2 Add seeders to blog ritual template
- [ ] 5.3.3 Create ritual deployment guide with migrations
