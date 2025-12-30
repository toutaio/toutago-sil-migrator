# Síl - Database Migration and Seeding Tool

[![CI](https://github.com/toutaio/toutago-sil-migrator/actions/workflows/ci.yml/badge.svg)](https://github.com/toutaio/toutago-sil-migrator/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/toutaio/toutago-sil-migrator.svg)](https://pkg.go.dev/github.com/toutaio/toutago-sil-migrator)
[![Go Report Card](https://goreportcard.com/badge/github.com/toutaio/toutago-sil-migrator)](https://goreportcard.com/report/github.com/toutaio/toutago-sil-migrator)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

**Síl** (Old Irish for "seed" or "lineage") is a standalone database migration and seeding tool for Go projects, inspired by Rails, Laravel, Flyway, and Alembic.

> ⚠️ **Status**: Under active development - Phase 1 (Foundation)

## Core Philosophy

- **Zero Required Dependencies**: Fully standalone, works with any Go project
- **Interface-Driven**: SOLID principles, pluggable database adapters
- **Production-Ready**: Distributed locking, transaction safety, comprehensive testing
- **Developer-Friendly**: Clear CLI, helpful error messages, great DX
- **Optional Integrations**: Seamless integration with toutago-datamapper (opt-in)

## Features

### Current (Phase 1)
- ✅ **Version-Based Migrations**: Sequential, timestamped migration files
- ✅ **PostgreSQL Support**: Production-ready PostgreSQL adapter
- ✅ **Transaction Safety**: Atomic migrations with auto-rollback
- ✅ **Distributed Locking**: Prevent concurrent migrations in multi-instance deployments
- ✅ **CLI Tool**: Intuitive command-line interface
- ✅ **Programmatic API**: Embed in Go applications

### Planned
- 🔄 **Multi-Database**: MySQL, SQLite adapters (Phase 2)
- 🔄 **Data Seeding**: Idempotent seeders with dependency management (Phase 3)
- 🔄 **Advanced Features**: Dry-run mode, migration helpers, performance optimization (Phase 4)
- 🔄 **Datamapper Bridge**: Optional integration with toutago-datamapper (Phase 2)

## Quick Start

### Installation

```bash
go get github.com/toutaio/toutago-sil-migrator
```

### Requirements
- Go 1.22 or higher
- PostgreSQL 10+ (Phase 1)

### Basic Usage

```bash
# Initialize migrations directory
sil init

# Create a new migration
sil create add_users_table

# Run pending migrations
sil migrate

# Check migration status
sil status

# Rollback last batch
sil rollback
```

### Configuration

Create `sil.yaml` in your project root:

```yaml
database_url: "postgres://user:pass@localhost:5432/mydb"
migrations_dir: "./migrations"
lock_timeout: 300s
migration_timeout: 1800s
max_connections: 10
environment: development
```

### Creating Migrations

```go
package migrations

import "github.com/toutaio/toutago-sil-migrator/pkg/sil"

func init() {
    sil.RegisterMigration(&Migration_20241230120000_add_users_table{})
}

type Migration_20241230120000_add_users_table struct {
    sil.BaseMigration
}

func (m *Migration_20241230120000_add_users_table) Up(adapter sil.DatabaseAdapter) error {
    return adapter.Exec(`
        CREATE TABLE users (
            id SERIAL PRIMARY KEY,
            name VARCHAR(255) NOT NULL,
            email VARCHAR(255) UNIQUE NOT NULL,
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        )
    `)
}

func (m *Migration_20241230120000_add_users_table) Down(adapter sil.DatabaseAdapter) error {
    return adapter.Exec(`DROP TABLE users`)
}
```

## Architecture

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
    │       ┌─────────┼──────────┬────────┐
    │       │         │          │        │
    │   ┌───▼──┐  ┌──▼───┐  ┌──▼───┐ ┌──▼───────┐
    │   │PostgreSQL MySQL SQLite │Datamapper│
    │   │       │  │      │  │      │ │  Bridge  │
    │   └───┬──┘  └──┬───┘  └──┬───┘ └──┬───────┘
    │       │        │         │        │
    └───────┼────────┼─────────┼────────┘
            │        │         │
         ┌──▼────────▼─────────▼──┐
         │   Database Layer        │
         │  (PostgreSQL/etc)       │
         └─────────────────────────┘
```

## Development Roadmap

| Phase | Timeline | Status | Deliverable |
|-------|----------|--------|-------------|
| **Phase 1** | Weeks 1-3 | 🔄 In Progress | Core engine + PostgreSQL |
| **Phase 2** | Weeks 4-5 | 📋 Planned | MySQL + SQLite + Datamapper |
| **Phase 3** | Weeks 6-7 | 📋 Planned | Seeder system |
| **Phase 4** | Weeks 8-9 | 📋 Planned | Advanced features + v0.1.0-alpha |

See [IMPLEMENTATION_PLAN.md](IMPLEMENTATION_PLAN.md) for detailed roadmap.

## Documentation

- [Implementation Plan](IMPLEMENTATION_PLAN.md) - Detailed development roadmap
- [Architecture Design](openspec/changes/create-sil-migrator/design.md) - System design decisions
- [Project Context](openspec/project.md) - Project background and conventions

### Coming Soon
- Migration Writing Guide
- PostgreSQL Adapter Guide
- CLI Command Reference
- Troubleshooting Guide
- Programmatic API Documentation

## Examples

Coming soon in Phase 1:
- Basic PostgreSQL migration example
- Programmatic API usage
- Docker deployment example

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

### Development Setup

```bash
git clone https://github.com/toutaio/toutago-sil-migrator.git
cd toutago-sil-migrator
go mod download
go test ./...
```

### Running Tests

```bash
# Run all tests
go test ./...

# Run with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run integration tests
go test -tags=integration ./tests/integration/...
```

### Code Quality

Please ensure:
- All tests pass
- Code coverage remains above 80% (target: 90%)
- Code is formatted with `gofmt`
- No linting errors from `golangci-lint`

## Project Information

- **Version**: 0.1.0-dev (Phase 1)
- **Started**: December 2024
- **Language**: Go 1.22+
- **Test Coverage**: Target 90%+
- **Status**: Under Active Development

## License

MIT License - see [LICENSE](LICENSE) file for details.

Copyright (c) 2024 Toutaio

## Acknowledgments

Inspired by the best practices from:
- **Rails ActiveRecord Migrations** - Sequential versioning, up/down migrations
- **Laravel Migrations** - Fluent schema builder, rollback capabilities
- **Flyway** - Version-based migrations with checksums
- **Alembic** - Branching and merging migration paths
- **Knex.js** - Transaction-wrapped migrations, seed management

---

Built with ❤️ for the Go community
