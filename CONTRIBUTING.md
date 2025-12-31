# Contributing to Síl Migrator

Thank you for your interest in contributing to Síl! This document provides guidelines and information for contributors.

## Code of Conduct

This project adheres to a code of conduct. By participating, you are expected to uphold this code. Please report unacceptable behavior to the project maintainers.

## How to Contribute

### Reporting Issues

- Use the GitHub issue tracker
- Check if the issue already exists
- Provide detailed information:
  - Go version
  - Database type and version
  - Operating system
  - Steps to reproduce
  - Expected vs actual behavior

### Pull Requests

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Write or update tests
5. Ensure tests pass (`go test ./...`)
6. Ensure code is formatted (`go fmt ./...`)
7. Run linter (`golangci-lint run`)
8. Commit with conventional commit format
9. Push to your fork
10. Open a Pull Request

### Commit Convention

We use [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>(<scope>): <subject>
```

**Types:**
- `feat`: New feature
- `fix`: Bug fix
- `perf`: Performance improvement
- `refactor`: Code restructuring
- `test`: Test additions or modifications
- `docs`: Documentation changes
- `chore`: Build, CI, or tooling changes

**Examples:**
```
feat(cli): add rollback command with batch support
fix(lock): prevent race condition in distributed lock
perf(migrator): optimize migration loading
docs(readme): add multi-database examples
test(adapter): add PostgreSQL adapter tests
```

## Development Setup

### Prerequisites

- Go 1.22 or higher
- PostgreSQL 10+ / MySQL 5.7+ / SQLite 3.25+ (for testing)
- Git
- golangci-lint (for linting)

### Getting Started

```bash
# Clone your fork
git clone https://github.com/YOUR_USERNAME/toutago-sil-migrator
cd toutago-sil-migrator

# Install dependencies
go mod download

# Build the CLI
go build -o sil ./cmd/sil

# Run tests
go test ./...

# Run with PostgreSQL test database
export DATABASE_URL="postgres://user:password@localhost:5432/testdb?sslmode=disable"
./sil migrate
```

## Testing Requirements

- All new code must include tests
- Target minimum 80% code coverage (currently 73.9%)
- Tests must pass with race detector: `go test -race ./...`
- Test against PostgreSQL, MySQL, and SQLite
- Integration tests for database adapters

## Code Quality Standards

- Follow Go best practices and idioms
- Use meaningful variable and function names
- Keep functions focused and small
- Document exported types and functions
- Pass golangci-lint without errors
- Ensure CLI output is user-friendly with colored formatting

## Documentation

- Update README.md for user-facing changes
- Update doc.go for API changes
- Add examples for new features
- Keep CHANGELOG.md current
- Update QUICKREF.md for CLI changes

## Architecture

Síl follows SOLID principles:
- Each adapter implements the `DatabaseAdapter` interface
- Migration engine is independent of specific databases
- Locking mechanisms are database-specific but follow common interface

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
