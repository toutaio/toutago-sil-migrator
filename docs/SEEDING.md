# Seeding Guide

Complete guide to using the Síl seeding system.

## Quick Start

```bash
# Create a seeder
sil seed create users

# Run all seeders  
sil seed --all

# Check status
sil seed status
```

## Creating Seeders

### Using CLI

```bash
sil seed create users
```

Creates `seeders/TIMESTAMP_users.go`:

```go
func init() {
    sil.RegisterSeeder(NewUsersSeeder())
}

func NewUsersSeeder() sil.Seeder {
    return sil.NewBaseSeeder("users", seedUsers).
        SetDependencies().
        SetEnvironments("development", "test")
}

func seedUsers(ctx context.Context, adapter sil.DatabaseAdapter) error {
    // Your seeding logic
    return nil
}
```

## Dependencies

Declare dependencies for correct execution order:

```go
sil.NewBaseSeeder("posts", seedPosts).
    SetDependencies("users", "categories")
```

Síl automatically:
- Sorts seeders topologically
- Detects circular dependencies  
- Validates dependencies exist

## Environment Filtering

```go
// Development and test only
seeder.SetEnvironments("development", "test")

// All environments (default)
seeder // No SetEnvironments()
```

## Idempotency

Prevent duplicate data:

```go
seeder.SetShouldRun(func(ctx context.Context, adapter sil.DatabaseAdapter) (bool, error) {
    rows, _ := adapter.Query(ctx, "SELECT COUNT(*) FROM users")
    defer rows.Close()
    var count int
    rows.Next()
    rows.Scan(&count)
    return count == 0, nil
})
```

Or use `ON CONFLICT`:

```go
INSERT INTO roles (name) VALUES ('admin')
ON CONFLICT (name) DO NOTHING
```

## CLI Commands

### Run Seeders

```bash
sil seed --all                    # Run all
sil seed users roles              # Run specific
sil seed --force --all            # Re-run executed
sil seed --dry-run --all          # Preview
sil seed --env production --all   # Specific environment
```

### Create Seeder

```bash
sil seed create users
```

### Check Status

```bash
sil seed status
```

## Examples

See `examples/seeders/main.go` for complete examples with dependencies.

