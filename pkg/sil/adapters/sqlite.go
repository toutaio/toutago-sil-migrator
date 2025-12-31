package adapters

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/toutaio/toutago-sil-migrator/pkg/sil"
)

// SQLiteAdapter implements DatabaseAdapter for SQLite databases.
type SQLiteAdapter struct {
	db       *sql.DB
	dbPath   string
	lockFile string
	config   *sil.Config
	logger   sil.Logger
}

// NewSQLiteAdapter creates a new SQLite database adapter.
func NewSQLiteAdapter(config *sil.Config) (*SQLiteAdapter, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	if config.DatabaseURL == "" {
		return nil, fmt.Errorf("database URL is required")
	}

	// Extract database path from URL (format: sqlite://path/to/db.sqlite or file:path/to/db.sqlite)
	dbPath := config.DatabaseURL
	if len(dbPath) > 9 && dbPath[:9] == "sqlite://" {
		dbPath = dbPath[9:]
	} else if len(dbPath) > 7 && dbPath[:7] == "file://" {
		dbPath = dbPath[7:]
	}

	// Create lock file path
	lockFile := dbPath + ".lock"

	return &SQLiteAdapter{
		dbPath:   dbPath,
		lockFile: lockFile,
		config:   config,
		logger:   sil.NewDefaultLogger(config.Verbose),
	}, nil
}

// Connect establishes a connection to the SQLite database.
func (a *SQLiteAdapter) Connect(ctx context.Context, config *sil.Config) error {
	a.logger.Debug("Connecting to SQLite database: %s", a.dbPath)

	// Ensure directory exists
	dir := filepath.Dir(a.dbPath)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create database directory: %w", err)
		}
	}

	// Open database with WAL mode for better concurrency
	dsn := fmt.Sprintf("%s?_journal_mode=WAL&_busy_timeout=5000", a.dbPath)
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	// SQLite performs better with single connection for migrations
	db.SetMaxOpenConns(1)

	// Test connection
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return fmt.Errorf("failed to ping database: %w", err)
	}

	// Enable foreign keys
	if _, err := db.ExecContext(ctx, "PRAGMA foreign_keys = ON"); err != nil {
		_ = db.Close()
		return fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	a.db = db
	a.logger.Info("Connected to SQLite database")
	return nil
}

// Close closes the database connection.
func (a *SQLiteAdapter) Close() error {
	a.logger.Debug("Closing database connection")
	if a.db != nil {
		return a.db.Close()
	}
	return nil
}

// Exec executes a SQL statement.
func (a *SQLiteAdapter) Exec(ctx context.Context, query string, args ...interface{}) error {
	a.logger.Debug("Executing query")
	_, err := a.db.ExecContext(ctx, query, args...)
	return err
}

// Query executes a SQL query and returns rows.
func (a *SQLiteAdapter) Query(ctx context.Context, query string, args ...interface{}) (sil.Rows, error) {
	a.logger.Debug("Executing query")
	rows, err := a.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

// BeginTx begins a new transaction.
func (a *SQLiteAdapter) BeginTx(ctx context.Context) (sil.Transaction, error) {
	a.logger.Debug("Beginning transaction")
	tx, err := a.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	return &sqliteTransaction{
		tx:     tx,
		logger: a.logger,
	}, nil
}

// CreateMigrationsTable creates the migrations tracking table.
func (a *SQLiteAdapter) CreateMigrationsTable(ctx context.Context) error {
	a.logger.Debug("Creating migrations table if not exists")

	query := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			version TEXT NOT NULL UNIQUE,
			description TEXT,
			batch INTEGER NOT NULL,
			executed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`, a.config.TableName)

	if err := a.Exec(ctx, query); err != nil {
		return err
	}

	// Create index for batch column
	indexQuery := fmt.Sprintf(`
		CREATE INDEX IF NOT EXISTS idx_%s_batch ON %s(batch)
	`, a.config.TableName, a.config.TableName)

	return a.Exec(ctx, indexQuery)
}

// GetAppliedMigrations returns all applied migrations.
func (a *SQLiteAdapter) GetAppliedMigrations(ctx context.Context) ([]sil.MigrationRecord, error) {
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
	defer func() { _ = rows.Close() }()

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
func (a *SQLiteAdapter) RecordMigration(ctx context.Context, version, description string, batch int) error {
	a.logger.Debug("Recording migration: %s", version)

	query := fmt.Sprintf(`
		INSERT INTO %s (version, description, batch)
		VALUES (?, ?, ?)
	`, a.config.TableName)

	return a.Exec(ctx, query, version, description, batch)
}

// RemoveMigration removes a migration record.
func (a *SQLiteAdapter) RemoveMigration(ctx context.Context, version string) error {
	a.logger.Debug("Removing migration record: %s", version)

	query := fmt.Sprintf(`
		DELETE FROM %s WHERE version = ?
	`, a.config.TableName)

	return a.Exec(ctx, query, version)
}

// GetLastBatchNumber returns the last batch number.
func (a *SQLiteAdapter) GetLastBatch(ctx context.Context) (int, error) {
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

// Lock acquires a file-based lock for migrations.
func (a *SQLiteAdapter) Lock(ctx context.Context) (sil.Lock, error) {
	a.logger.Debug("Acquiring file lock: %s", a.lockFile)

	// Check for stale lock file
	if info, err := os.Stat(a.lockFile); err == nil {
		// Lock file exists, check if stale (older than lock timeout)
		if time.Since(info.ModTime()) > a.config.LockTimeout {
			a.logger.Warn("Removing stale lock file")
			_ = os.Remove(a.lockFile)
		} else {
			return nil, sil.ErrLockAcquisitionFailed
		}
	}

	// Create lock file
	file, err := os.OpenFile(a.lockFile, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
	if err != nil {
		if os.IsExist(err) {
			return nil, sil.ErrLockAcquisitionFailed
		}
		return nil, fmt.Errorf("failed to create lock file: %w", err)
	}

	// Write process info to lock file
	_, _ = fmt.Fprintf(file, "pid: %d\ntime: %s\n", os.Getpid(), time.Now().Format(time.RFC3339))
	_ = file.Close()

	a.logger.Debug("File lock acquired")

	return &sqliteLock{
		lockFile: a.lockFile,
		locked:   true,
		logger:   a.logger,
	}, nil
}

// sqliteTransaction implements the Transaction interface for SQLite.
type sqliteTransaction struct {
	tx     *sql.Tx
	logger sil.Logger
}

func (t *sqliteTransaction) Commit() error {
	t.logger.Debug("Committing transaction")
	return t.tx.Commit()
}

func (t *sqliteTransaction) Rollback() error {
	t.logger.Debug("Rolling back transaction")
	return t.tx.Rollback()
}

func (t *sqliteTransaction) Exec(ctx context.Context, query string, args ...interface{}) error {
	t.logger.Debug("Executing query in transaction")
	_, err := t.tx.ExecContext(ctx, query, args...)
	return err
}

func (t *sqliteTransaction) Query(ctx context.Context, query string, args ...interface{}) (sil.Rows, error) {
	t.logger.Debug("Querying in transaction")
	rows, err := t.tx.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

// sqliteLock implements the Lock interface using file-based locking.
type sqliteLock struct {
	lockFile string
	locked   bool
	logger   sil.Logger
}

func (l *sqliteLock) Release() error {
	if !l.locked {
		return nil
	}

	l.logger.Debug("Releasing file lock")

	if err := os.Remove(l.lockFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove lock file: %w", err)
	}

	l.locked = false
	return nil
}

func (l *sqliteLock) IsLocked() bool {
	return l.locked
}
