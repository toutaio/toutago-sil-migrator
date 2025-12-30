# Quick Reference

Common development commands for Síl migration tool.

## Build

```bash
# Build the CLI binary
go build -o bin/sil ./cmd/sil

# Install globally
go install ./cmd/sil

# Build for different platforms
GOOS=linux GOARCH=amd64 go build -o bin/sil-linux ./cmd/sil
GOOS=darwin GOARCH=amd64 go build -o bin/sil-darwin ./cmd/sil
GOOS=windows GOARCH=amd64 go build -o bin/sil.exe ./cmd/sil
```

## Test

```bash
# Run all tests
go test ./...

# Run with verbose output
go test -v ./...

# Run unit tests only (skip integration)
go test -short ./...

# Run with race detector
go test -race ./...

# Run with coverage
go test -coverprofile=coverage.out ./...

# View coverage report in browser
go tool cover -html=coverage.out

# Run specific test
go test -v -run TestMigrator_Migrate ./pkg/sil

# Run benchmarks
go test -bench=. -benchmem ./...
```

## Code Quality

```bash
# Format all code
go fmt ./...

# Check for issues
go vet ./...

# Run linter (requires golangci-lint)
golangci-lint run

# Install linter
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

## Integration Testing

```bash
# Start PostgreSQL with Docker
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

## Development Workflow

```bash
# 1. Make changes to code

# 2. Format
go fmt ./...

# 3. Check for issues
go vet ./...

# 4. Run tests
go test -race ./...

# 5. Build
go build -o bin/sil ./cmd/sil

# All in one command
go fmt ./... && go vet ./... && go test -race ./... && go build -o bin/sil ./cmd/sil
```

## Dependencies

```bash
# Download dependencies
go mod download

# Verify dependencies
go mod verify

# Update dependencies
go get -u ./...
go mod tidy

# View dependency graph
go mod graph

# Why is this dependency included?
go mod why github.com/lib/pq
```

## Module Management

```bash
# Initialize module (already done)
go mod init github.com/toutaio/toutago-sil-migrator

# Add dependency
go get github.com/some/package

# Remove unused dependencies
go mod tidy

# Vendor dependencies (optional)
go mod vendor
```

## Profiling

```bash
# CPU profiling
go test -cpuprofile=cpu.prof -bench=. ./...
go tool pprof cpu.prof

# Memory profiling
go test -memprofile=mem.prof -bench=. ./...
go tool pprof mem.prof

# Block profiling
go test -blockprofile=block.prof -bench=. ./...
go tool pprof block.prof
```

## Documentation

```bash
# View package documentation
go doc ./pkg/sil

# View specific function
go doc sil.NewMigrator

# Start local documentation server
godoc -http=:6060
# Then open http://localhost:6060/pkg/github.com/toutaio/toutago-sil-migrator/
```

## Clean Up

```bash
# Remove binary
rm -f bin/sil

# Remove coverage files
rm -f coverage.out

# Remove all build artifacts
rm -rf bin/ *.out *.prof
```

## CI/CD Simulation

Run the same checks as CI locally:

```bash
# Format check (fails if not formatted)
test -z "$(gofmt -s -l .)" || (echo "Please run: go fmt ./..." && exit 1)

# Vet
go vet ./...

# Tests with race detector and coverage
go test -v -race -coverprofile=coverage.out -covermode=atomic ./...

# Build
go build -v -o bin/sil ./cmd/sil

# All CI checks
go fmt ./... && \
  go vet ./... && \
  go test -race -coverprofile=coverage.out ./... && \
  go build -o bin/sil ./cmd/sil
```
