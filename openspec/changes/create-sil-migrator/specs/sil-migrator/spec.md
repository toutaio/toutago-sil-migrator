# Síl Migration and Seeding System - Specification

## ADDED Requirements

### Requirement: Version-Based Migration System
The system SHALL manage database schema changes through version-based migration files that execute in chronological order.

#### Scenario: Create new migration file
- **WHEN** developer runs `sil create add_users_table`
- **THEN** a new migration file is created with format `YYYYMMDDHHMMSS_add_users_table.go`
- **AND** the file contains Up() and Down() function templates
- **AND** the file is created in the configured migrations directory

#### Scenario: Execute pending migrations
- **WHEN** developer runs `sil migrate`
- **THEN** all pending migrations are loaded from disk
- **AND** migrations are sorted by version chronologically
- **AND** each migration's Up() function is executed in order
- **AND** each successful migration is recorded in the migrations table
- **AND** execution stops on first error with clear error message

#### Scenario: Rollback last migration batch
- **WHEN** developer runs `sil rollback`
- **THEN** the last batch of executed migrations is identified
- **AND** each migration's Down() function is executed in reverse order
- **AND** each successful rollback removes the migration record from the migrations table

#### Scenario: View migration status
- **WHEN** developer runs `sil status`
- **THEN** a table is displayed showing all migrations
- **AND** each migration shows its version, description, and status (pending/applied)
- **AND** applied migrations show execution timestamp and batch number

### Requirement: Database Adapter Interface
The system SHALL provide a database-agnostic interface that allows supporting multiple database types through adapter implementations.

#### Scenario: Connect to PostgreSQL database
- **WHEN** migrator is initialized with PostgreSQL configuration
- **THEN** the PostgreSQL adapter establishes a connection
- **AND** connection pooling is configured according to settings
- **AND** migrations table is created if it doesn't exist

#### Scenario: Execute query through adapter
- **WHEN** migration calls adapter.Exec() with SQL query
- **THEN** the query is executed on the connected database
- **AND** errors are wrapped with context information
- **AND** affected rows count is returned

#### Scenario: Query data through adapter
- **WHEN** migration calls adapter.Query() with SQL query
- **THEN** the query is executed and result rows are returned
- **AND** rows can be scanned into Go types
- **AND** result set is properly closed after use

#### Scenario: Switch database adapter
- **WHEN** developer changes database type in configuration
- **THEN** the appropriate adapter is loaded based on connection string
- **AND** all migration operations work with the new adapter
- **AND** migrations table schema is compatible across adapters

### Requirement: Transaction Support
The system SHALL execute each migration within a database transaction to ensure atomicity and automatic rollback on failure.

#### Scenario: Successful migration with transaction
- **WHEN** migration Up() function is executed
- **THEN** a database transaction is started before execution
- **AND** all migration operations run within the transaction
- **AND** transaction is committed after successful execution
- **AND** migration is recorded in migrations table

#### Scenario: Failed migration with automatic rollback
- **WHEN** migration Up() function encounters an error
- **THEN** the transaction is automatically rolled back
- **AND** database state is restored to pre-migration state
- **AND** migration is NOT recorded in migrations table
- **AND** clear error message is displayed with migration details

#### Scenario: Non-transactional migration
- **WHEN** migration is marked as non-transactional (transactional: false)
- **THEN** operations execute without transaction wrapper
- **AND** warning is displayed about lack of automatic rollback
- **AND** migration is still recorded after successful execution

### Requirement: Migration Locking
The system SHALL prevent concurrent migration execution across multiple instances using database-level locks.

#### Scenario: Acquire migration lock successfully
- **WHEN** migrator starts migration process
- **THEN** database advisory lock is acquired
- **AND** lock acquisition is logged with timestamp
- **AND** migrations proceed after lock is acquired

#### Scenario: Lock already held by another instance
- **WHEN** migrator attempts to acquire lock but it's held by another instance
- **THEN** migration process waits for configured timeout period
- **AND** error message indicates migration is running elsewhere
- **AND** migrator exits without attempting migrations

#### Scenario: Lock released after completion
- **WHEN** migration process completes successfully
- **THEN** database advisory lock is released
- **AND** lock release is logged
- **AND** other instances can now acquire the lock

#### Scenario: Lock released on connection failure
- **WHEN** database connection is lost while holding lock
- **THEN** advisory lock is automatically released by database
- **AND** lock becomes available for other instances

#### Scenario: Manual lock release
- **WHEN** administrator runs `sil lock:release` command
- **THEN** current lock status is checked
- **AND** confirmation is requested before releasing
- **AND** lock is forcefully released after confirmation
- **AND** warning is logged about manual intervention

### Requirement: Migration Execution Control
The system SHALL provide fine-grained control over which migrations to execute and in what direction.

#### Scenario: Execute specific number of migrations
- **WHEN** developer runs `sil migrate --steps=3`
- **THEN** only the next 3 pending migrations are executed
- **AND** execution stops after 3 migrations regardless of remaining pending
- **AND** status shows updated migration state

#### Scenario: Rollback specific number of migrations
- **WHEN** developer runs `sil rollback --steps=2`
- **THEN** only the last 2 applied migrations are rolled back
- **AND** their Down() functions execute in reverse chronological order
- **AND** migration records are removed from migrations table

#### Scenario: Reset all migrations
- **WHEN** developer runs `sil reset`
- **THEN** confirmation is requested (unless --force flag is used)
- **AND** all applied migrations are rolled back in reverse order
- **AND** migrations table is cleared
- **AND** database is returned to empty schema state

#### Scenario: Re-run all migrations (fresh)
- **WHEN** developer runs `sil fresh`
- **THEN** all migrations are rolled back (reset)
- **AND** all migrations are executed from the beginning
- **AND** database is rebuilt from clean state

### Requirement: PostgreSQL Adapter
The system SHALL provide a production-ready PostgreSQL adapter supporting all core migration operations.

#### Scenario: Connect with connection string
- **WHEN** PostgreSQL adapter is initialized with DSN string
- **THEN** connection is established to PostgreSQL server
- **AND** connection parameters are parsed from DSN
- **AND** connection pool is configured with sensible defaults

#### Scenario: Create migrations table
- **WHEN** PostgreSQL adapter creates migrations table
- **THEN** table is created with columns: id, version, description, batch, executed_at
- **AND** version column has unique constraint
- **AND** table uses appropriate data types for PostgreSQL

#### Scenario: Use advisory locks
- **WHEN** PostgreSQL adapter acquires migration lock
- **THEN** pg_advisory_lock() is called with consistent key
- **AND** lock is held until explicitly released
- **AND** lock is automatically released if connection drops

#### Scenario: Execute DDL in transaction
- **WHEN** migration executes CREATE TABLE or ALTER TABLE
- **THEN** DDL executes within transaction (PostgreSQL supports this)
- **AND** changes are rolled back if migration fails
- **AND** changes are committed if migration succeeds

### Requirement: MySQL Adapter
The system SHALL provide a production-ready MySQL adapter supporting all core migration operations with MySQL-specific behavior.

#### Scenario: Connect to MySQL database
- **WHEN** MySQL adapter is initialized with DSN string
- **THEN** connection is established to MySQL server
- **AND** connection parameters include charset and parseTime settings
- **AND** connection pool is configured

#### Scenario: Handle DDL transaction limitations
- **WHEN** migration executes DDL statement in MySQL
- **THEN** warning is displayed about implicit commit
- **AND** transactional mode can be disabled for migration
- **AND** documentation explains MySQL DDL behavior

#### Scenario: Use GET_LOCK for locking
- **WHEN** MySQL adapter acquires migration lock
- **THEN** GET_LOCK() function is called with unique lock name
- **AND** lock timeout is configured
- **AND** lock is released with RELEASE_LOCK() after migrations

### Requirement: SQLite Adapter
The system SHALL provide a production-ready SQLite adapter supporting all core migration operations with SQLite-specific behavior.

#### Scenario: Connect to SQLite file
- **WHEN** SQLite adapter is initialized with file path
- **THEN** connection is established to SQLite database file
- **AND** file is created if it doesn't exist
- **AND** WAL mode is enabled for better concurrency

#### Scenario: Use file locking
- **WHEN** SQLite adapter acquires migration lock
- **THEN** file-based lock is acquired on database file
- **AND** exclusive lock prevents concurrent access
- **AND** lock is released after migrations complete

#### Scenario: Handle transaction limitations
- **WHEN** migration requires nested transaction
- **THEN** savepoints are used instead of nested transactions
- **AND** savepoint rollback is supported
- **AND** documentation explains SQLite transaction model

### Requirement: Seeder System
The system SHALL provide a data seeding system with dependency management and idempotency support.

#### Scenario: Create new seeder
- **WHEN** developer runs `sil seed:create UserSeeder`
- **THEN** a new seeder file is created with template
- **AND** file includes Name(), Dependencies(), and Seed() methods
- **AND** file includes ShouldRun() for idempotency checking

#### Scenario: Execute seeders in dependency order
- **WHEN** developer runs `sil seed:run --all`
- **THEN** all seeders are loaded from seeders directory
- **AND** dependency graph is built from Dependencies() declarations
- **AND** topological sort determines execution order
- **AND** seeders execute in correct dependency order

#### Scenario: Detect circular dependencies
- **WHEN** seeders have circular dependency relationships
- **THEN** error is raised before execution
- **AND** error message identifies the circular dependency chain
- **AND** no seeders are executed

#### Scenario: Skip already-run seeders
- **WHEN** seeder's ShouldRun() returns false
- **THEN** seeder execution is skipped
- **AND** skip is logged with reason
- **AND** dependent seeders still execute if their conditions are met

#### Scenario: Force re-run seeders
- **WHEN** developer runs `sil seed:run --force UserSeeder`
- **THEN** UserSeeder executes regardless of ShouldRun() result
- **AND** execution is logged as forced
- **AND** seed record is updated with new execution timestamp

### Requirement: Environment-Specific Seeding
The system SHALL support environment-specific seeders to enable different data sets per environment.

#### Scenario: Run development seeders
- **WHEN** developer runs `sil seed:run --env=development`
- **THEN** only seeders tagged for development environment execute
- **AND** production-only seeders are skipped
- **AND** environment is detected from config or flag

#### Scenario: Prevent production seeding accidentally
- **WHEN** developer runs seeders without specifying environment in production
- **THEN** confirmation is required before executing any seeders
- **AND** warning is displayed about production environment
- **AND** seeding is aborted if confirmation is not given

#### Scenario: Tag seeders for specific environments
- **WHEN** seeder declares Environments() returning ["development", "test"]
- **THEN** seeder only runs when --env matches one of declared environments
- **AND** seeder is skipped in staging and production
- **AND** skip is logged with environment reason

### Requirement: Dry-Run Mode
The system SHALL support dry-run mode to preview migration changes without applying them to the database.

#### Scenario: Dry-run migrations
- **WHEN** developer runs `sil migrate --dry-run`
- **THEN** all pending migrations are identified and loaded
- **AND** SQL statements are captured and displayed without execution
- **AND** migration plan is shown with expected order
- **AND** no changes are made to database
- **AND** no migrations are recorded

#### Scenario: Preview migration impact
- **WHEN** dry-run mode generates migration preview
- **THEN** output shows which migrations would execute
- **AND** estimated impact is displayed (new tables, columns, etc.)
- **AND** warnings are shown for potentially dangerous operations
- **AND** developer can review before actual execution

### Requirement: Migration Templates
The system SHALL provide migration templates to accelerate common migration patterns and reduce boilerplate.

#### Scenario: Generate create table migration
- **WHEN** developer runs `sil create:table users`
- **THEN** migration file is created with create_table template
- **AND** template includes common table creation patterns
- **AND** template includes drop_table in Down() function

#### Scenario: Generate add column migration
- **WHEN** developer runs `sil create:column users email`
- **THEN** migration file is created with add_column template
- **AND** template includes column addition code
- **AND** template includes column removal in Down() function

#### Scenario: Use timestamp helper
- **WHEN** migration uses add_timestamps() helper
- **THEN** created_at and updated_at columns are added
- **AND** columns use appropriate types for database adapter
- **AND** Down() function removes these columns

### Requirement: Error Handling and Recovery
The system SHALL provide comprehensive error handling with actionable recovery suggestions.

#### Scenario: Migration syntax error
- **WHEN** migration contains Go syntax error
- **THEN** error is caught during migration loading
- **AND** specific file and line number are reported
- **AND** compilation error message is displayed
- **AND** migration execution is prevented

#### Scenario: Migration runtime error
- **WHEN** migration encounters runtime error during execution
- **THEN** transaction is rolled back automatically
- **AND** error message includes migration version and description
- **AND** specific error from database is included
- **AND** recovery suggestions are provided

#### Scenario: Lock timeout during migration
- **WHEN** migration lock cannot be acquired within timeout
- **THEN** clear error message indicates another instance is running
- **AND** suggestion is provided to wait or check other instances
- **AND** manual unlock command is suggested for emergency

### Requirement: Programmatic API
The system SHALL provide a programmatic API for embedding migration functionality in Go applications.

#### Scenario: Initialize migrator programmatically
- **WHEN** application creates migrator with NewMigrator(config)
- **THEN** migrator instance is returned with all dependencies
- **AND** database connection is established
- **AND** migrator is ready to execute migrations

#### Scenario: Run migrations from application code
- **WHEN** application calls migrator.Migrate()
- **THEN** all pending migrations execute
- **AND** migration results are returned
- **AND** errors can be handled programmatically

#### Scenario: Register migration callbacks
- **WHEN** application registers BeforeMigration and AfterMigration callbacks
- **THEN** callbacks are invoked for each migration
- **AND** callbacks receive migration information
- **AND** callbacks can log, monitor, or perform custom actions

#### Scenario: Custom logger integration
- **WHEN** application provides custom logger implementing Logger interface
- **THEN** all migration logging uses custom logger
- **AND** log format and destination are controlled by application
- **AND** default logger is used if none provided

### Requirement: CLI User Experience
The system SHALL provide an intuitive command-line interface with helpful output and clear error messages.

#### Scenario: Display colorized output
- **WHEN** CLI commands execute in terminal supporting colors
- **THEN** success messages are displayed in green
- **AND** error messages are displayed in red
- **AND** warnings are displayed in yellow
- **AND** informational messages use default color

#### Scenario: Show progress for long operations
- **WHEN** migration execution takes longer than 2 seconds
- **THEN** progress indicator is displayed
- **AND** current migration being executed is shown
- **AND** elapsed time is updated in real-time

#### Scenario: Display helpful error messages
- **WHEN** command encounters error
- **THEN** error is displayed with context
- **AND** possible causes are suggested
- **AND** recovery steps are provided
- **AND** relevant documentation links are included

#### Scenario: Support verbose mode
- **WHEN** command is run with --verbose flag
- **THEN** detailed debug information is displayed
- **AND** SQL queries are logged
- **AND** connection details are shown (excluding passwords)
- **AND** timing information is included

### Requirement: Configuration Management
The system SHALL support flexible configuration through files, environment variables, and CLI flags.

#### Scenario: Load configuration from file
- **WHEN** migrator is initialized and config file exists
- **THEN** configuration is loaded from sil.yaml or sil.json
- **AND** database connection parameters are read
- **AND** migration directory path is configured
- **AND** default values are used for unspecified settings

#### Scenario: Override config with environment variables
- **WHEN** environment variables are set (SIL_DATABASE_URL, etc.)
- **THEN** environment variables override file configuration
- **AND** sensitive values can be kept in environment only
- **AND** variable names follow SIL_ prefix convention

#### Scenario: Override config with CLI flags
- **WHEN** CLI flags are provided (--database, --migrations-dir)
- **THEN** CLI flags take precedence over all other sources
- **AND** flag values override both file and environment config
- **AND** help text documents all available flags

#### Scenario: Validate configuration
- **WHEN** configuration is loaded from any source
- **THEN** required fields are validated (database connection)
- **AND** clear error is shown for missing required configuration
- **AND** invalid values are rejected with explanation
- **AND** warnings are shown for deprecated settings

### Requirement: Testing Support
The system SHALL provide utilities and patterns to facilitate testing database migrations and seeders.

#### Scenario: Mock database adapter for testing
- **WHEN** tests create migrator with MockAdapter
- **THEN** migrations execute without real database
- **AND** SQL queries are captured for assertions
- **AND** transaction behavior can be simulated
- **AND** errors can be injected for failure testing

#### Scenario: Integration testing with real database
- **WHEN** integration tests run migrations against test database
- **THEN** test database is created and cleaned up automatically
- **AND** migrations execute in isolated environment
- **AND** test data can be verified after migration
- **AND** database is reset between test runs

#### Scenario: Test migration rollback
- **WHEN** test executes migration Up() then Down()
- **THEN** database returns to initial state
- **AND** test verifies schema matches pre-migration state
- **AND** test verifies no orphaned data or tables remain

### Requirement: Documentation and Examples
The system SHALL provide comprehensive documentation and practical examples for common use cases.

#### Scenario: Quick start guide
- **WHEN** new user reads README
- **THEN** installation instructions are clear
- **AND** basic usage example is provided
- **AND** first migration can be created and run in under 5 minutes
- **AND** links to detailed documentation are included

#### Scenario: Example migrations
- **WHEN** developer explores examples directory
- **THEN** examples cover common migration patterns
- **AND** examples include create table, alter table, data migration
- **AND** examples demonstrate proper Up/Down implementation
- **AND** examples are well-commented

#### Scenario: API documentation
- **WHEN** developer browses GoDoc
- **THEN** all public interfaces and types are documented
- **AND** usage examples are included in comments
- **AND** package overview explains core concepts
- **AND** method documentation explains parameters and return values

#### Scenario: Troubleshooting guide
- **WHEN** developer encounters common issue
- **THEN** troubleshooting guide has relevant section
- **AND** issue symptoms are clearly described
- **AND** solutions are provided with examples
- **AND** related documentation is linked
