# Síl - Database Migration and Seeding Tool

[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go)](https://go.dev)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Coverage](https://img.shields.io/badge/Coverage-71.0%25-yellowgreen.svg)]()

Síl is a powerful, standalone database migration and seeding tool for Go applications, inspired by Rails, Laravel, Flyway, and Alembic.

## ✨ Features

- 🔒 **Distributed Locking** - Prevents concurrent migrations with PostgreSQL advisory locks
- 💾 **Transaction Safety** - Automatic rollback on migration failures
- 📦 **Batch Tracking** - Precise rollback control by migration batches
- 🎯 **Type-Safe Migrations** - Write migrations in Go
- 🔌 **Standalone** - Zero required dependencies (optional datamapper integration)
- 🗄️ **Multi-Database** - PostgreSQL (MySQL, SQLite coming in Phase 2)
- 🎨 **Beautiful CLI** - Colored output with progress tracking
- 📝 **Comprehensive Logging** - Track every migration step
- ⚡ **Fast** - Optimized for performance
- 🧪 **Well-Tested** - 71% test coverage (110+ unit tests + integration tests)

## 🚀 Quick Start

### Installation

```bash
go install github.com/toutaio/toutago-sil-migrator/cmd/sil@latest
```

Or build from source:

```bash
git clone https://github.com/toutaio/toutago-sil-migrator
cd toutago-sil-migrator
make install
```

### Initialize Project

```bash
# Create project structure
sil init

# Configure database
export DATABASE_URL="postgres://user:pass@localhost:5432/mydb?sslmode=disable"
```

### Create Migration

```bash
# Basic migration
sil create create_users_table

# With table template
sil create --table users create_users_table
```

Edit the generated migration:

```go
func (m *Migration_20241230100000_CreateUsersTable) Up(adapter sil.DatabaseAdapter) error {
    ctx := context.Background()
    return adapter.Exec(ctx, `
        CREATE TABLE users (
            id SERIAL PRIMARY KEY,
            name VARCHAR(255) NOT NULL,
            email VARCHAR(255) UNIQUE NOT NULL,
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        )
    `)
}

func (m *Migration_20241230100000_CreateUsersTable) Down(adapter sil.DatabaseAdapter) error {
    ctx := context.Background()
    return adapter.Exec(ctx, `DROP TABLE IF EXISTS users`)
}
```

### Run Migrations

```bash
# Run all pending migrations
sil migrate

# Run specific number of migrations
sil migrate --steps 3

# Check status
sil status

# Rollback last batch
sil rollback

# Rollback specific number
sil rollback --steps 2
```

## 📖 Documentation

- [CLI Reference](cmd/sil/README.md) - Complete CLI documentation
- [Testing Guide](tests/README.md) - How to test
- [Implementation Plan](IMPLEMENTATION_PLAN.md) - Development roadmap
- [Phase 1 Progress](PHASE1_PROGRESS.md) - Current progress

## 🏗️ Architecture

```
┌─────────────────────────────────────────────────────────┐
│                     CLI (Cobra)                         │
│  init  create  migrate  rollback  status  reset         │
└────────────────────┬────────────────────────────────────┘
                     │
┌────────────────────┴────────────────────────────────────┐
│                  Migrator Core                          │
│  • Migration Registry                                   │
│  • Batch Tracking                                       │
│  • Execution Engine                                     │
└────────────────────┬────────────────────────────────────┘
                     │
┌────────────────────┴────────────────────────────────────┐
│              Database Adapters                          │
│  PostgreSQL │ MySQL │ SQLite │ Custom                   │
└─────────────────────────────────────────────────────────┘
```

## 💡 Examples

### Basic Example

See [`examples/basic/`](examples/basic/) for a complete working example.

### Programmatic Usage

```go
package main

import (
    "context"
    "github.com/toutaio/toutago-sil-migrator/pkg/sil"
    "github.com/toutaio/toutago-sil-migrator/pkg/sil/adapters"
)

func main() {
    // Configure
    config := sil.DefaultConfig()
    config.DatabaseURL = "postgres://localhost/mydb"
    
    // Create adapter
    adapter, _ := adapters.NewPostgresAdapter(config)
    adapter.Connect(context.Background(), config)
    defer adapter.Close()
    
    // Create migrator
    migrator, _ := sil.NewMigrator(config, adapter)
    
    // Run migrations
    migrator.Migrate(context.Background())
}
```

## 🛠️ Development

### Prerequisites

- Go 1.22+
- PostgreSQL 12+ (for integration testing)

### Setup

```bash
# Clone repository
git clone https://github.com/toutaio/toutago-sil-migrator
cd toutago-sil-migrator

# Download dependencies
go mod download

# Verify dependencies
go mod verify

# Run tests
go test ./...

# Build
go build -o bin/sil ./cmd/sil
```

### Testing

```bash
# Run all tests
go test -v ./...

# Run with race detector
go test -race ./...

# Run unit tests only (skip integration)
go test -short ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...

# View coverage in terminal
go tool cover -func=coverage.out

# View coverage in browser
go tool cover -html=coverage.out

# Run specific test
go test -v -run TestMigrator_Migrate ./pkg/sil

# Run benchmarks
go test -bench=. -benchmem ./...
```

#### Integration Testing with Docker

```bash
# Start PostgreSQL
docker run --name sil-test-db \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=sil_test \
  -p 5432:5432 \
  -d postgres:15

# Run tests with database
DATABASE_URL="postgres://postgres:postgres@localhost:5432/sil_test?sslmode=disable" \
  go test -v ./...

# Cleanup
docker stop sil-test-db && docker rm sil-test-db
```

See [Testing Guide](tests/README.md) for detailed information.

## 🗺️ Roadmap

### Phase 1: Foundation (Weeks 1-3) ✅ 20% Complete
- [x] Core migration engine
- [x] PostgreSQL adapter
- [x] CLI with 7 commands
- [x] Configuration management
- [x] Testing framework
- [ ] 80%+ test coverage

### Phase 2: Multi-Database (Weeks 4-5)
- [ ] MySQL adapter
- [ ] SQLite adapter
- [ ] Optional datamapper integration
- [ ] Enhanced migration templates

### Phase 3: Seeding System (Weeks 6-7)
- [ ] Seeder framework
- [ ] Dependency management
- [ ] Environment-specific seeds
- [ ] Seeder CLI commands

### Phase 4: Advanced Features (Weeks 8-9)
- [ ] Migration versioning strategies
- [ ] Dry-run mode
- [ ] Migration validation
- [ ] Performance optimization
- [ ] Production hardening

## 🤝 Contributing

Contributions are welcome! Please read our contributing guidelines.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Development Workflow

```bash
# Format code
go fmt ./...

# Run linter (install first: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
golangci-lint run

# Run vet
go vet ./...

# Run tests
go test ./...

# Run tests with race detector and coverage
go test -race -coverprofile=coverage.out ./...
```

## 📊 Project Status

**Phase 1 Progress**: 20% (Day 4 of 15)

**Current Milestone**: Testing & Documentation

**Next Up**: Integration testing and coverage improvement

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

Inspired by:
- [Rails Migrations](https://guides.rubyonrails.org/active_record_migrations.html)
- [Laravel Migrations](https://laravel.com/docs/migrations)
- [Flyway](https://flywaydb.org/)
- [Alembic](https://alembic.sqlalchemy.org/)
- [golang-migrate](https://github.com/golang-migrate/migrate)

## 📧 Contact

- GitHub: [@toutaio](https://github.com/toutaio)
- Issues: [GitHub Issues](https://github.com/toutaio/toutago-sil-migrator/issues)

---

**Made with ❤️ by the Toutago team**
