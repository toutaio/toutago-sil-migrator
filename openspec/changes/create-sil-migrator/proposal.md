# Change: Create Síl - Standalone Database Migration and Seeding Component

## Why

Toutā requires a robust, independent database migration and seeding system that can:
- Manage schema evolution across development, staging, and production environments
- Provide reliable data seeding capabilities for testing and deployment
- Work as a standalone tool that can be used in any Go project, not just Toutā
- Follow SOLID principles and best practices from proven migration tools

The component must be decoupled from Toutā's core to enable:
- Independent versioning and releases
- Use in external projects
- Easier testing and maintenance
- Clear separation of concerns

## What Changes

This proposal creates **Síl** (Old Irish for "seed" or "lineage") as an independent Go component with the following capabilities:

### Core Features
- **Version-based migrations**: Sequential migration files with up/down operations
- **Rollback support**: Safe rollback to any previous migration state
- **Database-agnostic interface**: Support for PostgreSQL, MySQL, SQLite, MongoDB, and custom adapters
- **Seeding system**: Reusable, idempotent data seeders with dependency management
- **CLI tool**: Standalone binary for migration management (`sil` command)
- **Programmatic API**: Library interface for integration into applications
- **Migration locking**: Prevent concurrent migrations across multiple instances
- **Dry-run mode**: Preview migration changes without applying them
- **Transaction support**: Atomic migrations with automatic rollback on failure

### Technical Approach
- **Repository**: Independent repository at `github.com/toutaio/toutago-sil-migrator`
- **Development location**: `/home/nestor/Proyects/toutago-sil-migrator`
- **Go modules**: Independent versioning with semantic versioning
- **Interface-driven**: All database operations through adapter interfaces
- **SOLID compliance**: Single responsibility, Open/closed, Liskov substitution, Interface segregation, Dependency inversion
- **Zero dependencies on Toutā**: Can be used as standalone library

### Phased Implementation
- **Phase 1**: Core migration engine and PostgreSQL adapter
- **Phase 2**: Multi-database support and advanced features
- **Phase 3**: Seeding system and environment management
- **Phase 4**: Advanced features and optimization

### Integration Points
- Toutā will import Síl as a standard Go dependency
- Optional Toutā CLI integration via `touta migrate` and `touta seed` commands
- Storage adapters can leverage Síl for schema management

## Impact

### New Components
- **New repository**: `github.com/toutaio/toutago-sil-migrator`
- **New capability spec**: `sil-migrator` (defines requirements and behavior)
- **New CLI binary**: `sil` (standalone migration tool)
- **New Go package**: Importable library for migration management

### Affected Toutā Components
- **Toutā CLI** (optional): New `touta migrate` and `touta seed` command integration
- **Storage Adapters** (future): Can use Síl for schema initialization
- **Ritual Templates** (future): Can include migrations and seeds

### Breaking Changes
- **None** - This is a new component with no breaking changes to existing Toutā functionality

### Documentation Impact
- New repository README and documentation
- Toutā documentation update to reference Síl for database management
- Migration guide for projects using database storage

### Timeline
- Phase 1: 2-3 weeks (core engine + PostgreSQL)
- Phase 2: 2 weeks (multi-database support)
- Phase 3: 2 weeks (seeding system)
- Phase 4: 1-2 weeks (optimization and polish)

### Success Criteria
- [ ] Síl can run independently without Toutā
- [ ] At least 3 database adapters implemented
- [ ] 90%+ test coverage
- [ ] Production-ready migration locking
- [ ] Comprehensive documentation with examples
- [ ] Successful integration example in Toutā project
