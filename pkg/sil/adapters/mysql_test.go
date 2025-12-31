package adapters

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/toutaio/toutago-sil-migrator/pkg/sil"
)

func TestMySQLAdapter_CreateMigrationsTable(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	adapter := &MySQLAdapter{
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

func TestMySQLAdapter_GetAppliedMigrations(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	adapter := &MySQLAdapter{
		db: db,
		config: &sil.Config{
			TableName: "migrations",
		},
		logger: sil.NewDefaultLogger(false),
	}

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "version", "description", "batch", "executed_at"}).
		AddRow(1, "20240101_create_users", "Create users table", 1, now).
		AddRow(2, "20240102_create_posts", "Create posts table", 1, now)

	mock.ExpectQuery("SELECT (.+) FROM migrations ORDER BY id ASC").
		WillReturnRows(rows)

	ctx := context.Background()
	migrations, err := adapter.GetAppliedMigrations(ctx)

	if err != nil {
		t.Errorf("GetAppliedMigrations() error = %v", err)
	}

	if len(migrations) != 2 {
		t.Errorf("GetAppliedMigrations() got %d migrations, want 2", len(migrations))
	}

	if migrations[0].Version != "20240101_create_users" {
		t.Errorf("GetAppliedMigrations() first migration = %v, want 20240101_create_users", migrations[0].Version)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestMySQLAdapter_GetAppliedMigrations_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	adapter := &MySQLAdapter{
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

func TestMySQLAdapter_RecordMigration(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	adapter := &MySQLAdapter{
		db: db,
		config: &sil.Config{
			TableName: "migrations",
		},
		logger: sil.NewDefaultLogger(false),
	}

	mock.ExpectExec("INSERT INTO migrations").
		WithArgs("20240101_create_users", "Create users table", 1).
		WillReturnResult(sqlmock.NewResult(1, 1))

	ctx := context.Background()
	err = adapter.RecordMigration(ctx, "20240101_create_users", "Create users table", 1)

	if err != nil {
		t.Errorf("RecordMigration() error = %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestMySQLAdapter_RemoveMigration(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	adapter := &MySQLAdapter{
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

func TestMySQLAdapter_GetLastBatch(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	adapter := &MySQLAdapter{
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

func TestMySQLAdapter_Lock(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	adapter := &MySQLAdapter{
		db: db,
		config: &sil.Config{
			TableName:   "migrations",
			LockTimeout: 10 * time.Second,
		},
		logger: sil.NewDefaultLogger(false),
	}

	rows := sqlmock.NewRows([]string{"acquired"}).AddRow(1)
	mock.ExpectQuery("SELECT GET_LOCK").
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

func TestMySQLAdapter_Lock_Failed(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	adapter := &MySQLAdapter{
		db: db,
		config: &sil.Config{
			TableName:   "migrations",
			LockTimeout: 10 * time.Second,
		},
		logger: sil.NewDefaultLogger(false),
	}

	rows := sqlmock.NewRows([]string{"acquired"}).AddRow(0)
	mock.ExpectQuery("SELECT GET_LOCK").
		WillReturnRows(rows)

	ctx := context.Background()
	_, err = adapter.Lock(ctx)

	if err == nil {
		t.Error("Lock() expected error when lock acquisition fails, got nil")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestMySQLAdapter_Exec(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	adapter := &MySQLAdapter{
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

func TestMySQLAdapter_Query(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	adapter := &MySQLAdapter{
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

	defer result.Close()

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestMySQLAdapter_BeginTx(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	adapter := &MySQLAdapter{
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

func TestMySQLAdapter_Close(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}

	adapter := &MySQLAdapter{
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

func TestMySQLTx_Commit(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	mock.ExpectBegin()
	sqlTx, _ := db.Begin()

	tx := &mysqlTransaction{
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

func TestMySQLTx_Rollback(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	mock.ExpectBegin()
	sqlTx, _ := db.Begin()

	tx := &mysqlTransaction{
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

func TestMySQLLock_Release(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	lock := &mysqlLock{
		db:       db,
		lockName: "test_lock",
		locked:   true,
		logger:   sil.NewDefaultLogger(false),
	}

	rows := sqlmock.NewRows([]string{"released"}).AddRow(1)
	mock.ExpectQuery("SELECT RELEASE_LOCK").
		WillReturnRows(rows)

	err = lock.Release()

	if err != nil {
		t.Errorf("Release() error = %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestMySQLLock_IsLocked(t *testing.T) {
	lock := &mysqlLock{
		lockName: "test_lock",
		locked:   true,
		logger:   sil.NewDefaultLogger(false),
	}

	if !lock.IsLocked() {
		t.Error("IsLocked() = false, want true")
	}

	lock.locked = false
	if lock.IsLocked() {
		t.Error("IsLocked() = true, want false")
	}
}

func TestHashString(t *testing.T) {
	hash1 := hashString("test_lock")
	hash2 := hashString("test_lock")
	hash3 := hashString("different_lock")

	if hash1 != hash2 {
		t.Error("hashString() should return same hash for same input")
	}

	if hash1 == hash3 {
		t.Error("hashString() should return different hash for different input")
	}
}
