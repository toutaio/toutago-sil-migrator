# Basic Síl Migration Example

This example demonstrates basic usage of Síl for database migrations.

## Prerequisites

- Go 1.22+
- PostgreSQL database

## Setup

1. Start a PostgreSQL database:
```bash
docker run --name sil-postgres -e POSTGRES_PASSWORD=postgres -p 5432:5432 -d postgres:15
```

2. Set the database URL (optional):
```bash
export DATABASE_URL="postgres://postgres:postgres@localhost:5432/sil_example?sslmode=disable"
```

## Running the Example

```bash
# From the project root
go run ./examples/basic/main.go
```

This will:
1. Connect to the PostgreSQL database
2. Run all pending migrations
3. Display migration status

## What Happens

The example includes one migration that creates a `users` table with:
- `id` (SERIAL PRIMARY KEY)
- `name` (VARCHAR NOT NULL)
- `email` (VARCHAR UNIQUE NOT NULL)
- `created_at` (TIMESTAMP)
- `updated_at` (TIMESTAMP)

## Expected Output

```
Running migrations...
[INFO] Starting migration...
[INFO] Connected to PostgreSQL database
[INFO] Found 1 pending migrations
[INFO] [1/1] Migrating: 20241230100000 - create users table
[INFO] ✓ Migrated: 20241230100000
[INFO] Migration complete! Ran 1 migrations

Migration status:
  20241230100000 - create users table - ✅ Applied (batch 1)

✅ Migration complete!
```

## Database Tables Created

After running, you'll have two tables:
1. `users` - Your application table
2. `sil_migrations` - Migration tracking table

## Rollback

To rollback the migration:

```go
// Add to main.go before migrator.Migrate()
if err := migrator.Rollback(ctx); err != nil {
    log.Fatalf("Rollback failed: %v", err)
}
```

## Next Steps

1. Add more migrations to the `migrations/` directory
2. Implement the Down() method for rollback support
3. Try `MigrateUp(ctx, N)` to run N migrations
4. Try `Reset(ctx)` to rollback all migrations
