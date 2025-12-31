package adapters

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/toutaio/toutago-sil-migrator/pkg/sil"
)

func TestPostgresAdapter_CreateMigrationsTable(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer func() { _ = db.Close() }()

	adapter := &PostgresAdapter{
		db: db,
		config: &sil.Config{
			TableName: "migrations",
		},
		logger: sil.NewDefaultLogger(false),
	}

	mock.ExpectExec("CREATE TABLE IF NOT EXISTS migrations").
		WillReturnResult(sqlmock.NewResult(0, 0))

	ctx := context.Background()
	err = adapter.CreateMigrationsTable(ctx)

	if err != nil {
		t.Errorf("CreateMigrationsTable() error = %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestPostgresAdapter_GetAppliedMigrations(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer func() { _ = db.Close() }()

	adapter := &PostgresAdapter{
		db: db,
		config: &sil.Config{
			TableName: "migrations",
		},
		logger: sil.NewDefaultLogger(false),
	}

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "version", "description", "batch", "executed_at"}).
		AddRow(1, "20240101_create_users", "Create users", 1, now).
		AddRow(2, "20240102_create_posts", "Create posts", 1, now)

	mock.ExpectQuery("SELECT (.+) FROM migrations ORDER BY executed_at ASC").
		WillReturnRows(rows)

	ctx := context.Background()
	migrations, err := adapter.GetAppliedMigrations(ctx)

	if err != nil {
		t.Errorf("GetAppliedMigrations() error = %v", err)
	}

	if len(migrations) != 2 {
		t.Errorf("GetAppliedMigrations() got %d migrations, want 2", len(migrations))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestPostgresAdapter_RecordMigration(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer func() { _ = db.Close() }()

	adapter := &PostgresAdapter{
		db: db,
		config: &sil.Config{
			TableName: "migrations",
		},
		logger: sil.NewDefaultLogger(false),
	}

	mock.ExpectExec("INSERT INTO migrations").
		WithArgs("20240101_create_users", "Create users", 1).
		WillReturnResult(sqlmock.NewResult(1, 1))

	ctx := context.Background()
	err = adapter.RecordMigration(ctx, "20240101_create_users", "Create users", 1)

	if err != nil {
		t.Errorf("RecordMigration() error = %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestPostgresAdapter_Exec(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer func() { _ = db.Close() }()

	adapter := &PostgresAdapter{
		db:     db,
		logger: sil.NewDefaultLogger(false),
	}

	mock.ExpectExec("CREATE TABLE test").
		WillReturnResult(sqlmock.NewResult(0, 0))

	ctx := context.Background()
	err = adapter.Exec(ctx, "CREATE TABLE test")

	if err != nil {
		t.Errorf("Exec() error = %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestPostgresAdapter_Query(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer func() { _ = db.Close() }()

	adapter := &PostgresAdapter{
		db:     db,
		logger: sil.NewDefaultLogger(false),
	}

	rows := sqlmock.NewRows([]string{"id", "name"}).AddRow(1, "test")
	mock.ExpectQuery("SELECT (.+) FROM test").
		WillReturnRows(rows)

	ctx := context.Background()
	result, err := adapter.Query(ctx, "SELECT * FROM test")

	if err != nil {
		t.Errorf("Query() error = %v", err)
	}

	if result == nil {
		t.Error("Query() returned nil result")
	}

	defer func() { _ = result.Close() }()

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestPostgresAdapter_BeginTx(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer func() { _ = db.Close() }()

	adapter := &PostgresAdapter{
		db:     db,
		logger: sil.NewDefaultLogger(false),
	}

	mock.ExpectBegin()

	ctx := context.Background()
	tx, err := adapter.BeginTx(ctx)

	if err != nil {
		t.Errorf("BeginTx() error = %v", err)
	}

	if tx == nil {
		t.Error("BeginTx() returned nil transaction")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestPostgresAdapter_Close(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}

	adapter := &PostgresAdapter{
		db:     db,
		logger: sil.NewDefaultLogger(false),
	}

	mock.ExpectClose()

	err = adapter.Close()

	if err != nil {
		t.Errorf("Close() error = %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestPostgresAdapter_Lock(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer func() { _ = db.Close() }()

	adapter := &PostgresAdapter{
		db: db,
		config: &sil.Config{
			TableName:   "migrations",
			LockTimeout: 10 * time.Second,
		},
		logger: sil.NewDefaultLogger(false),
	}

	rows := sqlmock.NewRows([]string{"locked"}).AddRow(true)
	mock.ExpectQuery("SELECT pg_try_advisory_lock").
		WillReturnRows(rows)

	ctx := context.Background()
	lock, err := adapter.Lock(ctx)

	if err != nil {
		t.Errorf("Lock() error = %v", err)
	}

	if lock == nil {
		t.Error("Lock() returned nil lock")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestPostgresAdapter_GetAppliedMigrations_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer func() { _ = db.Close() }()

	adapter := &PostgresAdapter{
		db: db,
		config: &sil.Config{
			TableName: "migrations",
		},
		logger: sil.NewDefaultLogger(false),
	}

	mock.ExpectQuery("SELECT (.+) FROM migrations").
		WillReturnError(errors.New("query failed"))

	ctx := context.Background()
	_, err = adapter.GetAppliedMigrations(ctx)

	if err == nil {
		t.Error("GetAppliedMigrations() expected error, got nil")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestPostgresAdapter_RemoveMigration(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer func() { _ = db.Close() }()

	adapter := &PostgresAdapter{
		db: db,
		config: &sil.Config{
			TableName: "migrations",
		},
		logger: sil.NewDefaultLogger(false),
	}

	mock.ExpectExec("DELETE FROM migrations").
		WithArgs("20240101_create_users").
		WillReturnResult(sqlmock.NewResult(0, 1))

	ctx := context.Background()
	err = adapter.RemoveMigration(ctx, "20240101_create_users")

	if err != nil {
		t.Errorf("RemoveMigration() error = %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestPostgresAdapter_GetLastBatch(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer func() { _ = db.Close() }()

	adapter := &PostgresAdapter{
		db: db,
		config: &sil.Config{
			TableName: "migrations",
		},
		logger: sil.NewDefaultLogger(false),
	}

	rows := sqlmock.NewRows([]string{"batch"}).AddRow(5)
	mock.ExpectQuery("SELECT COALESCE\\(MAX\\(batch\\), 0\\) FROM migrations").
		WillReturnRows(rows)

	ctx := context.Background()
	batch, err := adapter.GetLastBatch(ctx)

	if err != nil {
		t.Errorf("GetLastBatch() error = %v", err)
	}

	if batch != 5 {
		t.Errorf("GetLastBatch() = %d, want 5", batch)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestPostgresTx_Commit(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer func() { _ = db.Close() }()

	mock.ExpectBegin()
	sqlTx, _ := db.Begin()

	tx := &transaction{
		tx:     sqlTx,
		logger: sil.NewDefaultLogger(false),
	}

	mock.ExpectCommit()

	err = tx.Commit()

	if err != nil {
		t.Errorf("Commit() error = %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestPostgresTx_Rollback(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer func() { _ = db.Close() }()

	mock.ExpectBegin()
	sqlTx, _ := db.Begin()

	tx := &transaction{
		tx:     sqlTx,
		logger: sil.NewDefaultLogger(false),
	}

	mock.ExpectRollback()

	err = tx.Rollback()

	if err != nil {
		t.Errorf("Rollback() error = %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestPostgresTx_Exec(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer func() { _ = db.Close() }()

	mock.ExpectBegin()
	sqlTx, _ := db.Begin()

	tx := &transaction{
		tx:     sqlTx,
		logger: sil.NewDefaultLogger(false),
	}

	mock.ExpectExec("INSERT INTO test").
		WillReturnResult(sqlmock.NewResult(1, 1))

	ctx := context.Background()
	err = tx.Exec(ctx, "INSERT INTO test VALUES (1)")

	if err != nil {
		t.Errorf("Exec() error = %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestPostgresTx_Query(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer func() { _ = db.Close() }()

	mock.ExpectBegin()
	sqlTx, _ := db.Begin()

	tx := &transaction{
		tx:     sqlTx,
		logger: sil.NewDefaultLogger(false),
	}

	rows := sqlmock.NewRows([]string{"id"}).AddRow(1)
	mock.ExpectQuery("SELECT (.+) FROM test").
		WillReturnRows(rows)

	ctx := context.Background()
	result, err := tx.Query(ctx, "SELECT * FROM test")

	if err != nil {
		t.Errorf("Query() error = %v", err)
	}

	if result == nil {
		t.Error("Query() returned nil result")
	}

	defer func() { _ = result.Close() }()

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestPostgresLock_Release(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer func() { _ = db.Close() }()

	releaseCalled := false
	releaseFunc := func() error {
		releaseCalled = true
		return nil
	}

	lock := newLock(releaseFunc, sil.NewDefaultLogger(false))

	err = lock.Release()

	if err != nil {
		t.Errorf("Release() error = %v", err)
	}

	if !releaseCalled {
		t.Error("Release function was not called")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestPostgresLock_IsLocked(t *testing.T) {
	releaseFunc := func() error {
		return nil
	}

	lock := newLock(releaseFunc, sil.NewDefaultLogger(false))

	if !lock.IsLocked() {
		t.Error("IsLocked() should return true for new lock")
	}

	lock.Release()

	if lock.IsLocked() {
		t.Error("IsLocked() should return false after release")
	}
}
