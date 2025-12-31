package sil_test

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/toutaio/toutago-sil-migrator/pkg/sil"
	"github.com/toutaio/toutago-sil-migrator/pkg/sil/adapters"
)

// IntegrationTestConfig returns a config for integration testing.
// Set DATABASE_URL environment variable to run against a real database.
func IntegrationTestConfig() *sil.Config {
	config := sil.DefaultConfig()
	config.DatabaseURL = "postgres://postgres:postgres@localhost:5433/sil_test?sslmode=disable"
	config.MigrationsDir = "./test_migrations"
	config.TableName = "sil_test_migrations"
	config.Verbose = true
	return config
}

// TestIntegrationPostgresAdapter tests the PostgreSQL adapter with a real database.
func TestIntegrationPostgresAdapter(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	config := IntegrationTestConfig()

	// Check if database is available
	if !isDatabaseAvailable(config.DatabaseURL) {
		t.Skip("PostgreSQL database not available")
	}

	// Clean up any existing test data
	cleanupDatabase(t, config)

	adapter, err := adapters.NewPostgresAdapter(config)
	if err != nil {
		t.Fatalf("Failed to create adapter: %v", err)
	}

	ctx := context.Background()

	// Test connection
	if err := adapter.Connect(ctx, config); err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer func() { _ = adapter.Close() }()

	// Test create migrations table
	if err := adapter.CreateMigrationsTable(ctx); err != nil {
		t.Fatalf("Failed to create migrations table: %v", err)
	}

	// Test record migration
	if err := adapter.RecordMigration(ctx, "20241230000001", "test migration", 1); err != nil {
		t.Fatalf("Failed to record migration: %v", err)
	}

	// Test get applied migrations
	applied, err := adapter.GetAppliedMigrations(ctx)
	if err != nil {
		t.Fatalf("Failed to get applied migrations: %v", err)
	}

	if len(applied) != 1 {
		t.Errorf("Expected 1 applied migration, got %d", len(applied))
	}

	if applied[0].Version != "20241230000001" {
		t.Errorf("Expected version 20241230000001, got %s", applied[0].Version)
	}

	// Test get last batch
	batch, err := adapter.GetLastBatch(ctx)
	if err != nil {
		t.Fatalf("Failed to get last batch: %v", err)
	}

	if batch != 1 {
		t.Errorf("Expected batch 1, got %d", batch)
	}

	// Test remove migration
	if err := adapter.RemoveMigration(ctx, "20241230000001"); err != nil {
		t.Fatalf("Failed to remove migration: %v", err)
	}

	applied, err = adapter.GetAppliedMigrations(ctx)
	if err != nil {
		t.Fatalf("Failed to get applied migrations: %v", err)
	}

	if len(applied) != 0 {
		t.Errorf("Expected 0 applied migrations, got %d", len(applied))
	}
}

// TestIntegrationLocking tests distributed locking with real database.
func TestIntegrationLocking(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	config := IntegrationTestConfig()

	if !isDatabaseAvailable(config.DatabaseURL) {
		t.Skip("PostgreSQL database not available")
	}

	cleanupDatabase(t, config)

	adapter1, err := adapters.NewPostgresAdapter(config)
	if err != nil {
		t.Fatalf("Failed to create adapter1: %v", err)
	}

	ctx := context.Background()
	if err := adapter1.Connect(ctx, config); err != nil {
		t.Fatalf("Failed to connect adapter1: %v", err)
	}
	defer func() { _ = adapter1.Close() }()

	// Acquire lock with first adapter
	lock1, err := adapter1.Lock(ctx)
	if err != nil {
		t.Fatalf("Failed to acquire lock1: %v", err)
	}
	defer lock1.Release()

	if !lock1.IsLocked() {
		t.Error("Lock1 should be locked")
	}

	// Try to acquire lock with second adapter (should timeout)
	adapter2, err := adapters.NewPostgresAdapter(config)
	if err != nil {
		t.Fatalf("Failed to create adapter2: %v", err)
	}

	if err := adapter2.Connect(ctx, config); err != nil {
		t.Fatalf("Failed to connect adapter2: %v", err)
	}
	defer func() { _ = adapter2.Close() }()

	// Create a context with short timeout
	ctxTimeout, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	// This should fail because lock1 is held
	_, err = adapter2.Lock(ctxTimeout)
	if err == nil {
		t.Error("Expected lock acquisition to fail, but it succeeded")
	}

	// Release first lock
	if err := lock1.Release(); err != nil {
		t.Fatalf("Failed to release lock1: %v", err)
	}

	if lock1.IsLocked() {
		t.Error("Lock1 should not be locked after release")
	}

	// Now second adapter should be able to acquire lock
	lock2, err := adapter2.Lock(ctx)
	if err != nil {
		t.Fatalf("Failed to acquire lock2 after release: %v", err)
	}
	defer lock2.Release()

	if !lock2.IsLocked() {
		t.Error("Lock2 should be locked")
	}
}

// TestIntegrationMigrationLifecycle tests the full migration lifecycle.
func TestIntegrationMigrationLifecycle(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	config := IntegrationTestConfig()

	if !isDatabaseAvailable(config.DatabaseURL) {
		t.Skip("PostgreSQL database not available")
	}

	cleanupDatabase(t, config)

	// Clear registry and register test migrations
	sil.ClearRegisteredMigrations()
	defer sil.ClearRegisteredMigrations()

	// Register test migrations
	sil.RegisterMigration(createTestMigration("20241230000001", "create_test_table", `
		CREATE TABLE test_users (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) NOT NULL
		)
	`, `DROP TABLE IF EXISTS test_users`))

	sil.RegisterMigration(createTestMigration("20241230000002", "add_email_column", `
		ALTER TABLE test_users ADD COLUMN email VARCHAR(255)
	`, `ALTER TABLE test_users DROP COLUMN IF EXISTS email`))

	// Create adapter
	adapter, err := adapters.NewPostgresAdapter(config)
	if err != nil {
		t.Fatalf("Failed to create adapter: %v", err)
	}

	ctx := context.Background()
	if err := adapter.Connect(ctx, config); err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer func() { _ = adapter.Close() }()

	// Create migrator
	migrator, err := sil.NewMigrator(config, adapter)
	if err != nil {
		t.Fatalf("Failed to create migrator: %v", err)
	}

	// Test migrate all
	if err := migrator.Migrate(ctx); err != nil {
		t.Fatalf("Failed to migrate: %v", err)
	}

	// Verify migrations were applied
	statuses, err := migrator.Status(ctx)
	if err != nil {
		t.Fatalf("Failed to get status: %v", err)
	}

	if len(statuses) != 2 {
		t.Fatalf("Expected 2 migrations, got %d", len(statuses))
	}

	for _, status := range statuses {
		if !status.Applied {
			t.Errorf("Migration %s should be applied", status.Version)
		}
	}

	// Verify table was created
	var exists bool
	rows, err := adapter.Query(ctx, `
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_name = 'test_users'
		)
	`)
	if err != nil {
		t.Fatalf("Failed to check table existence: %v", err)
	}
	defer func() { _ = rows.Close() }()

	if rows.Next() {
		if err := rows.Scan(&exists); err != nil {
			t.Fatalf("Failed to scan result: %v", err)
		}
	}

	if !exists {
		t.Error("test_users table should exist")
	}

	// Test rollback last batch
	if err := migrator.Rollback(ctx); err != nil {
		t.Fatalf("Failed to rollback: %v", err)
	}

	// Verify migrations were rolled back
	statuses, err = migrator.Status(ctx)
	if err != nil {
		t.Fatalf("Failed to get status after rollback: %v", err)
	}

	for _, status := range statuses {
		if status.Applied {
			t.Errorf("Migration %s should not be applied after rollback", status.Version)
		}
	}

	// Verify table was dropped
	rows2, err := adapter.Query(ctx, `
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_name = 'test_users'
		)
	`)
	if err != nil {
		t.Fatalf("Failed to check table existence: %v", err)
	}
	defer func() { _ = rows2.Close() }()

	exists = false
	if rows2.Next() {
		if err := rows2.Scan(&exists); err != nil {
			t.Fatalf("Failed to scan result: %v", err)
		}
	}

	if exists {
		t.Error("test_users table should not exist after rollback")
	}
}

// TestIntegrationTransactionRollback tests that failed migrations rollback.
func TestIntegrationTransactionRollback(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	config := IntegrationTestConfig()

	if !isDatabaseAvailable(config.DatabaseURL) {
		t.Skip("PostgreSQL database not available")
	}

	cleanupDatabase(t, config)

	// Clear registry and register test migrations
	sil.ClearRegisteredMigrations()
	defer sil.ClearRegisteredMigrations()

	// Register migration that will fail
	sil.RegisterMigration(createTestMigration("20241230000001", "failing_migration", `
		CREATE TABLE test_table (id SERIAL PRIMARY KEY);
		INSERT INTO non_existent_table VALUES (1);
	`, `DROP TABLE IF EXISTS test_table`))

	// Create adapter
	adapter, err := adapters.NewPostgresAdapter(config)
	if err != nil {
		t.Fatalf("Failed to create adapter: %v", err)
	}

	ctx := context.Background()
	if err := adapter.Connect(ctx, config); err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer func() { _ = adapter.Close() }()

	// Create migrator
	migrator, err := sil.NewMigrator(config, adapter)
	if err != nil {
		t.Fatalf("Failed to create migrator: %v", err)
	}

	// Migration should fail
	err = migrator.Migrate(ctx)
	if err == nil {
		t.Fatal("Expected migration to fail, but it succeeded")
	}

	// Verify test_table was not created (rollback worked)
	var exists bool
	rows, queryErr := adapter.Query(ctx, `
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_name = 'test_table'
		)
	`)
	if queryErr != nil {
		t.Fatalf("Failed to check table existence: %v", queryErr)
	}
	defer func() { _ = rows.Close() }()

	if rows.Next() {
		if err := rows.Scan(&exists); err != nil {
			t.Fatalf("Failed to scan result: %v", err)
		}
	}

	if exists {
		t.Error("test_table should not exist (transaction should have rolled back)")
	}

	// Verify migration was not recorded
	statuses, err := migrator.Status(ctx)
	if err != nil {
		t.Fatalf("Failed to get status: %v", err)
	}

	for _, status := range statuses {
		if status.Applied {
			t.Error("Failed migration should not be recorded as applied")
		}
	}
}

// Helper functions

func isDatabaseAvailable(dbURL string) bool {
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return false
	}
	defer func() { _ = db.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	return db.PingContext(ctx) == nil
}

func cleanupDatabase(t *testing.T, config *sil.Config) {
	db, err := sql.Open("postgres", config.DatabaseURL)
	if err != nil {
		t.Logf("Failed to open database for cleanup: %v", err)
		return
	}
	defer func() { _ = db.Close() }()

	// Drop test tables
	db.Exec("DROP TABLE IF EXISTS test_users CASCADE")
	db.Exec("DROP TABLE IF EXISTS test_table CASCADE")
	db.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s CASCADE", config.TableName))
}

func createTestMigration(version, description, upSQL, downSQL string) sil.Migration {
	return &testMigration{
		version:     version,
		description: description,
		upSQL:       upSQL,
		downSQL:     downSQL,
	}
}

type testMigration struct {
	sil.BaseMigration
	version     string
	description string
	upSQL       string
	downSQL     string
}

func (m *testMigration) Version() string {
	return m.version
}

func (m *testMigration) Description() string {
	return m.description
}

func (m *testMigration) Up(adapter sil.DatabaseAdapter) error {
	ctx := context.Background()
	return adapter.Exec(ctx, m.upSQL)
}

func (m *testMigration) Down(adapter sil.DatabaseAdapter) error {
	ctx := context.Background()
	return adapter.Exec(ctx, m.downSQL)
}

// TestIntegrationTransactionCommit tests transaction commit path.
func TestIntegrationTransactionCommit(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	config := IntegrationTestConfig()
	if !isDatabaseAvailable(config.DatabaseURL) {
		t.Skip("PostgreSQL database not available")
	}

	cleanupDatabase(t, config)

	adapter, err := adapters.NewPostgresAdapter(config)
	if err != nil {
		t.Fatalf("Failed to create adapter: %v", err)
	}

	ctx := context.Background()
	if err := adapter.Connect(ctx, config); err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer func() { _ = adapter.Close() }()

	if err := adapter.CreateMigrationsTable(ctx); err != nil {
		t.Fatalf("Failed to create migrations table: %v", err)
	}

	// Begin transaction
	tx, err := adapter.BeginTx(ctx)
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}

	// Execute within transaction
	if err := tx.Exec(ctx, "INSERT INTO "+config.TableName+" (version, description, batch) VALUES ($1, $2, $3)", "20240101000000", "test", 1); err != nil {
		t.Fatalf("Failed to execute in transaction: %v", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		t.Fatalf("Failed to commit transaction: %v", err)
	}

	// Verify data was committed
	applied, err := adapter.GetAppliedMigrations(ctx)
	if err != nil {
		t.Fatalf("Failed to get applied migrations: %v", err)
	}

	if len(applied) != 1 {
		t.Errorf("Expected 1 applied migration, got %d", len(applied))
	}
}

// TestIntegrationTransactionRollbackOnError tests automatic rollback on error.
func TestIntegrationTransactionRollbackOnError(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	config := IntegrationTestConfig()
	if !isDatabaseAvailable(config.DatabaseURL) {
		t.Skip("PostgreSQL database not available")
	}

	cleanupDatabase(t, config)

	adapter, err := adapters.NewPostgresAdapter(config)
	if err != nil {
		t.Fatalf("Failed to create adapter: %v", err)
	}

	ctx := context.Background()
	if err := adapter.Connect(ctx, config); err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer func() { _ = adapter.Close() }()

	if err := adapter.CreateMigrationsTable(ctx); err != nil {
		t.Fatalf("Failed to create migrations table: %v", err)
	}

	// Begin transaction
	tx, err := adapter.BeginTx(ctx)
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}

	// Execute within transaction
	if err := tx.Exec(ctx, "INSERT INTO "+config.TableName+" (version, description, batch) VALUES ($1, $2, $3)", "20240101000000", "test", 1); err != nil {
		t.Fatalf("Failed to execute in transaction: %v", err)
	}

	// Rollback transaction
	if err := tx.Rollback(); err != nil {
		t.Fatalf("Failed to rollback transaction: %v", err)
	}

	// Verify data was rolled back
	applied, err := adapter.GetAppliedMigrations(ctx)
	if err != nil {
		t.Fatalf("Failed to get applied migrations: %v", err)
	}

	if len(applied) != 0 {
		t.Errorf("Expected 0 applied migrations after rollback, got %d", len(applied))
	}
}

// TestIntegrationLockAcquisition tests lock acquisition and release.
func TestIntegrationLockAcquisition(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	config := IntegrationTestConfig()
	if !isDatabaseAvailable(config.DatabaseURL) {
		t.Skip("PostgreSQL database not available")
	}

	adapter, err := adapters.NewPostgresAdapter(config)
	if err != nil {
		t.Fatalf("Failed to create adapter: %v", err)
	}

	ctx := context.Background()
	if err := adapter.Connect(ctx, config); err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer func() { _ = adapter.Close() }()

	// Acquire lock
	lock, err := adapter.Lock(ctx)
	if err != nil {
		t.Fatalf("Failed to acquire lock: %v", err)
	}

	// Verify lock is held
	if !lock.IsLocked() {
		t.Error("Expected lock to be held")
	}

	// Release lock
	if err := lock.Release(); err != nil {
		t.Fatalf("Failed to release lock: %v", err)
	}

	// Verify lock is released
	if lock.IsLocked() {
		t.Error("Expected lock to be released")
	}
}

// TestIntegrationTransactionQuery tests query within transaction.
func TestIntegrationTransactionQuery(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	config := IntegrationTestConfig()
	if !isDatabaseAvailable(config.DatabaseURL) {
		t.Skip("PostgreSQL database not available")
	}

	cleanupDatabase(t, config)

	adapter, err := adapters.NewPostgresAdapter(config)
	if err != nil {
		t.Fatalf("Failed to create adapter: %v", err)
	}

	ctx := context.Background()
	if err := adapter.Connect(ctx, config); err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer func() { _ = adapter.Close() }()

	if err := adapter.CreateMigrationsTable(ctx); err != nil {
		t.Fatalf("Failed to create migrations table: %v", err)
	}

	// Begin transaction
	tx, err := adapter.BeginTx(ctx)
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	// Query within transaction
	rows, err := tx.Query(ctx, "SELECT version FROM "+config.TableName)
	if err != nil {
		t.Fatalf("Failed to query in transaction: %v", err)
	}
	defer func() { _ = rows.Close() }()

	// Count rows
	count := 0
	for rows.Next() {
		count++
	}

	if count != 0 {
		t.Errorf("Expected 0 rows, got %d", count)
	}
}
