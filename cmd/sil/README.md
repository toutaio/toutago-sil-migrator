# Síl CLI

Command-line interface for the Síl migration and seeding tool.

## Installation

### From Source

```bash
go install github.com/toutaio/toutago-sil-migrator/cmd/sil@latest
```

### Build Locally

```bash
git clone https://github.com/toutaio/toutago-sil-migrator
cd toutago-sil-migrator
go build -o bin/sil ./cmd/sil
```

## Quick Start

### 1. Initialize a New Project

```bash
sil init
```

This creates:
- `migrations/` directory
- `seeders/` directory
- `sil.yaml` configuration file

### 2. Configure Database

Edit `sil.yaml` and set your database URL:

```yaml
database_url: postgres://user:pass@localhost:5432/mydb?sslmode=disable
migrations_dir: ./migrations
seeders_dir: ./seeders
table_name: sil_migrations
seeds_table_name: sil_seeds
environment: development
```

Or use environment variable:

```bash
export DATABASE_URL="postgres://user:pass@localhost:5432/mydb?sslmode=disable"
```

### 3. Create Your First Migration

```bash
sil create create_users_table
```

Or use templates:

```bash
# Create table template
sil create --table users create_users_table

# Add column template
sil create --table users --column email add_email_to_users
```

### 4. Edit the Migration

Open the generated file in `migrations/` and add your SQL:

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

### 5. Run Migrations

```bash
# Run all pending migrations
sil migrate

# Run specific number of migrations
sil migrate --steps 3
```

### 6. Check Status

```bash
sil status
```

Output:
```
VERSION          DESCRIPTION           STATUS      BATCH  EXECUTED AT
-------          -----------           ------      -----  -----------
20241230100000   create users table    ✅ Applied  1      2024-12-30 10:30:45

Total: 1 migrations (1 applied, 0 pending)
```

## Commands

### `sil init`

Initialize a new Síl project.

```bash
sil init                    # Create sil.yaml
sil init --config custom.yaml  # Custom config name
sil init --force            # Overwrite existing files
```

### `sil create`

Create a new migration file.

```bash
sil create <name>                                    # Basic migration
sil create --table <table> <name>                    # Create table template
sil create --table <table> --column <col> <name>     # Add column template
```

Examples:
```bash
sil create create_users_table
sil create add_status_to_orders
sil create --table products create_products_table
sil create --table users --column phone add_phone_column
```

### `sil migrate`

Run pending migrations.

```bash
sil migrate                 # Run all pending
sil migrate --steps 5       # Run next 5 migrations
sil migrate -v              # Verbose output
```

### `sil rollback`

Rollback migrations.

```bash
sil rollback                # Rollback last batch
sil rollback --steps 3      # Rollback last 3 migrations
```

### `sil status`

Show migration status.

```bash
sil status                  # Table format
sil status --json           # JSON format (coming soon)
```

### `sil reset`

Rollback all migrations.

```bash
sil reset                   # With confirmation
sil reset --force           # Skip confirmation
```

⚠️ **Warning**: This will rollback ALL migrations!

### `sil refresh`

Rollback all migrations and re-run them.

```bash
sil refresh                 # With confirmation
sil refresh --force         # Skip confirmation
```

⚠️ **Warning**: This will rollback and re-run ALL migrations!

## Global Flags

- `-c, --config <file>` - Specify config file (default: sil.yaml)
- `-v, --verbose` - Enable verbose output
- `--version` - Show version information

## Configuration

### File-based Configuration

Create `sil.yaml`:

```yaml
database_url: postgres://localhost/mydb
migrations_dir: ./migrations
seeders_dir: ./seeders
table_name: sil_migrations
seeds_table_name: sil_seeds
environment: development
lock_timeout: 5m
migration_timeout: 1h
max_connections: 10
max_idle_connections: 2
connection_max_lifetime: 1h
verbose: false
```

### Environment Variables

All configuration can be overridden with environment variables:

- `SIL_DATABASE_URL`
- `SIL_MIGRATIONS_DIR`
- `SIL_SEEDERS_DIR`
- `SIL_TABLE_NAME`
- `SIL_SEEDS_TABLE_NAME`
- `SIL_ENVIRONMENT`
- `SIL_LOCK_TIMEOUT`
- `SIL_MIGRATION_TIMEOUT`
- `SIL_MAX_CONNECTIONS`
- `SIL_MAX_IDLE_CONNECTIONS`
- `SIL_CONNECTION_MAX_LIFETIME`
- `SIL_VERBOSE`

Example:
```bash
export SIL_DATABASE_URL="postgres://localhost/mydb"
export SIL_VERBOSE="true"
sil migrate
```

## CI/CD Integration

### GitHub Actions

```yaml
name: Migrations
on: [push]

jobs:
  migrate:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:15
        env:
          POSTGRES_PASSWORD: postgres
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.22'
      
      - name: Install Síl
        run: go install github.com/toutaio/toutago-sil-migrator/cmd/sil@latest
      
      - name: Run migrations
        env:
          DATABASE_URL: postgres://postgres:postgres@localhost:5432/test?sslmode=disable
        run: sil migrate
```

### Docker

```dockerfile
FROM golang:1.22-alpine

WORKDIR /app

# Install Síl
RUN go install github.com/toutaio/toutago-sil-migrator/cmd/sil@latest

# Copy migrations
COPY migrations ./migrations
COPY sil.yaml ./

# Run migrations on container start
CMD ["sil", "migrate"]
```

## Examples

See the `examples/` directory for complete working examples:

- `examples/basic/` - Basic migration setup
- `examples/multi-db/` - Multiple database support (coming soon)
- `examples/with-datamapper/` - Integration with toutago-datamapper (coming soon)

## Troubleshooting

### Lock Timeout

If you get a lock timeout error:

```
Error: lock acquisition failed: timeout after 5m
```

This means another migration is running. Wait for it to complete or:

1. Check for stuck locks in PostgreSQL:
   ```sql
   SELECT * FROM pg_locks WHERE locktype = 'advisory';
   ```

2. Manually release the lock (use with caution):
   ```sql
   SELECT pg_advisory_unlock_all();
   ```

### Migration Failed

If a migration fails:

1. Check the error message in the output
2. The migration is rolled back automatically
3. Fix the migration and run again
4. The failed migration is not recorded

### No Migrations Found

If you see "No migrations found":

1. Ensure you've imported your migrations package in your main.go
2. Check that migrations are registered in `init()`
3. Verify the migrations directory path in config

## Best Practices

1. **Always write Down() methods** - Enable rollbacks
2. **Test migrations locally first** - Before running in production
3. **Use transactions** - Migrations are automatically wrapped in transactions
4. **Keep migrations small** - One change per migration
5. **Never modify applied migrations** - Create new ones instead
6. **Use descriptive names** - Make migration purpose clear
7. **Review generated SQL** - Especially for production

## Support

- GitHub Issues: https://github.com/toutaio/toutago-sil-migrator/issues
- Documentation: https://github.com/toutaio/toutago-sil-migrator

## License

MIT License - see LICENSE file for details
