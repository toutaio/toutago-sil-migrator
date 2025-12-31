package adapters

import (
	"context"
	"database/sql"
	"fmt"
	"hash/fnv"
	"net/url"
	"time"

	_ "github.com/lib/pq" // PostgreSQL driver
	"github.com/toutaio/toutago-sil-migrator/pkg/sil"
)

// PostgresAdapter implements DatabaseAdapter for PostgreSQL.
type PostgresAdapter struct {
	db     *sql.DB
	config *sil.Config
	logger sil.Logger
}

// NewPostgresAdapter creates a new PostgreSQL adapter.
func NewPostgresAdapter(config *sil.Config) (*PostgresAdapter, error) {
	if config == nil {
		return nil, fmt.Errorf("config is nil")
	}

	logger := sil.NewDefaultLogger(config.Verbose)
	if config.Verbose {
		logger = sil.NewColorLogger(config.Verbose)
	}

	return &PostgresAdapter{
		config: config,
		logger: logger,
	}, nil
}

// Connect establishes a connection to PostgreSQL.
func (a *PostgresAdapter) Connect(ctx context.Context, config *sil.Config) error {
	a.logger.Debug("Connecting to PostgreSQL database")

	db, err := sql.Open("postgres", config.DatabaseURL)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(config.MaxConnections)
	db.SetMaxIdleConns(config.MaxIdleConnections)
	db.SetConnMaxLifetime(config.ConnectionMaxLifetime)

	// Test connection
	pingCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := db.PingContext(pingCtx); err != nil {
		_ = db.Close()
		return fmt.Errorf("failed to ping database: %w", err)
	}

	a.db = db
	a.logger.Info("Connected to PostgreSQL database")

	return nil
}

// Close closes the database connection.
func (a *PostgresAdapter) Close() error {
	if a.db == nil {
		return nil
	}

	a.logger.Debug("Closing database connection")

	if err := a.db.Close(); err != nil {
		return fmt.Errorf("failed to close database: %w", err)
	}

	a.db = nil
	return nil
}

// Exec executes a SQL statement.
func (a *PostgresAdapter) Exec(ctx context.Context, query string, args ...interface{}) error {
	if a.db == nil {
		return fmt.Errorf("database not connected")
	}

	a.logger.Debug("Executing query")

	_, err := a.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("exec failed: %w", err)
	}

	return nil
}

// Query executes a SQL query.
func (a *PostgresAdapter) Query(ctx context.Context, query string, args ...interface{}) (sil.Rows, error) {
	if a.db == nil {
		return nil, fmt.Errorf("database not connected")
	}

	a.logger.Debug("Executing query")

	rows, err := a.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}

	return sil.NewRowsAdapter(rows), nil
}

// BeginTx starts a new transaction.
func (a *PostgresAdapter) BeginTx(ctx context.Context) (sil.Transaction, error) {
	if a.db == nil {
		return nil, fmt.Errorf("database not connected")
	}

	a.logger.Debug("Beginning transaction")

	tx, err := a.db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	return newTransaction(tx, a.logger), nil
}

// CreateMigrationsTable creates the migrations tracking table.
func (a *PostgresAdapter) CreateMigrationsTable(ctx context.Context) error {
	a.logger.Debug("Creating migrations table if not exists")

	//#nosec G201 -- Table name is from config file, not user input
	query := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			id SERIAL PRIMARY KEY,
			version VARCHAR(255) UNIQUE NOT NULL,
			description TEXT,
			batch INTEGER NOT NULL,
			executed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`, a.config.TableName)

	return a.Exec(ctx, query)
}

// GetAppliedMigrations returns all applied migrations.
func (a *PostgresAdapter) GetAppliedMigrations(ctx context.Context) ([]sil.MigrationRecord, error) {
	a.logger.Debug("Fetching applied migrations")

	//#nosec G201 -- Table name is from config file, not user input
	query := fmt.Sprintf(`
		SELECT id, version, description, batch, executed_at
		FROM %s
		ORDER BY executed_at ASC
	`, a.config.TableName)

	rows, err := a.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var records []sil.MigrationRecord
	for rows.Next() {
		var record sil.MigrationRecord
		if err := rows.Scan(&record.ID, &record.Version, &record.Description, &record.Batch, &record.ExecutedAt); err != nil {
			return nil, fmt.Errorf("failed to scan migration record: %w", err)
		}
		records = append(records, record)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return records, nil
}

// RecordMigration records a migration as applied.
func (a *PostgresAdapter) RecordMigration(ctx context.Context, version, description string, batch int) error {
	a.logger.Debug("Recording migration: %s", version)

	//#nosec G201 -- Table name is from config file, not user input
	query := fmt.Sprintf(`
		INSERT INTO %s (version, description, batch)
		VALUES ($1, $2, $3)
	`, a.config.TableName)

	return a.Exec(ctx, query, version, description, batch)
}

// RemoveMigration removes a migration record.
func (a *PostgresAdapter) RemoveMigration(ctx context.Context, version string) error {
	a.logger.Debug("Removing migration record: %s", version)

	//#nosec G201 -- Table name is from config file, not user input
	query := fmt.Sprintf(`
		DELETE FROM %s WHERE version = $1
	`, a.config.TableName)

	return a.Exec(ctx, query, version)
}

// Lock acquires a migration lock using PostgreSQL advisory locks.
func (a *PostgresAdapter) Lock(ctx context.Context) (sil.Lock, error) {
	if a.db == nil {
		return nil, fmt.Errorf("database not connected")
	}

	// Generate a consistent lock key from database name
	lockKey := a.getLockKey()

	a.logger.Debug("Acquiring advisory lock with key: %d", lockKey)

	// Try to acquire lock with timeout
	deadline := time.Now().Add(a.config.LockTimeout)
	for {
		// Try to acquire lock
		var locked bool
		err := a.db.QueryRowContext(ctx, "SELECT pg_try_advisory_lock($1)", lockKey).Scan(&locked)
		if err != nil {
			return nil, fmt.Errorf("failed to acquire lock: %w", err)
		}

		if locked {
			a.logger.Debug("Advisory lock acquired")

			// Return lock with release function
			releaseFunc := func() error {
				a.logger.Debug("Releasing advisory lock")
				_, err := a.db.ExecContext(context.Background(), "SELECT pg_advisory_unlock($1)", lockKey)
				return err
			}

			return newLock(releaseFunc, a.logger), nil
		}

		// Check timeout
		if time.Now().After(deadline) {
			return nil, fmt.Errorf("%w: timeout after %v", sil.ErrLockTimeout, a.config.LockTimeout)
		}

		// Check context
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("lock acquisition cancelled: %w", ctx.Err())
		case <-time.After(1 * time.Second):
			// Retry after 1 second
		}
	}
}

// GetLastBatch returns the last batch number.
func (a *PostgresAdapter) GetLastBatch(ctx context.Context) (int, error) {
	a.logger.Debug("Getting last batch number")

	//#nosec G201 -- Table name is from config file, not user input
	query := fmt.Sprintf(`
		SELECT COALESCE(MAX(batch), 0) FROM %s
	`, a.config.TableName)

	var batch int
	err := a.db.QueryRowContext(ctx, query).Scan(&batch)
	if err != nil {
		return 0, fmt.Errorf("failed to get last batch: %w", err)
	}

	return batch, nil
}

// getLockKey generates a consistent lock key from the database URL.
func (a *PostgresAdapter) getLockKey() int64 {
	// Parse database URL to get database name
	dbName := "sil_migrations"

	if u, err := url.Parse(a.config.DatabaseURL); err == nil {
		if len(u.Path) > 1 {
			dbName = u.Path[1:] // Remove leading slash
		}
	}

	// Generate hash
	h := fnv.New64a()
	_, _ = h.Write([]byte("sil_migration_lock_" + dbName)) // hash.Hash.Write never returns error

	//#nosec G115 -- Conversion is intentional for database advisory lock ID
	return int64(h.Sum64() & 0x7FFFFFFFFFFFFFFF) // Ensure positive
}

// transaction implements sil.Transaction for PostgreSQL.
type transaction struct {
	tx     *sql.Tx
	logger sil.Logger
}

// newTransaction creates a new PostgreSQL transaction.
func newTransaction(tx *sql.Tx, logger sil.Logger) *transaction {
	return &transaction{
		tx:     tx,
		logger: logger,
	}
}

// Commit commits the transaction.
func (t *transaction) Commit() error {
	t.logger.Debug("Committing transaction")
	return t.tx.Commit()
}

// Rollback rolls back the transaction.
func (t *transaction) Rollback() error {
	t.logger.Debug("Rolling back transaction")
	return t.tx.Rollback()
}

// Exec executes a statement within the transaction.
func (t *transaction) Exec(ctx context.Context, query string, args ...interface{}) error {
	_, err := t.tx.ExecContext(ctx, query, args...)
	return err
}

// Query executes a query within the transaction.
func (t *transaction) Query(ctx context.Context, query string, args ...interface{}) (sil.Rows, error) {
	rows, err := t.tx.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return sil.NewRowsAdapter(rows), nil
}

// lock implements sil.Lock for PostgreSQL.
type lock struct {
	releaseFunc func() error
	locked      bool
	logger      sil.Logger
}

// newLock creates a new PostgreSQL lock.
func newLock(releaseFunc func() error, logger sil.Logger) *lock {
	return &lock{
		releaseFunc: releaseFunc,
		locked:      true,
		logger:      logger,
	}
}

// Release releases the lock.
func (l *lock) Release() error {
	if !l.locked {
		return nil
	}

	if err := l.releaseFunc(); err != nil {
		return err
	}

	l.locked = false
	return nil
}

// IsLocked returns whether the lock is held.
func (l *lock) IsLocked() bool {
	return l.locked
}
