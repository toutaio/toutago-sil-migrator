# Design: Síl Migration and Seeding System

## Context

Síl is designed as a standalone database migration and seeding tool inspired by the best practices from:
- **Rails ActiveRecord Migrations**: Sequential versioning, up/down migrations
- **Laravel Migrations**: Fluent schema builder, rollback capabilities
- **Alembic (Python)**: Branching and merging migration paths
- **Flyway**: Version-based migrations with checksums
- **Knex.js**: Transaction-wrapped migrations, seed management

The system must be:
- **Independent**: Usable in any Go project, not coupled to Toutā
- **Robust**: Handle edge cases, concurrent access, and failure scenarios
- **Flexible**: Support multiple databases and custom adapters
- **Developer-friendly**: Clear CLI, helpful error messages, good DX

### Stakeholders
- **Toutā developers**: Need reliable schema management
- **External Go developers**: Want a modern migration tool
- **DevOps teams**: Need safe production deployment tools

### Constraints
- Must work with Go 1.21+
- Must support SQL and NoSQL databases
- Must prevent data loss during migrations
- Must handle distributed environments (multiple instances)

## Goals / Non-Goals

### Goals
- Create a production-ready migration and seeding system
- Support major SQL databases (PostgreSQL, MySQL, SQLite) in Phase 1-2
- Provide both CLI and programmatic APIs
- Implement migration locking to prevent concurrent runs
- Enable safe rollback capabilities
- Support idempotent seeders with dependency ordering
- Maintain zero dependencies on Toutā framework
- Achieve 90%+ test coverage

### Non-Goals
- Not a complete ORM (focus on migrations/seeds only)
- Not a query builder (use raw SQL or separate tools)
- Not a schema introspection tool (focused on forward evolution)
- Not supporting database version < 5 years old
- Not providing GUI interface (CLI and API only)

## Decisions

### Decision 1: Version-Based Sequential Migrations
**What**: Each migration is a timestamped file with up/down functions.

**Why**: 
- Industry standard approach (Rails, Laravel, Alembic)
- Clear chronological order
- Easy to understand and debug
- Works well with version control systems

**Format**: `YYYYMMDDHHMMSS_description.go`

**Alternatives considered**:
- **Hash-based versions**: More complex, harder to understand chronologically
- **Semantic versioning**: Doesn't map well to chronological deployment order

### Decision 2: Interface-Based Database Adapters
**What**: All database operations go through a `DatabaseAdapter` interface.

**Why**:
- Follows SOLID principles (Open/closed, Dependency inversion)
- Easy to add new databases
- Simplifies testing with mock adapters
- Decouples migration logic from database specifics

**Interface**:
```go
type DatabaseAdapter interface {
    Connect(config Config) error
    Close() error
    Exec(query string, args ...interface{}) error
    Query(query string, args ...interface{}) (Rows, error)
    BeginTx() (Transaction, error)
    CreateMigrationsTable() error
    GetAppliedMigrations() ([]Migration, error)
    LockMigrations() (Lock, error)
}
```

**Alternatives considered**:
- **Direct database/sql usage**: Couples implementation to specific drivers
- **Third-party ORM**: Adds unnecessary dependencies and complexity

### Decision 3: File-Based Migration Storage
**What**: Migrations are Go source files in a designated directory.

**Why**:
- Type safety and compile-time checks
- Full power of Go (conditionals, loops, helpers)
- IDE support and refactoring capabilities
- No DSL to learn

**Alternatives considered**:
- **SQL files only**: Limited flexibility, no conditionals, error handling
- **YAML/JSON DSL**: Requires custom parser, limited expressiveness
- **Embedded in code**: Harder to track, version, and organize

### Decision 4: Distributed Migration Locking
**What**: Database-level advisory locks prevent concurrent migrations.

**Why**:
- Prevents race conditions in multi-instance deployments
- Uses native database features (no external dependencies)
- Automatic cleanup on connection loss
- Industry standard approach (Flyway, Liquibase)

**Implementation per database**:
- PostgreSQL: `pg_advisory_lock()`
- MySQL: `GET_LOCK()`
- SQLite: File locking
- MongoDB: Unique index on lock collection

**Alternatives considered**:
- **File-based locks**: Don't work in containerized/distributed environments
- **External lock service (Redis, etcd)**: Adds external dependencies
- **No locking**: Unsafe for production

### Decision 5: Transaction-Wrapped Migrations
**What**: Each migration runs in a database transaction with auto-rollback on error.

**Why**:
- Atomic operations - all-or-nothing
- Automatic cleanup on failure
- Prevents partial migrations
- Standard practice in migration tools

**Caveats**:
- Some DDL operations don't support transactions (documented per database)
- Non-transactional mode available when needed

### Decision 6: Seeder Dependency Graph
**What**: Seeders declare dependencies, executed in topological order.

**Why**:
- Ensures foreign key relationships work correctly
- Makes seeder order explicit and predictable
- Prevents runtime errors from missing data

**Example**:
```go
type UserSeeder struct{}
func (s *UserSeeder) Dependencies() []string { return []string{} }

type PostSeeder struct{}
func (s *PostSeeder) Dependencies() []string { return []string{"UserSeeder"} }
```

**Alternatives considered**:
- **Manual ordering**: Error-prone, hard to maintain
- **Alphabetical**: Doesn't reflect actual dependencies

### Decision 7: Phased Development Approach

**Phase 1: Foundation (Weeks 1-3)**
- Core migration engine
- PostgreSQL adapter (most feature-complete)
- Basic CLI (`sil migrate`, `sil rollback`, `sil status`)
- Migration locking
- Transaction support

**Phase 2: Multi-Database (Weeks 4-5)**
- MySQL adapter
- SQLite adapter
- Adapter testing framework
- Database-specific documentation

**Phase 3: Seeding System (Weeks 6-7)**
- Seeder interface and engine
- Dependency resolution
- Idempotent seed execution
- Seed CLI commands

**Phase 4: Advanced Features (Weeks 8-9)**
- Dry-run mode
- Migration squashing
- Environment-specific seeds
- Performance optimization
- Comprehensive examples

## Architecture

### Directory Structure
```
toutago-sil-migrator/
├── README.md
├── LICENSE
├── go.mod
├── cmd/
│   └── sil/                    # CLI binary
│       └── main.go
├── pkg/
│   └── sil/
│       ├── migrator.go         # Core migration engine
│       ├── seeder.go           # Seeding engine
│       ├── interfaces.go       # All interfaces
│       ├── migration.go        # Migration types
│       ├── config.go           # Configuration
│       ├── lock.go             # Locking mechanism
│       ├── adapters/
│       │   ├── postgres.go     # PostgreSQL adapter
│       │   ├── mysql.go        # MySQL adapter
│       │   ├── sqlite.go       # SQLite adapter
│       │   └── mock.go         # Testing adapter
│       └── templates/          # Migration templates
├── examples/
│   ├── basic/                  # Basic usage example
│   ├── multi-db/               # Multi-database example
│   └── with-touta/             # Toutā integration example
├── migrations/                  # Example migrations
└── tests/
    ├── unit/
    ├── integration/
    └── e2e/
```

### Core Interfaces

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

// Seeder represents a data seeder
type Seeder interface {
    Name() string
    Dependencies() []string
    Seed(adapter DatabaseAdapter) error
    ShouldRun(adapter DatabaseAdapter) (bool, error)
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

// SeedManager coordinates seeder execution
type SeedManager interface {
    Seed(seeders ...string) error
    SeedAll() error
    Status() ([]SeedStatus, error)
}
```

### Migration Flow

```
1. Initialize Migrator
   ↓
2. Acquire migration lock
   ↓
3. Load migration files from disk
   ↓
4. Query applied migrations from DB
   ↓
5. Calculate pending migrations
   ↓
6. For each pending migration:
   ├─ Begin transaction
   ├─ Execute Up() function
   ├─ Record migration in DB
   ├─ Commit transaction
   └─ (on error) Rollback transaction
   ↓
7. Release lock
```

### Seeding Flow

```
1. Initialize SeedManager
   ↓
2. Load all seeders
   ↓
3. Build dependency graph
   ↓
4. Topological sort
   ↓
5. For each seeder (in order):
   ├─ Check ShouldRun()
   ├─ (if true) Execute Seed()
   └─ Record execution
   ↓
6. Report results
```

## Risks / Trade-offs

### Risk 1: Migration Lock Failures
**Risk**: Database connection drops while holding migration lock.

**Mitigation**:
- Use advisory locks (auto-release on connection loss)
- Implement lock timeout (default 5 minutes)
- Add manual unlock command for emergency
- Monitor lock duration in production

### Risk 2: Non-Transactional DDL
**Risk**: Some databases don't support transactional DDL (MySQL DDL, etc.)

**Mitigation**:
- Document which operations are non-transactional per database
- Provide `transactional: false` option for migrations
- Test rollback scenarios extensively
- Recommend migration splitting for complex changes

### Risk 3: Migration Conflicts in Teams
**Risk**: Multiple developers create migrations with same/conflicting timestamps.

**Mitigation**:
- Use timestamp + random suffix for uniqueness
- Detect conflicts at migration load time
- Provide merge/reorder tooling
- Document branching best practices

### Risk 4: Performance with Large Migrations
**Risk**: Large data migrations timeout or lock tables for too long.

**Mitigation**:
- Support batch processing in migrations
- Add progress reporting for long operations
- Document chunking strategies
- Provide dry-run to estimate duration

### Trade-off 1: Go Files vs SQL Files
**Choice**: Go files with full programming capabilities

**Trade-off**:
- ✅ Type safety, IDE support, full Go power
- ❌ Requires compilation, more verbose than SQL
- ❌ Can't easily preview SQL without running

**Decision**: Worth it for type safety and flexibility. Provide templates to reduce verbosity.

### Trade-off 2: Database-Specific Features
**Choice**: Support database-specific features through adapter extensions

**Trade-off**:
- ✅ Leverage unique database capabilities
- ❌ Reduces portability between databases
- ❌ More complex adapter implementations

**Decision**: Support extensions but encourage portable patterns in documentation.

## Migration Plan

### Phase 1 Deliverables
- [ ] Core migration engine with PostgreSQL
- [ ] CLI with basic commands
- [ ] Migration locking
- [ ] Transaction support
- [ ] Unit and integration tests
- [ ] Basic documentation

### Phase 2 Deliverables
- [ ] MySQL adapter
- [ ] SQLite adapter
- [ ] Multi-database testing
- [ ] Adapter documentation

### Phase 3 Deliverables
- [ ] Seeder engine
- [ ] Dependency resolution
- [ ] Seeder CLI commands
- [ ] Seeding examples

### Phase 4 Deliverables
- [ ] Dry-run mode
- [ ] Advanced features
- [ ] Performance optimization
- [ ] Comprehensive examples
- [ ] Integration with Toutā

### Rollback Plan
If Síl doesn't meet requirements:
- Keep as experimental repository
- Don't integrate into Toutā
- Users can still import directly if desired
- No breaking changes to Toutā

### Backwards Compatibility
- Semantic versioning (v0.x.x initially)
- Clear migration path for breaking changes
- Deprecation warnings before removals
- Maintain v1.x compatibility once stable

## Open Questions

1. **Question**: Should we support branching migrations (like Alembic)?
   - **Answer needed by**: End of Phase 1
   - **Impact**: Architecture complexity, use case coverage

2. **Question**: Should migrations support reversible data transformations automatically?
   - **Answer needed by**: Phase 1 design
   - **Impact**: API design, complexity

3. **Question**: What's the strategy for schema drift detection?
   - **Answer needed by**: Phase 4
   - **Impact**: Additional tooling needed

4. **Question**: Should we support MongoDB as a first-class citizen?
   - **Answer needed by**: Phase 2 planning
   - **Impact**: Adapter complexity, NoSQL patterns

5. **Question**: CLI standalone binary vs library-only approach?
   - **Answer needed by**: Phase 1 start
   - **Impact**: Build process, distribution strategy
   - **Current leaning**: Both (library + optional CLI binary)

## References

- Rails Migrations: https://guides.rubyonrails.org/active_record_migrations.html
- Laravel Migrations: https://laravel.com/docs/migrations
- Flyway: https://flywaydb.org/
- Alembic: https://alembic.sqlalchemy.org/
- Knex.js: http://knexjs.org/
