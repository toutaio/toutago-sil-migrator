package adapters

import (
	"context"
	"database/sql"
	"fmt"
	"hash/fnv"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/toutaio/toutago-sil-migrator/pkg/sil"
)

// MySQLAdapter implements DatabaseAdapter for MySQL databases.
type MySQLAdapter struct {
	db     *sql.DB
	config *sil.Config
	logger sil.Logger
}

// NewMySQLAdapter creates a new MySQL database adapter.
func NewMySQLAdapter(config *sil.Config) (*MySQLAdapter, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	if config.DatabaseURL == "" {
		return nil, fmt.Errorf("database URL is required")
	}

	return &MySQLAdapter{
		config: config,
		logger: sil.NewDefaultLogger(config.Verbose),
	}, nil
}

// Connect establishes a connection to the MySQL database.
func (a *MySQLAdapter) Connect(ctx context.Context, config *sil.Config) error {
	a.logger.Debug("Connecting to MySQL database")

	db, err := sql.Open("mysql", config.DatabaseURL)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(config.MaxConnections)
	db.SetMaxIdleConns(config.MaxIdleConnections)
	db.SetConnMaxLifetime(config.ConnectionMaxLifetime)

	// Test connection
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return fmt.Errorf("failed to ping database: %w", err)
	}

	a.db = db
	a.logger.Info("Connected to MySQL database")
	return nil
}

// Close closes the database connection.
func (a *MySQLAdapter) Close() error {
	a.logger.Debug("Closing database connection")
	if a.db != nil {
		return a.db.Close()
	}
	return nil
}

// Exec executes a SQL statement.
func (a *MySQLAdapter) Exec(ctx context.Context, query string, args ...interface{}) error {
	a.logger.Debug("Executing query")
	_, err := a.db.ExecContext(ctx, query, args...)
	return err
}

// Query executes a SQL query and returns rows.
func (a *MySQLAdapter) Query(ctx context.Context, query string, args ...interface{}) (sil.Rows, error) {
	a.logger.Debug("Executing query")
	rows, err := a.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

// BeginTx begins a new transaction.
func (a *MySQLAdapter) BeginTx(ctx context.Context) (sil.Transaction, error) {
	a.logger.Debug("Beginning transaction")
	tx, err := a.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	return &mysqlTransaction{
		tx:     tx,
		logger: a.logger,
	}, nil
}

// CreateMigrationsTable creates the migrations tracking table.
func (a *MySQLAdapter) CreateMigrationsTable(ctx context.Context) error {
	a.logger.Debug("Creating migrations table if not exists")

	query := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			version VARCHAR(255) NOT NULL UNIQUE,
			description TEXT,
			batch INT NOT NULL,
			executed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			INDEX idx_batch (batch)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
	`, a.config.TableName)

	return a.Exec(ctx, query)
}

// GetAppliedMigrations returns all applied migrations.
func (a *MySQLAdapter) GetAppliedMigrations(ctx context.Context) ([]sil.MigrationRecord, error) {
	a.logger.Debug("Fetching applied migrations")

	query := fmt.Sprintf(`
		SELECT id, version, description, batch, executed_at
		FROM %s
		ORDER BY id ASC
	`, a.config.TableName)

	rows, err := a.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []sil.MigrationRecord
	for rows.Next() {
		var record sil.MigrationRecord
		if err := rows.Scan(&record.ID, &record.Version, &record.Description, &record.Batch, &record.ExecutedAt); err != nil {
			return nil, err
		}
		records = append(records, record)
	}

	return records, rows.Err()
}

// RecordMigration records a migration as applied.
func (a *MySQLAdapter) RecordMigration(ctx context.Context, version, description string, batch int) error {
	a.logger.Debug("Recording migration: %s", version)

	query := fmt.Sprintf(`
		INSERT INTO %s (version, description, batch)
		VALUES (?, ?, ?)
	`, a.config.TableName)

	return a.Exec(ctx, query, version, description, batch)
}

// RemoveMigration removes a migration record.
func (a *MySQLAdapter) RemoveMigration(ctx context.Context, version string) error {
	a.logger.Debug("Removing migration record: %s", version)

	query := fmt.Sprintf(`
		DELETE FROM %s WHERE version = ?
	`, a.config.TableName)

	return a.Exec(ctx, query, version)
}

// GetLastBatchNumber returns the last batch number.
func (a *MySQLAdapter) GetLastBatch(ctx context.Context) (int, error) {
	a.logger.Debug("Getting last batch number")

	query := fmt.Sprintf(`
		SELECT COALESCE(MAX(batch), 0) FROM %s
	`, a.config.TableName)

	var batch int
	row := a.db.QueryRowContext(ctx, query)
	if err := row.Scan(&batch); err != nil {
		return 0, err
	}

	return batch, nil
}

// Lock acquires a database lock for migrations.
// MySQL uses GET_LOCK() function for named locks.
func (a *MySQLAdapter) Lock(ctx context.Context) (sil.Lock, error) {
	a.logger.Debug("Acquiring MySQL named lock")

	// Generate lock name from database name
	lockName := fmt.Sprintf("sil_migration_%s", a.config.TableName)
	lockTimeout := int(a.config.LockTimeout.Seconds())

	// Acquire lock with timeout
	query := "SELECT GET_LOCK(?, ?)"
	var result int
	if err := a.db.QueryRowContext(ctx, query, lockName, lockTimeout).Scan(&result); err != nil {
		return nil, fmt.Errorf("failed to acquire lock: %w", err)
	}

	if result != 1 {
		return nil, sil.ErrLockAcquisitionFailed
	}

	a.logger.Debug("MySQL named lock acquired")

	return &mysqlLock{
		db:       a.db,
		lockName: lockName,
		locked:   true,
		logger:   a.logger,
	}, nil
}

// mysqlTransaction implements the Transaction interface for MySQL.
type mysqlTransaction struct {
	tx     *sql.Tx
	logger sil.Logger
}

func (t *mysqlTransaction) Commit() error {
	t.logger.Debug("Committing transaction")
	return t.tx.Commit()
}

func (t *mysqlTransaction) Rollback() error {
	t.logger.Debug("Rolling back transaction")
	return t.tx.Rollback()
}

func (t *mysqlTransaction) Exec(ctx context.Context, query string, args ...interface{}) error {
	t.logger.Debug("Executing query in transaction")
	_, err := t.tx.ExecContext(ctx, query, args...)
	return err
}

func (t *mysqlTransaction) Query(ctx context.Context, query string, args ...interface{}) (sil.Rows, error) {
	t.logger.Debug("Querying in transaction")
	rows, err := t.tx.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

// mysqlLock implements the Lock interface for MySQL.
type mysqlLock struct {
	db       *sql.DB
	lockName string
	locked   bool
	logger   sil.Logger
}

func (l *mysqlLock) Release() error {
	if !l.locked {
		return nil
	}

	l.logger.Debug("Releasing MySQL named lock")

	query := "SELECT RELEASE_LOCK(?)"
	var result sql.NullInt64
	if err := l.db.QueryRow(query, l.lockName).Scan(&result); err != nil {
		return fmt.Errorf("failed to release lock: %w", err)
	}

	l.locked = false
	return nil
}

func (l *mysqlLock) IsLocked() bool {
	return l.locked
}

// hashString generates a consistent hash for a string.
func hashString(s string) int64 {
	h := fnv.New64a()
	h.Write([]byte(s))
	return int64(h.Sum64())
}
