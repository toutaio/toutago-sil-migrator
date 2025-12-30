# Project Context

## Purpose
**Síl** (Old Irish for "seed" or "lineage") is a standalone database migration and seeding tool for Go projects. It provides:
- Robust schema evolution management across all environments
- Reliable data seeding capabilities for testing and deployment
- Production-ready distributed migration locking
- Support for multiple databases (PostgreSQL, MySQL, SQLite, and extensible to NoSQL)

Síl is designed to be independent from Toutā, enabling use in any Go project while following SOLID principles and best practices from proven migration tools like Rails, Laravel, Flyway, and Alembic.

## Tech Stack
- **Language**: Go 1.22+
- **Primary Databases**: PostgreSQL, MySQL, SQLite (extensible to MongoDB and others)
- **Architecture**: Interface-driven adapters, dependency injection
- **Testing**: Unit, integration, and E2E tests targeting 90%+ coverage
- **Distribution**: CLI binary (`sil`) and importable Go library
- **Optional Integration**: Toutago-datamapper bridge adapter for seamless integration

## Project Conventions

### Code Style
- Follow standard Go conventions (`gofmt`, `golint`)
- Use clear, descriptive names for interfaces and types
- Minimize comments except where needed for clarification
- Migration filenames: `YYYYMMDDHHMMSS_description.go`
- Package structure: `pkg/sil/` for core library, `cmd/sil/` for CLI

### Architecture Patterns
- **SOLID Principles**: Single responsibility, Open/closed, Liskov substitution, Interface segregation, Dependency inversion
- **Interface-driven design**: All database operations through `DatabaseAdapter` interface
- **File-based migrations**: Go source files for type safety and IDE support
- **Transaction-wrapped operations**: Atomic migrations with auto-rollback on failure
- **Dependency injection**: Adapters and components injected, not hard-coded

### Testing Strategy
- Target 90%+ test coverage
- Unit tests for core logic (migration engine, dependency resolution)
- Integration tests per database adapter
- E2E tests for CLI commands
- Mock adapters for testing without real databases
- Test rollback scenarios extensively

### Git Workflow
- Semantic versioning (v0.x.x during development, v1.x.x when stable)
- Feature branches for new capabilities
- Clear commit messages describing changes
- Follow OpenSpec process for major changes (see `openspec/AGENTS.md`)

## Domain Context
- **Migrations**: Sequential, version-based schema changes with up/down operations
- **Seeders**: Reusable, idempotent data population with dependency ordering
- **Migration Locking**: Database-level advisory locks prevent concurrent migrations in distributed environments
- **Adapters**: Database-specific implementations of the `DatabaseAdapter` interface
- **Phases**: Development organized in 4 phases (Foundation, Multi-DB, Seeding, Advanced Features)

### Key Terminology
- **Up/Down**: Migration operations for applying and reverting schema changes
- **Dry-run**: Preview migration changes without executing them
- **Topological sort**: Ordering seeders based on dependency graph
- **Advisory lock**: Database-native locking mechanism (e.g., PostgreSQL's `pg_advisory_lock()`)

## Important Constraints
- Must work with Go 1.22 or later
- Zero required dependencies (fully standalone)
- Optional datamapper integration through bridge adapter
- Must prevent data loss during migrations
- Must handle distributed environments (multiple instances)
- Must support both SQL and NoSQL databases through adapters
- Some DDL operations are non-transactional (database-specific, must be documented)
- Not supporting database versions older than 5 years

## External Dependencies
- **Toutā (optional)**: Can integrate with Toutā CLI via `touta migrate` and `touta seed` commands
- **Toutago-datamapper (optional)**: Bridge adapter enables using datamapper's database adapters (MySQL, PostgreSQL)
- **Database drivers**: Standard Go database drivers (`database/sql` for SQL databases)
- **No external lock services**: Uses native database locking features only
- Repository: `github.com/toutaio/toutago-sil-migrator`
- Development location: `/home/nestor/Proyects/toutago-sil-migrator`
