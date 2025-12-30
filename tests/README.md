# Testing Guide

This document describes how to test the Síl migration tool.

## Test Structure

The test suite is organized into three levels:

1. **Unit Tests** - Test individual components in isolation
2. **Integration Tests** - Test with a real PostgreSQL database
3. **CLI Tests** - Test the command-line interface

## Running Tests

### Quick Test (All Tests)

```bash
# Run all tests
go test -v ./...

# Run tests for specific package
go test -v ./pkg/sil

# Run with coverage
go test -cover ./...

# Run specific test
go test -v -run TestMigrator_Migrate ./pkg/sil

# Run with race detector
go test -race ./...
```

### Integration Tests

Integration tests require a PostgreSQL database. The easiest way is to use Docker:

```bash
# Start PostgreSQL for testing
docker run --name sil-test-db \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=sil_test \
  -p 5432:5432 \
  -d postgres:15

# Run all tests (including integration)
DATABASE_URL="postgres://postgres:postgres@localhost:5432/sil_test?sslmode=disable" \
  go test -v ./...

# Run only integration tests
DATABASE_URL="postgres://postgres:postgres@localhost:5432/sil_test?sslmode=disable" \
  go test -v -run Integration ./...

# Stop and remove container
docker stop sil-test-db && docker rm sil-test-db
```

### Unit Tests Only (Skip Integration)

```bash
go test -short ./...
```

## Test Coverage

### Generate Coverage Report

```bash
# Generate coverage profile
go test -coverprofile=coverage.out ./...

# View coverage in terminal
go tool cover -func=coverage.out

# View coverage in browser (opens in default browser)
go tool cover -html=coverage.out

# Generate coverage for specific package
go test -coverprofile=coverage.out ./pkg/sil
```

### Current Coverage

As of Day 5:
- **Unit Tests**: 43.5% coverage
- **Integration Tests**: Additional real-world scenarios
- **Target**: 80%+ for Phase 1 completion

## Test Categories

### 1. Unit Tests

#### Migration Tests (`migration_test.go`)
- Migration creation and validation
- Migration sorting
- Migration comparison
- Version validation

#### Migrator Tests (`migrator_test.go`)
- Migration execution
- Rollback functionality
- Status reporting
- Batch tracking
- Error handling

#### Config Tests (`config_test.go`)
- Configuration loading (YAML/JSON)
- Environment variable override
- Configuration validation
- Configuration merging
- File name generation

#### Loader Tests (`loader_test.go`)
- File discovery
- Migration loading
- Directory scanning
- Error handling

#### Generator Tests (`generator_test.go`)
- Template generation
- File creation
- Version generation

#### Logger Tests (`logger_test.go`)
- Log level filtering
- Color output
- Noop logger

### 2. Integration Tests (`integration_test.go`)

#### Database Adapter Tests
- Connection management
- Migrations table creation
- Migration recording
- Migration retrieval
- Batch tracking

#### Locking Tests
- Lock acquisition
- Lock release
- Concurrent lock attempts
- Lock timeout

#### Migration Lifecycle Tests
- Full migration execution
- Rollback functionality
- Transaction handling
- Table creation verification

#### Transaction Rollback Tests
- Automatic rollback on failure
- Data integrity verification
- Error propagation

## Writing New Tests

### Unit Test Template

```go
func TestMyFeature(t *testing.T) {
    // Arrange
    config := sil.DefaultConfig()
    
    // Act
    result, err := MyFunction(config)
    
    // Assert
    if err != nil {
        t.Fatalf("MyFunction() error = %v", err)
    }
    
    if result != expected {
        t.Errorf("MyFunction() = %v, want %v", result, expected)
    }
}
```

### Integration Test Template

```go
func TestIntegrationMyFeature(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test in short mode")
    }
    
    config := IntegrationTestConfig()
    
    if !isDatabaseAvailable(config.DatabaseURL) {
        t.Skip("PostgreSQL database not available")
    }
    
    cleanupDatabase(t, config)
    defer cleanupDatabase(t, config)
    
    // Test logic here...
}
```

### Table-Driven Tests

```go
func TestMyFeature(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    string
        wantErr bool
    }{
        {
            name:    "valid input",
            input:   "test",
            want:    "TEST",
            wantErr: false,
        },
        {
            name:    "invalid input",
            input:   "",
            want:    "",
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := MyFunction(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if got != tt.want {
                t.Errorf("got %v, want %v", got, tt.want)
            }
        })
    }
}
```

## Continuous Integration

### GitHub Actions Example

The project includes a CI workflow at `.github/workflows/ci.yml`:

```yaml
name: CI
on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    
    services:
      postgres:
        image: postgres:15
        env:
          POSTGRES_PASSWORD: postgres
          POSTGRES_DB: sil_test
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
        ports:
          - 5432:5432
    
    steps:
      - uses: actions/checkout@v4
      
      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.22'
      
      - name: Run Tests
        run: go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
        env:
          DATABASE_URL: postgres://postgres:postgres@localhost:5432/sil_test?sslmode=disable
      
      - name: Upload Coverage
        uses: codecov/codecov-action@v4
        with:
          files: ./coverage.out
```

### Local CI Simulation

Run the same checks locally before pushing:

```bash
# Format check
test -z "$(gofmt -s -l .)"

# Vet
go vet ./...

# Tests with race detector and coverage
go test -v -race -coverprofile=coverage.out -covermode=atomic ./...

# All in one
go fmt ./... && go vet ./... && go test -race -coverprofile=coverage.out ./...
```

## Mock Objects

The test suite includes several mock implementations:

### Mock Adapter (`migrator_test.go`)
```go
type mockAdapter struct {
    appliedMigrations []MigrationRecord
    lastBatch         int
    lockAcquired      bool
}
```

Used for:
- Testing migration logic without database
- Simulating various database states
- Testing error conditions

### Mock Transaction
```go
type mockTransaction struct {
    committed  bool
    rolledBack bool
}
```

Used for:
- Testing transaction behavior
- Verifying commit/rollback calls

### Mock Lock
```go
type mockLock struct {
    adapter  *mockAdapter
    released bool
}
```

Used for:
- Testing lock acquisition/release
- Simulating concurrent access

## Test Data

### Test Migrations

Test migrations are created dynamically:

```go
sil.RegisterMigration(createTestMigration(
    "20241230000001",
    "create_test_table",
    `CREATE TABLE test_users (id SERIAL PRIMARY KEY)`,
    `DROP TABLE IF EXISTS test_users`,
))
```

### Cleanup

Always cleanup test data:

```go
func cleanupDatabase(t *testing.T, config *sil.Config) {
    db, _ := sql.Open("postgres", config.DatabaseURL)
    defer db.Close()
    
    db.Exec("DROP TABLE IF EXISTS test_users CASCADE")
    db.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s CASCADE", config.TableName))
}
```

## Benchmarking

```bash
# Run all benchmarks
go test -bench=. -benchmem ./...

# Run benchmarks for specific package
go test -bench=. -benchmem ./pkg/sil

# Run specific benchmark
go test -bench=BenchmarkMigrator -benchmem ./pkg/sil

# Generate CPU profile
go test -bench=. -cpuprofile=cpu.prof ./pkg/sil
go tool pprof cpu.prof

# Generate memory profile
go test -bench=. -memprofile=mem.prof ./pkg/sil
go tool pprof mem.prof
```

## Best Practices

1. **Use t.TempDir()** - For temporary files/directories
2. **Use t.Cleanup()** - For cleanup functions
3. **Use t.Parallel()** - For parallel tests (when safe)
4. **Use testing.Short()** - Skip slow tests in short mode
5. **Use table-driven tests** - For testing multiple scenarios
6. **Test error paths** - Don't just test happy paths
7. **Use meaningful test names** - Describe what is being tested
8. **Keep tests independent** - Tests should not depend on each other
9. **Clean up after tests** - Don't leave artifacts
10. **Use subtests** - For better organization and filtering

## Troubleshooting

### Database Connection Fails

```bash
# Check if PostgreSQL is running
docker ps | grep postgres

# Check connection
psql -h localhost -U postgres -d sil_test

# View logs
docker logs sil-test-db
```

### Tests Hang

```bash
# Run with timeout
go test -timeout 30s ./...

# Run with race detector
go test -race ./...

# Verbose output with race detector
go test -v -race ./...
```

### Flaky Tests

```bash
# Run tests multiple times
go test -count=10 ./...

# Run with shuffle (randomize test order)
go test -shuffle=on ./...

# Combine shuffle and multiple runs
go test -shuffle=on -count=5 ./...
```

## Resources

- [Go Testing Package](https://pkg.go.dev/testing)
- [Testing Best Practices](https://go.dev/doc/effective_go#testing)
- [Table-Driven Tests](https://dave.cheney.net/2013/06/09/writing-table-driven-tests-in-go)
- [Test Coverage](https://go.dev/blog/cover)
