# Síl Migration and Seeding System - Implementation Plan

**Project**: Toutago-Sil-Migrator  
**Repository**: `github.com/toutaio/toutago-sil-migrator`  
**Go Version**: 1.22+  
**Timeline**: 8-10 weeks across 4 phases  
**Test Coverage Target**: 90%+

---

## Executive Summary

Síl is a standalone database migration and seeding tool for Go projects, inspired by Rails, Laravel, Flyway, and Alembic. This implementation plan details the phased delivery of:

- **Phase 1 (Weeks 1-3)**: Core migration engine with PostgreSQL support
- **Phase 2 (Weeks 4-5)**: Multi-database support (MySQL, SQLite) and optional datamapper integration
- **Phase 3 (Weeks 6-7)**: Seeding system with dependency management
- **Phase 4 (Weeks 8-9)**: Advanced features and production optimization

**Key Principles**:
- ✅ Zero required dependencies (fully standalone)
- ✅ Optional toutago-datamapper integration via bridge adapter
- ✅ Interface-driven architecture (SOLID principles)
- ✅ Production-ready with distributed locking
- ✅ Developer-friendly CLI and programmatic API

---

## Phase 1: Foundation (Weeks 1-3)

**Goal**: Establish core migration engine with PostgreSQL adapter and basic CLI.

### Week 1: Project Setup & Core Interfaces

#### 1.1 Repository Initialization
**Duration**: 1 day  
**Priority**: Critical

**Tasks**:
- [ ] Create local repository at `/home/nestor/Proyects/toutago-sil-migrator`
- [ ] Initialize Go module: `go mod init github.com/toutaio/toutago-sil-migrator`
- [ ] Create directory structure:
  ```
  toutago-sil-migrator/
  ├── cmd/sil/              # CLI binary
  ├── pkg/sil/              # Core library
  │   ├── adapters/         # Database adapters
  │   └── templates/        # Migration templates
  ├── examples/             # Usage examples
  │   ├── basic/
  │   ├── multi-db/
  │   └── with-datamapper/
  ├── tests/
  │   ├── unit/
  │   ├── integration/
  │   └── e2e/
  ├── migrations/           # Example migrations
  ├── README.md
  ├── LICENSE
  └── .gitignore
  ```
- [ ] Create README.md with project vision and quick start placeholder
- [ ] Add LICENSE file (MIT)
- [ ] Create .gitignore for Go projects
- [ ] Setup GitHub repository at `https://github.com/toutaio/toutago-sil-migrator`
- [ ] Configure GitHub Actions CI/CD workflow (.github/workflows/ci.yml)
  - Go test with coverage
  - Linting (golangci-lint)
  - Build verification

**Deliverables**:
- ✅ Repository structure in place
- ✅ CI/CD pipeline active
- ✅ Initial commit pushed to GitHub

---

#### 1.2 Core Interfaces Definition
**Duration**: 2 days  
**Priority**: Critical

**Tasks**:
- [ ] Create `pkg/sil/interfaces.go` with all core interfaces:

```go
// Migration represents a database migration
type Migration interface {
    Version() string
    Description() string
    Up(adapter DatabaseAdapter) error
    Down(adapter DatabaseAdapter) error
}

// DatabaseAdapter handles database-specific operations
type DatabaseAdapter interface {
    Connect(config Config) error
    Close() error
    Exec(query string, args ...interface{}) error
    Query(query string, args ...interface{}) (Rows, error)
    BeginTx() (Transaction, error)
    CreateMigrationsTable() error
    GetAppliedMigrations() ([]MigrationRecord, error)
    RecordMigration(version, description string) error
    RemoveMigration(version string) error
    Lock() (Lock, error)
}

// Migrator coordinates migration execution
type Migrator interface {
    Migrate() error
    MigrateUp(steps int) error
    MigrateDown(steps int) error
    Rollback() error
    Status() ([]MigrationStatus, error)
    Reset() error
}

// Transaction represents a database transaction
type Transaction interface {
    Commit() error
    Rollback() error
    Exec(query string, args ...interface{}) error
    Query(query string, args ...interface{}) (Rows, error)
}

// Lock represents a migration lock
type Lock interface {
    Release() error
    IsLocked() bool
}
```

- [ ] Define supporting types in `pkg/sil/types.go`:
  - `Config` struct (database connection, directories, timeouts)
  - `MigrationRecord` struct (version, description, batch, timestamp)
  - `MigrationStatus` struct (version, applied, timestamp)
  - `Rows` interface for query results

- [ ] Add comprehensive GoDoc comments to all interfaces
- [ ] Create `pkg/sil/errors.go` with custom error types:
  - `ErrMigrationNotFound`
  - `ErrMigrationFailed`
  - `ErrLockAcquisitionFailed`
  - `ErrInvalidConfiguration`

**Deliverables**:
- ✅ Complete interface definitions
- ✅ Type definitions with documentation
- ✅ Error types defined

---

#### 1.3 Migration Types & File Management
**Duration**: 2 days  
**Priority**: Critical

**Tasks**:
- [ ] Create `pkg/sil/migration.go`:
  - Implement `BaseMigration` struct
  - Version parsing from filename format: `YYYYMMDDHHMMSS_description.go`
  - Version validation (timestamp format, uniqueness)
  
- [ ] Create `pkg/sil/loader.go`:
  - Migration file discovery from directory
  - Go plugin loading mechanism or compiled-in migrations
  - Migration sorting by version (chronological)
  - Duplicate detection

- [ ] Create `pkg/sil/validator.go`:
  - Version format validation
  - Duplicate version detection
  - Migration integrity checks

- [ ] Create `pkg/sil/generator.go`:
  - Migration template generator
  - Timestamp generation for new migrations
  - File naming with description sanitization

**Implementation Example**:
```go
type BaseMigration struct {
    version     string
    description string
    upFunc      func(DatabaseAdapter) error
    downFunc    func(DatabaseAdapter) error
}

func (m *BaseMigration) Version() string { return m.version }
func (m *BaseMigration) Description() string { return m.description }
func (m *BaseMigration) Up(adapter DatabaseAdapter) error { return m.upFunc(adapter) }
func (m *BaseMigration) Down(adapter DatabaseAdapter) error { return m.downFunc(adapter) }
```

**Deliverables**:
- ✅ Migration type implementation
- ✅ File loader with validation
- ✅ Template generator

---

### Week 2: Migration Engine & PostgreSQL Adapter

#### 1.4 Core Migration Engine
**Duration**: 3 days  
**Priority**: Critical

**Tasks**:
- [ ] Create `pkg/sil/migrator.go` implementing `Migrator` interface:

```go
type migrator struct {
    adapter       DatabaseAdapter
    migrationsDir string
    logger        Logger
}

func NewMigrator(config Config, adapter DatabaseAdapter) (Migrator, error)
```

- [ ] Implement core methods:
  - `Migrate()` - Run all pending migrations
  - `MigrateUp(steps int)` - Run N pending migrations
  - `MigrateDown(steps int)` - Rollback N migrations
  - `Rollback()` - Rollback last batch
  - `Status()` - Show migration status
  - `Reset()` - Rollback all migrations

- [ ] Add migration execution logic:
  - Load migrations from disk
  - Query applied migrations from database
  - Calculate pending migrations (set difference)
  - Execute migrations in transaction
  - Record successful migrations
  - Handle failures with rollback

- [ ] Implement batch tracking:
  - Group migrations by execution batch
  - Enable rollback of migration batches
  - Track batch numbers

- [ ] Add comprehensive logging:
  - Migration start/complete events
  - Execution time tracking
  - Error details with context

**Deliverables**:
- ✅ Complete Migrator implementation
- ✅ Batch tracking system
- ✅ Execution logging

---

#### 1.5 PostgreSQL Adapter
**Duration**: 3 days  
**Priority**: Critical

**Tasks**:
- [ ] Create `pkg/sil/adapters/postgres.go`:

```go
type PostgresAdapter struct {
    db     *sql.DB
    config Config
    logger Logger
}

func NewPostgresAdapter(config Config) (*PostgresAdapter, error)
```

- [ ] Implement `DatabaseAdapter` interface methods:
  - `Connect()` - Establish connection with pgx driver
  - `Close()` - Close database connection
  - `Exec()` - Execute SQL statement
  - `Query()` - Execute SQL query, return rows
  - `BeginTx()` - Start transaction
  - `CreateMigrationsTable()` - Create `sil_migrations` table:
    ```sql
    CREATE TABLE IF NOT EXISTS sil_migrations (
        id SERIAL PRIMARY KEY,
        version VARCHAR(255) UNIQUE NOT NULL,
        description TEXT,
        batch INTEGER NOT NULL,
        executed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    )
    ```
  - `GetAppliedMigrations()` - Query applied migrations
  - `RecordMigration()` - Insert migration record
  - `RemoveMigration()` - Delete migration record
  - `Lock()` - Acquire advisory lock

- [ ] Add connection pooling configuration:
  - Max connections
  - Idle connections
  - Connection lifetime

- [ ] Implement connection retry logic:
  - Exponential backoff
  - Configurable retry attempts
  - Timeout handling

**Deliverables**:
- ✅ Complete PostgreSQL adapter
- ✅ Connection pooling configured
- ✅ Retry logic implemented

---

### Week 3: Locking, Transactions & CLI

#### 1.6 Migration Locking System
**Duration**: 2 days  
**Priority**: High

**Tasks**:
- [ ] Create `pkg/sil/lock.go`:

```go
type lock struct {
    adapter DatabaseAdapter
    lockID  int64
    logger  Logger
}

func (l *lock) Release() error
func (l *lock) IsLocked() bool
```

- [ ] Implement PostgreSQL advisory locking:
  - Use `pg_advisory_lock(key)` for acquisition
  - Use `pg_advisory_unlock(key)` for release
  - Generate consistent lock key from database name
  - Automatic release on connection loss

- [ ] Add lock timeout configuration:
  - Default: 5 minutes
  - Configurable via `Config.LockTimeout`
  - Fail-fast if lock cannot be acquired

- [ ] Implement lock status checking:
  - Query current lock holder
  - Display lock acquisition time
  - Show which process/instance holds lock

- [ ] Add manual unlock functionality:
  - Emergency unlock command
  - Force release with confirmation
  - Logging of manual interventions

**Deliverables**:
- ✅ Lock implementation
- ✅ Advisory lock support
- ✅ Lock monitoring

---

#### 1.7 Transaction Support
**Duration**: 2 days  
**Priority**: High

**Tasks**:
- [ ] Create `pkg/sil/transaction.go`:

```go
type transaction struct {
    tx     *sql.Tx
    logger Logger
}

func (t *transaction) Commit() error
func (t *transaction) Rollback() error
func (t *transaction) Exec(query string, args ...interface{}) error
func (t *transaction) Query(query string, args ...interface{}) (Rows, error)
```

- [ ] Implement transaction wrapper:
  - Begin transaction before migration
  - Commit on successful migration
  - Rollback on migration error
  - Rollback on panic (defer + recover)

- [ ] Add savepoint support:
  - Create savepoints for partial rollback
  - Rollback to savepoint on error
  - Release savepoints on success

- [ ] Implement transaction timeout:
  - Configurable timeout per migration
  - Default: 30 minutes
  - Cancel long-running migrations

**Deliverables**:
- ✅ Transaction wrapper
- ✅ Auto-rollback on failure
- ✅ Savepoint support

---

#### 1.8 Configuration Management
**Duration**: 1 day  
**Priority**: High

**Tasks**:
- [ ] Create `pkg/sil/config.go`:

```go
type Config struct {
    DatabaseURL     string
    MigrationsDir   string
    LockTimeout     time.Duration
    MigrationTimeout time.Duration
    MaxConnections  int
    Environment     string
}

func LoadConfig(path string) (*Config, error)
func (c *Config) Validate() error
```

- [ ] Implement configuration loading:
  - YAML file support (`sil.yaml`)
  - JSON file support (`sil.json`)
  - Environment variable overrides (`SIL_DATABASE_URL`, etc.)
  - CLI flag overrides

- [ ] Add configuration validation:
  - Required fields check
  - Format validation
  - Default value application

- [ ] Configuration file example:
```yaml
database_url: "postgres://user:pass@localhost:5432/mydb"
migrations_dir: "./migrations"
lock_timeout: 300s
migration_timeout: 1800s
max_connections: 10
environment: development
```

**Deliverables**:
- ✅ Config struct and loader
- ✅ Multi-source configuration
- ✅ Validation logic

---

#### 1.9 CLI Foundation
**Duration**: 2 days  
**Priority**: Critical

**Tasks**:
- [ ] Create `cmd/sil/main.go` with Cobra CLI framework:

```go
var rootCmd = &cobra.Command{
    Use:   "sil",
    Short: "Síl - Database Migration and Seeding Tool",
}

func main() {
    if err := rootCmd.Execute(); err != nil {
        os.Exit(1)
    }
}
```

- [ ] Implement commands:
  - `sil init` - Initialize migrations directory
  - `sil create <name>` - Generate new migration file
  - `sil migrate` - Run pending migrations
  - `sil rollback [--steps=N]` - Rollback migrations
  - `sil status` - Show migration status
  - `sil reset` - Rollback all migrations
  - `sil version` - Show Síl version

- [ ] Add global flags:
  - `--config` - Configuration file path
  - `--verbose` - Enable verbose logging
  - `--dry-run` - Preview without executing
  - `--env` - Environment selection

- [ ] Implement colored output:
  - Green for success messages
  - Red for errors
  - Yellow for warnings
  - Blue for informational messages

- [ ] Add progress indicators:
  - Spinner for long operations
  - Progress bar for batch migrations
  - Elapsed time display

**Deliverables**:
- ✅ Complete CLI with all commands
- ✅ Colored output
- ✅ Progress indicators

---

#### 1.10 Testing - Phase 1
**Duration**: 2 days  
**Priority**: Critical

**Tasks**:
- [ ] Create `pkg/sil/adapters/mock.go` - Mock adapter for testing
- [ ] Write unit tests:
  - `migration_test.go` - Migration loading, sorting, validation
  - `migrator_test.go` - Migration execution logic
  - `loader_test.go` - File discovery and loading
  - `validator_test.go` - Validation rules
  - `config_test.go` - Configuration loading

- [ ] Write integration tests with real PostgreSQL:
  - Use testcontainers-go for database
  - Test full migration lifecycle
  - Test transaction rollback scenarios
  - Test lock acquisition/release
  - Test concurrent migration attempts

- [ ] Achieve 80%+ code coverage
- [ ] Setup coverage reporting in CI/CD

**Deliverables**:
- ✅ Comprehensive test suite
- ✅ 80%+ coverage
- ✅ Integration tests with PostgreSQL

---

#### 1.11 Documentation - Phase 1
**Duration**: 1 day  
**Priority**: High

**Tasks**:
- [ ] Write comprehensive README.md:
  - Project overview
  - Quick start guide
  - Installation instructions
  - Basic usage examples
  - Configuration reference

- [ ] Create PostgreSQL adapter guide:
  - Connection string format
  - Configuration options
  - Best practices

- [ ] Write migration guide:
  - Creating migrations
  - Writing Up/Down functions
  - Migration patterns
  - Common pitfalls

- [ ] Document CLI commands:
  - Command reference
  - Flag descriptions
  - Usage examples

- [ ] Create troubleshooting guide:
  - Common errors and solutions
  - Lock issues
  - Transaction failures

- [ ] Add GoDoc comments to all public APIs

**Deliverables**:
- ✅ Complete documentation
- ✅ Usage examples
- ✅ Troubleshooting guide

---

### Phase 1 Success Criteria

✅ Core migration engine functional  
✅ PostgreSQL adapter production-ready  
✅ CLI commands working  
✅ Migration locking prevents concurrent runs  
✅ Transactions protect data integrity  
✅ 80%+ test coverage achieved  
✅ Documentation complete  

---

## Phase 2: Multi-Database Support (Weeks 4-5)

**Goal**: Add MySQL and SQLite adapters, plus optional datamapper integration.

### Week 4: MySQL & SQLite Adapters

#### 2.1 MySQL Adapter
**Duration**: 2 days  
**Priority**: High

**Tasks**:
- [ ] Create `pkg/sil/adapters/mysql.go`:

```go
type MySQLAdapter struct {
    db     *sql.DB
    config Config
    logger Logger
}

func NewMySQLAdapter(config Config) (*MySQLAdapter, error)
```

- [ ] Implement all `DatabaseAdapter` methods for MySQL
- [ ] Handle MySQL-specific considerations:
  - DDL statements trigger implicit commit (cannot rollback)
  - Use `GET_LOCK()` and `RELEASE_LOCK()` for locking
  - Different migrations table structure if needed
  - Character set and collation settings

- [ ] Add MySQL connection configuration:
  - Character set (utf8mb4)
  - Parse time settings
  - Connection parameters

- [ ] Write MySQL-specific integration tests
- [ ] Document MySQL limitations and best practices

**Deliverables**:
- ✅ MySQL adapter implementation
- ✅ MySQL-specific tests
- ✅ Documentation

---

#### 2.2 SQLite Adapter
**Duration**: 2 days  
**Priority**: High

**Tasks**:
- [ ] Create `pkg/sil/adapters/sqlite.go`:

```go
type SQLiteAdapter struct {
    db       *sql.DB
    dbPath   string
    lockFile string
    config   Config
    logger   Logger
}

func NewSQLiteAdapter(config Config) (*SQLiteAdapter, error)
```

- [ ] Implement all `DatabaseAdapter` methods for SQLite
- [ ] Handle SQLite-specific considerations:
  - File-based locking (lock file in same directory)
  - Enable WAL mode for better concurrency
  - Transaction limitations
  - Single writer constraint

- [ ] Implement file-based locking:
  - Create `.lock` file before migrations
  - Remove lock file after completion
  - Handle stale lock files

- [ ] Write SQLite-specific integration tests
- [ ] Document SQLite limitations and use cases

**Deliverables**:
- ✅ SQLite adapter implementation
- ✅ File locking system
- ✅ Documentation

---

### Week 5: Datamapper Integration & Testing

#### 2.3 Datamapper Bridge Adapter
**Duration**: 3 days  
**Priority**: Medium

**Tasks**:
- [ ] Create `pkg/sil/adapters/datamapper.go`:

```go
// +build datamapper

package adapters

import (
    "github.com/toutaio/toutago-datamapper/engine"
    "github.com/toutaio/toutago-datamapper/adapter"
)

type DatamapperAdapter struct {
    mapper     *engine.Mapper
    sourceName string
    adapter    adapter.Adapter
    logger     Logger
}

func NewDatamapperAdapter(mapper *engine.Mapper, sourceName string) (*DatamapperAdapter, error)
```

- [ ] Implement bridge methods:
  - Map `Exec()` to datamapper adapter execute
  - Map `Query()` to datamapper adapter query
  - Map `BeginTx()` to datamapper transaction
  - Create migrations table using datamapper

- [ ] Add optional dependency:
  - Use build tags for optional compilation
  - Only import datamapper if tag is set
  - Document build tag usage: `go build -tags=datamapper`

- [ ] Support datamapper configuration:
  - Load datamapper config file
  - Use datamapper source configuration
  - Reuse datamapper adapters (MySQL, PostgreSQL)

- [ ] Write integration tests:
  - Test with datamapper MySQL adapter
  - Test with datamapper PostgreSQL adapter
  - Test configuration sharing

- [ ] Create example in `examples/with-datamapper/`:
  - Show datamapper + Síl integration
  - Shared configuration example
  - Migration execution with datamapper

**Deliverables**:
- ✅ Datamapper bridge adapter
- ✅ Build tags configured
- ✅ Integration example
- ✅ Documentation

---

#### 2.4 Adapter Testing Framework
**Duration**: 1 day  
**Priority**: Medium

**Tasks**:
- [ ] Create `tests/adapter_compliance_test.go`:
  - Standard test scenarios for all adapters
  - Connection test
  - Table creation test
  - Migration CRUD test
  - Transaction test
  - Lock test

- [ ] Run compliance tests against all adapters:
  - PostgreSQL
  - MySQL
  - SQLite
  - Datamapper (if build tag enabled)

- [ ] Add performance benchmarks:
  - Migration execution speed
  - Lock acquisition time
  - Query performance

- [ ] Create adapter comparison documentation:
  - Feature matrix
  - Performance comparison
  - Use case recommendations

**Deliverables**:
- ✅ Compliance test suite
- ✅ All adapters pass tests
- ✅ Benchmark results

---

#### 2.5 Adapter Factory & Registry
**Duration**: 1 day  
**Priority**: Medium

**Tasks**:
- [ ] Create `pkg/sil/adapters/factory.go`:

```go
type AdapterFactory func(config Config) (DatabaseAdapter, error)

var adapterRegistry = map[string]AdapterFactory{}

func RegisterAdapter(name string, factory AdapterFactory)
func CreateAdapter(name string, config Config) (DatabaseAdapter, error)
func DetectAdapter(databaseURL string) (string, error)
```

- [ ] Implement auto-detection:
  - Parse connection string prefix
  - Detect `postgres://`, `mysql://`, `sqlite://`
  - Support custom adapter registration

- [ ] Register built-in adapters:
  - PostgreSQL
  - MySQL
  - SQLite
  - Datamapper (if build tag enabled)

- [ ] Add adapter validation:
  - Verify adapter implements interface
  - Check adapter is registered
  - Provide helpful error messages

**Deliverables**:
- ✅ Adapter factory system
- ✅ Auto-detection working
- ✅ Custom adapter support

---

#### 2.6 Testing - Phase 2
**Duration**: 1 day  
**Priority**: High

**Tasks**:
- [ ] Write cross-database migration tests:
  - Create migration on PostgreSQL, verify on MySQL
  - Test migration portability
  - Test database-specific SQL handling

- [ ] Test adapter switching scenarios:
  - Switch from PostgreSQL to MySQL
  - Verify migration state consistency
  - Test rollback across adapters

- [ ] Write multi-database application tests:
  - Multiple adapters in single application
  - Separate migration directories per database
  - Concurrent migrations to different databases

- [ ] Achieve 85%+ code coverage

**Deliverables**:
- ✅ Cross-database tests
- ✅ 85%+ coverage
- ✅ All adapters verified

---

#### 2.7 Documentation - Phase 2
**Duration**: 1 day  
**Priority**: High

**Tasks**:
- [ ] Document all supported databases:
  - PostgreSQL guide
  - MySQL guide
  - SQLite guide
  - Datamapper integration guide

- [ ] Create database-specific migration guides:
  - PostgreSQL best practices
  - MySQL DDL limitations
  - SQLite constraints

- [ ] Document adapter selection guide:
  - When to use each adapter
  - Performance considerations
  - Feature comparison

- [ ] Create multi-database examples:
  - Example in `examples/multi-db/`
  - Show multiple database handling
  - Configuration examples

- [ ] Document custom adapter creation:
  - Adapter interface requirements
  - Registration process
  - Testing custom adapters

- [ ] Document datamapper integration:
  - Setup instructions
  - Configuration sharing
  - Benefits and use cases

**Deliverables**:
- ✅ Complete database documentation
- ✅ Adapter selection guide
- ✅ Custom adapter guide
- ✅ Datamapper integration docs

---

### Phase 2 Success Criteria

✅ MySQL adapter production-ready  
✅ SQLite adapter production-ready  
✅ Datamapper bridge adapter functional (optional)  
✅ Adapter factory and auto-detection working  
✅ Cross-database tests passing  
✅ 85%+ test coverage achieved  
✅ Multi-database documentation complete  

---

## Phase 3: Seeding System (Weeks 6-7)

**Goal**: Implement data seeding with dependency management and idempotency.

### Week 6: Seeder Engine

#### 3.1 Seeder Interfaces
**Duration**: 1 day  
**Priority**: Critical

**Tasks**:
- [ ] Add to `pkg/sil/interfaces.go`:

```go
// Seeder represents a data seeder
type Seeder interface {
    Name() string
    Dependencies() []string
    Environments() []string
    Seed(adapter DatabaseAdapter) error
    ShouldRun(adapter DatabaseAdapter) (bool, error)
}

// SeedManager coordinates seeder execution
type SeedManager interface {
    Seed(seeders ...string) error
    SeedAll() error
    Status() ([]SeedStatus, error)
}
```

- [ ] Define supporting types:
  - `SeedRecord` struct (name, environment, executed_at)
  - `SeedStatus` struct (name, executed, skipped, reason)

- [ ] Create `pkg/sil/seeder.go`:
  - `BaseSeeder` struct implementation
  - Seeder registration system

**Deliverables**:
- ✅ Seeder interfaces defined
- ✅ Supporting types created
- ✅ Base implementation

---

#### 3.2 Seeder Engine Implementation
**Duration**: 3 days  
**Priority**: Critical

**Tasks**:
- [ ] Create `pkg/sil/seed_manager.go`:

```go
type seedManager struct {
    adapter    DatabaseAdapter
    seedersDir string
    logger     Logger
}

func NewSeedManager(config Config, adapter DatabaseAdapter) (SeedManager, error)
```

- [ ] Implement seeder file loader:
  - Discover seeder files in directory
  - Load seeder definitions
  - Validate seeder implementations

- [ ] Implement dependency graph builder:
  - Parse `Dependencies()` from each seeder
  - Build directed graph
  - Detect circular dependencies

- [ ] Implement topological sort:
  - Order seeders by dependencies
  - Ensure dependencies run first
  - Handle multiple valid orderings

- [ ] Implement `Seed()` method:
  - Load specified seeders
  - Resolve dependency order
  - Check `ShouldRun()` for idempotency
  - Execute seeders in order
  - Record execution in database

- [ ] Implement `SeedAll()` method:
  - Load all seeders from directory
  - Apply same logic as `Seed()`

- [ ] Implement `Status()` method:
  - Show all seeders
  - Display execution status
  - Show skip reasons

- [ ] Add execution tracking:
  - Create `sil_seeds` table
  - Record seeder executions
  - Track environment and timestamp

**Deliverables**:
- ✅ SeedManager implementation
- ✅ Dependency resolution working
- ✅ Execution tracking functional

---

#### 3.3 Seeder Templates
**Duration**: 1 day  
**Priority**: Medium

**Tasks**:
- [ ] Create `pkg/sil/seed_generator.go`:
  - Generate seeder file templates
  - Support different seeder types

- [ ] Create seeder templates:
  - Basic seeder template
  - User seeder example
  - Reference data seeder
  - Factory-based seeder

- [ ] Add faker integration (optional):
  - Generate realistic test data
  - Configurable data volumes
  - Repeatable with seeds

**Example Seeder Template**:
```go
package seeders

import "github.com/toutaio/toutago-sil-migrator/pkg/sil"

type UserSeeder struct {
    sil.BaseSeeder
}

func (s *UserSeeder) Name() string {
    return "UserSeeder"
}

func (s *UserSeeder) Dependencies() []string {
    return []string{} // No dependencies
}

func (s *UserSeeder) Environments() []string {
    return []string{"development", "test"}
}

func (s *UserSeeder) Seed(adapter sil.DatabaseAdapter) error {
    // Insert seed data
    return adapter.Exec(`INSERT INTO users (name, email) VALUES (?, ?)`, "Alice", "alice@example.com")
}

func (s *UserSeeder) ShouldRun(adapter sil.DatabaseAdapter) (bool, error) {
    // Check if already run
    rows, err := adapter.Query(`SELECT COUNT(*) FROM users WHERE email = ?`, "alice@example.com")
    if err != nil {
        return false, err
    }
    // Logic to determine if should run
    return count == 0, nil
}
```

**Deliverables**:
- ✅ Seeder generator
- ✅ Template library
- ✅ Example seeders

---

### Week 7: Environment Support & CLI

#### 3.4 Environment-Specific Seeds
**Duration**: 2 days  
**Priority**: High

**Tasks**:
- [ ] Implement environment detection:
  - Read from `SIL_ENVIRONMENT` env var
  - Read from config file
  - Read from CLI flag `--env`
  - Default to "development"

- [ ] Implement environment filtering:
  - Check `Environments()` method
  - Skip seeders not for current environment
  - Log skipped seeders with reason

- [ ] Add production safety:
  - Require confirmation for production seeding
  - Display warning about destructive operations
  - Implement `--force` flag to bypass

- [ ] Create environment-specific examples:
  - Development seeds (large test datasets)
  - Test seeds (minimal data)
  - Staging seeds (production-like)
  - Production seeds (reference data only)

**Deliverables**:
- ✅ Environment detection
- ✅ Environment filtering
- ✅ Production safeguards

---

#### 3.5 Seed CLI Commands
**Duration**: 1 day  
**Priority**: High

**Tasks**:
- [ ] Implement `sil seed:create <name>` command:
  - Generate seeder file
  - Use template
  - Place in seeders directory

- [ ] Implement `sil seed:run [seeders...]` command:
  - Run specific seeders
  - Resolve dependencies automatically
  - Display execution results

- [ ] Implement `sil seed:run --all` command:
  - Run all seeders
  - Apply environment filtering
  - Show progress

- [ ] Implement `sil seed:status` command:
  - List all seeders
  - Show execution status
  - Display last run time

- [ ] Add flags:
  - `--env` - Specify environment
  - `--force` - Force re-run (ignore `ShouldRun()`)
  - `--dry-run` - Preview without executing

**Deliverables**:
- ✅ All seed commands implemented
- ✅ Flags working
- ✅ User-friendly output

---

#### 3.6 Testing - Phase 3
**Duration**: 2 days  
**Priority**: High

**Tasks**:
- [ ] Write unit tests:
  - `dependency_resolver_test.go` - Graph building, topological sort
  - `seed_manager_test.go` - Seeder execution logic
  - `circular_dependency_test.go` - Detect circular deps

- [ ] Write integration tests:
  - Test seeder execution with real database
  - Test idempotency (`ShouldRun()` working)
  - Test environment filtering
  - Test dependency ordering

- [ ] Test edge cases:
  - Empty dependencies
  - Circular dependencies
  - Missing dependencies
  - Multiple execution attempts

- [ ] Achieve 90%+ code coverage

**Deliverables**:
- ✅ Comprehensive seeder tests
- ✅ 90%+ coverage
- ✅ Edge cases handled

---

#### 3.7 Documentation - Phase 3
**Duration**: 1 day  
**Priority**: High

**Tasks**:
- [ ] Write seeder creation guide:
  - Creating seeders
  - Naming conventions
  - Best practices

- [ ] Document dependency management:
  - Declaring dependencies
  - Dependency resolution
  - Avoiding circular deps

- [ ] Document idempotency patterns:
  - Implementing `ShouldRun()`
  - Check-then-insert pattern
  - Upsert strategies

- [ ] Create seeder best practices guide:
  - When to use seeders
  - Seeder vs migration
  - Performance considerations

- [ ] Add seeder examples:
  - User seeder
  - Role seeder with dependencies
  - Bulk data seeder
  - Reference data seeder

- [ ] Document environment-specific seeding:
  - Environment configuration
  - Safety measures
  - Deployment strategies

**Deliverables**:
- ✅ Complete seeder documentation
- ✅ Best practices guide
- ✅ Example seeders

---

### Phase 3 Success Criteria

✅ Seeder engine functional  
✅ Dependency resolution working  
✅ Idempotency checks reliable  
✅ Environment filtering operational  
✅ CLI seed commands complete  
✅ 90%+ test coverage achieved  
✅ Documentation comprehensive  

---

## Phase 4: Advanced Features (Weeks 8-9)

**Goal**: Add production-ready features, optimization, and comprehensive examples.

### Week 8: Advanced Features

#### 4.1 Dry-Run Mode
**Duration**: 1 day  
**Priority**: Medium

**Tasks**:
- [ ] Implement dry-run for migrations:
  - Capture SQL statements without executing
  - Display migration plan
  - Show order of execution
  - Estimate impact (tables/columns affected)

- [ ] Add to CLI:
  - `--dry-run` flag for `sil migrate`
  - Formatted output of SQL
  - Summary of changes

- [ ] Create dry-run example and documentation

**Deliverables**:
- ✅ Dry-run mode functional
- ✅ SQL preview working
- ✅ Documentation

---

#### 4.2 Migration Helpers
**Duration**: 2 days  
**Priority**: Medium

**Tasks**:
- [ ] Create `pkg/sil/helpers.go`:

```go
// Schema helpers
func CreateTable(adapter DatabaseAdapter, name string, columns ...Column) error
func DropTable(adapter DatabaseAdapter, name string) error
func AddColumn(adapter DatabaseAdapter, table, column string, columnType string) error
func RemoveColumn(adapter DatabaseAdapter, table, column string) error
func AddIndex(adapter DatabaseAdapter, table string, columns ...string) error

// Common patterns
func AddTimestamps(adapter DatabaseAdapter, table string) error
func AddSoftDeletes(adapter DatabaseAdapter, table string) error

// Batch processing
func BatchInsert(adapter DatabaseAdapter, table string, records []map[string]interface{}, batchSize int) error
func BatchUpdate(adapter DatabaseAdapter, query string, args [][]interface{}, batchSize int) error
```

- [ ] Implement helpers for common patterns:
  - Timestamps (created_at, updated_at)
  - Soft deletes (deleted_at)
  - UUID primary keys
  - Foreign keys

- [ ] Add batch processing helpers:
  - Chunked inserts
  - Progress reporting
  - Error handling

- [ ] Write tests for all helpers
- [ ] Document helper usage

**Deliverables**:
- ✅ Migration helpers library
- ✅ Common patterns implemented
- ✅ Batch processing support

---

#### 4.3 Advanced Locking Features
**Duration**: 1 day  
**Priority**: Low

**Tasks**:
- [ ] Implement `sil lock:status` command:
  - Show current lock status
  - Display lock holder information
  - Show lock acquisition time

- [ ] Implement `sil lock:release` command:
  - Emergency unlock functionality
  - Require confirmation
  - Log manual interventions

- [ ] Add lock monitoring:
  - Track lock duration
  - Alert on long-running locks
  - Log lock events

**Deliverables**:
- ✅ Lock management commands
- ✅ Lock monitoring
- ✅ Emergency unlock

---

#### 4.4 Error Handling & Recovery
**Duration**: 1 day  
**Priority**: High

**Tasks**:
- [ ] Enhance error reporting:
  - Detailed error messages
  - Stack traces in verbose mode
  - Error context (which migration, which line)

- [ ] Add recovery suggestions:
  - Suggest fixes for common errors
  - Link to troubleshooting docs
  - Provide migration rollback info

- [ ] Implement migration checksums:
  - Hash migration file contents
  - Detect modified migrations
  - Warn on checksum mismatches

- [ ] Create error handling guide:
  - Common errors and solutions
  - Recovery procedures
  - When to rollback vs fix forward

**Deliverables**:
- ✅ Enhanced error reporting
- ✅ Recovery suggestions
- ✅ Checksum validation

---

### Week 9: Optimization & Release

#### 4.5 Performance Optimization
**Duration**: 2 days  
**Priority**: Medium

**Tasks**:
- [ ] Optimize migration file loading:
  - Lazy loading
  - Caching migration metadata
  - Parallel loading (if safe)

- [ ] Optimize database queries:
  - Index on sil_migrations table
  - Batch queries where possible
  - Connection pooling tuning

- [ ] Implement parallel seeding:
  - Identify independent seeders
  - Run in parallel where safe
  - Configurable parallelism

- [ ] Run performance benchmarks:
  - Large migration sets (100+ migrations)
  - Bulk data seeders (1M+ records)
  - Concurrent execution stress tests

- [ ] Document performance tuning:
  - Configuration recommendations
  - Optimization strategies
  - Bottleneck identification

**Deliverables**:
- ✅ Performance optimizations
- ✅ Benchmark results
- ✅ Tuning guide

---

#### 4.6 Programmatic API
**Duration**: 1 day  
**Priority**: Medium

**Tasks**:
- [ ] Create programmatic API:

```go
// NewMigrator creates a migrator instance
func NewMigrator(config Config, adapter DatabaseAdapter) (Migrator, error)

// Event callbacks
type MigrationCallback func(migration Migration, direction string) error

func (m *Migrator) OnBeforeMigration(callback MigrationCallback)
func (m *Migrator) OnAfterMigration(callback MigrationCallback)
func (m *Migrator) OnMigrationError(callback func(migration Migration, err error))

// Custom logger
type Logger interface {
    Info(msg string, args ...interface{})
    Warn(msg string, args ...interface{})
    Error(msg string, args ...interface{})
    Debug(msg string, args ...interface{})
}

func (m *Migrator) SetLogger(logger Logger)
```

- [ ] Implement event callbacks:
  - Before migration execution
  - After migration execution
  - On migration error
  - On seeder execution

- [ ] Add custom logger support:
  - Logger interface
  - Default stdout logger
  - Integrate with popular loggers (logrus, zap)

- [ ] Create programmatic usage examples:
  - Embed in Go application
  - Custom callbacks
  - Custom logger

- [ ] Document library usage patterns

**Deliverables**:
- ✅ Programmatic API
- ✅ Event callbacks
- ✅ Custom logger support
- ✅ Usage examples

---

#### 4.7 Examples & Integration
**Duration**: 2 days  
**Priority**: High

**Tasks**:
- [ ] Create `examples/basic/`:
  - Simple migration example
  - PostgreSQL usage
  - CLI commands demonstration

- [ ] Create `examples/multi-db/`:
  - Multiple database example
  - Different adapters
  - Configuration examples

- [ ] Create `examples/with-datamapper/`:
  - Datamapper integration
  - Shared configuration
  - Migration execution with datamapper

- [ ] Create `examples/with-touta/`:
  - Toutā integration example
  - CLI wrapper commands
  - Project template

- [ ] Create Docker example:
  - Dockerfile for migrations
  - Docker Compose setup
  - Migration on container startup

- [ ] Create CI/CD integration example:
  - GitHub Actions workflow
  - GitLab CI pipeline
  - Automated migrations on deploy

**Deliverables**:
- ✅ 6 complete examples
- ✅ README for each example
- ✅ Working code

---

#### 4.8 Testing - Phase 4
**Duration**: 1 day  
**Priority**: High

**Tasks**:
- [ ] Write E2E tests:
  - Full migration lifecycle
  - CLI command integration
  - Multi-database scenarios
  - Seeding with dependencies

- [ ] Write stress tests:
  - Large migration sets
  - Concurrent executions
  - Long-running migrations

- [ ] Write dry-run tests:
  - SQL preview accuracy
  - No database modification

- [ ] Final coverage push:
  - Achieve 90%+ coverage
  - Cover edge cases
  - Integration test all adapters

**Deliverables**:
- ✅ E2E test suite
- ✅ Stress tests
- ✅ 90%+ coverage achieved

---

#### 4.9 Documentation - Phase 4
**Duration**: 2 days  
**Priority**: Critical

**Tasks**:
- [ ] Create comprehensive API documentation:
  - All public interfaces documented
  - Code examples for each method
  - Package overview

- [ ] Write advanced usage guide:
  - Complex migration scenarios
  - Production deployment strategies
  - Migration strategies (Blue/Green, Canary)

- [ ] Document all CLI commands:
  - Full command reference
  - All flags and options
  - Usage examples

- [ ] Create migration cookbook:
  - Add column
  - Remove column
  - Rename table
  - Data migration
  - Complex transformations

- [ ] Write performance tuning guide:
  - Configuration optimization
  - Large dataset handling
  - Bottleneck identification

- [ ] Create troubleshooting guide:
  - Common errors
  - Lock issues
  - Transaction failures
  - Recovery procedures

- [ ] Write contribution guide:
  - Development setup
  - Testing requirements
  - PR process
  - Code style

**Deliverables**:
- ✅ Complete API docs
- ✅ Advanced guides
- ✅ Troubleshooting resources
- ✅ Contribution guide

---

#### 4.10 Release Preparation
**Duration**: 1 day  
**Priority**: Critical

**Tasks**:
- [ ] Review all documentation:
  - Accuracy check
  - Link verification
  - Code example testing

- [ ] Create CHANGELOG.md:
  - Version 0.1.0 features
  - Breaking changes (if any)
  - Migration guide

- [ ] Prepare release artifacts:
  - Binary builds for major platforms (Linux, macOS, Windows)
  - Installation scripts
  - Docker image

- [ ] Tag v0.1.0-alpha release:
  - Git tag
  - GitHub release
  - Release notes

- [ ] Publish to GitHub:
  - Push all commits
  - Create GitHub release
  - Attach binaries

- [ ] Create announcement materials:
  - Blog post draft
  - Social media posts
  - Submit to awesome-go

**Deliverables**:
- ✅ v0.1.0-alpha released
- ✅ Binaries available
- ✅ Announcement published

---

### Phase 4 Success Criteria

✅ Dry-run mode functional  
✅ Migration helpers library complete  
✅ Advanced locking features implemented  
✅ Performance optimized  
✅ Programmatic API ready  
✅ 6 examples created  
✅ 90%+ test coverage achieved  
✅ Complete documentation  
✅ v0.1.0-alpha released  

---

## Post-Release: Toutā Integration (Optional)

### 5.1 Toutā CLI Integration
**Duration**: 2 days  
**Priority**: Low (Optional)

**Tasks**:
- [ ] Add Síl to Toutā dependencies
- [ ] Create `touta migrate` command wrapper
- [ ] Create `touta seed` command wrapper
- [ ] Document Toutā integration in both projects
- [ ] Create Toutā project template with migrations

**Deliverables**:
- ✅ Toutā CLI integration (optional)

---

## Risk Mitigation

### Technical Risks

**Risk**: Migration lock failures in production  
**Mitigation**: 
- Advisory locks auto-release on connection loss
- Lock timeout (default 5 minutes)
- Manual unlock command for emergencies
- Lock monitoring and alerting

**Risk**: Non-transactional DDL in MySQL  
**Mitigation**:
- Document MySQL DDL behavior clearly
- Provide `transactional: false` option
- Test rollback scenarios extensively
- Recommend migration splitting for complex changes

**Risk**: Concurrent migration conflicts in teams  
**Mitigation**:
- Timestamp + random suffix for uniqueness
- Conflict detection at load time
- Document branching best practices
- Provide merge/reorder tooling

**Risk**: Performance with large migrations  
**Mitigation**:
- Batch processing helpers
- Progress reporting
- Dry-run to estimate duration
- Chunking strategies documented

---

## Success Metrics

### Phase 1
- ✅ 80% test coverage
- ✅ PostgreSQL adapter functional
- ✅ CLI commands working
- ✅ Basic documentation complete

### Phase 2
- ✅ 85% test coverage
- ✅ 3 database adapters working
- ✅ Datamapper integration functional (optional)
- ✅ Multi-database examples

### Phase 3
- ✅ 90% test coverage
- ✅ Seeder system functional
- ✅ Dependency resolution working
- ✅ Environment filtering operational

### Phase 4
- ✅ 90%+ test coverage
- ✅ Production-ready features
- ✅ Performance optimized
- ✅ v0.1.0-alpha released
- ✅ 6 working examples

---

## Timeline Summary

| Phase | Duration | Deliverable |
|-------|----------|-------------|
| Phase 1: Foundation | 3 weeks | Core engine + PostgreSQL |
| Phase 2: Multi-Database | 2 weeks | MySQL + SQLite + Datamapper |
| Phase 3: Seeding | 2 weeks | Seeder system |
| Phase 4: Advanced | 2 weeks | Optimization + Release |
| **Total** | **9 weeks** | **v0.1.0-alpha** |

---

## Dependencies

### Required
- Go 1.22+
- PostgreSQL driver: `github.com/lib/pq`
- MySQL driver: `github.com/go-sql-driver/mysql`
- SQLite driver: `github.com/mattn/go-sqlite3`
- CLI framework: `github.com/spf13/cobra`
- YAML parser: `gopkg.in/yaml.v3`

### Optional
- Toutago-datamapper: `github.com/toutaio/toutago-datamapper` (with build tags)
- Testcontainers: `github.com/testcontainers/testcontainers-go` (testing)
- Faker: `github.com/jaswdr/faker` (seeding)

---

## Deployment Strategy

### Alpha Release (v0.1.0-alpha)
- Internal testing
- Early adopters
- Feedback collection

### Beta Release (v0.2.0-beta)
- Production testing
- Bug fixes
- Performance tuning

### Stable Release (v1.0.0)
- Production ready
- Long-term support
- Semantic versioning commitment

---

## Appendix: Architecture Diagram

```
┌─────────────────────────────────────────────────────┐
│                   CLI / Application                  │
│              (cmd/sil or programmatic API)           │
└─────────────────┬───────────────────────────────────┘
                  │
         ┌────────▼─────────┐
         │    Migrator      │
         │  (orchestrates)  │
         └────────┬─────────┘
                  │
    ┌─────────────┼─────────────┐
    │             │             │
┌───▼────┐  ┌────▼─────┐  ┌───▼──────┐
│Loader  │  │ Executor │  │  Lock    │
│        │  │          │  │ Manager  │
└───┬────┘  └────┬─────┘  └────┬─────┘
    │            │              │
    │       ┌────▼──────────────▼────┐
    │       │  DatabaseAdapter        │
    │       │    (interface)          │
    │       └─────────┬───────────────┘
    │                 │
    │       ┌─────────┼──────────┬────────┬──────────┐
    │       │         │          │        │          │
    │   ┌───▼──┐  ┌──▼───┐  ┌──▼───┐ ┌──▼───────┐ ┌▼────────┐
    │   │Postgres MySQL SQLite │Datamapper│ Custom │
    │   │Adapter│ Adapter Adapter Bridge    Adapter│
    │   └───┬──┘  └──┬───┘  └──┬───┘ └──┬───────┘ └┬────────┘
    │       │        │         │        │          │
    └───────┼────────┼─────────┼────────┼──────────┘
            │        │         │        │
         ┌──▼────────▼─────────▼────────▼──┐
         │      Database Layer              │
         │  (PostgreSQL/MySQL/SQLite/etc)   │
         └──────────────────────────────────┘
```

---

**Document Version**: 1.0  
**Last Updated**: 2024-12-30  
**Status**: Ready for Implementation
