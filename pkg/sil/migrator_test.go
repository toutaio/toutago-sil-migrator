package sil

import (
	"context"
	"errors"
	"testing"
	"time"
)

// mockAdapter is a mock database adapter for testing.
type mockAdapter struct {
	appliedMigrations []MigrationRecord
	lastBatch         int
	execFunc          func(ctx context.Context, query string, args ...interface{}) error
	queryFunc         func(ctx context.Context, query string, args ...interface{}) (Rows, error)
	lockAcquired      bool
	lockReleased      bool
}

func newMockAdapter() *mockAdapter {
	return &mockAdapter{
		appliedMigrations: []MigrationRecord{},
		lastBatch:         0,
	}
}

func (m *mockAdapter) Connect(ctx context.Context, config *Config) error {
	return nil
}

func (m *mockAdapter) Close() error {
	return nil
}

func (m *mockAdapter) Exec(ctx context.Context, query string, args ...interface{}) error {
	if m.execFunc != nil {
		return m.execFunc(ctx, query, args...)
	}
	return nil
}

func (m *mockAdapter) Query(ctx context.Context, query string, args ...interface{}) (Rows, error) {
	if m.queryFunc != nil {
		return m.queryFunc(ctx, query, args...)
	}
	return nil, nil
}

func (m *mockAdapter) BeginTx(ctx context.Context) (Transaction, error) {
	return &mockTransaction{}, nil
}

func (m *mockAdapter) CreateMigrationsTable(ctx context.Context) error {
	return nil
}

func (m *mockAdapter) GetAppliedMigrations(ctx context.Context) ([]MigrationRecord, error) {
	return m.appliedMigrations, nil
}

func (m *mockAdapter) RecordMigration(ctx context.Context, version, description string, batch int) error {
	m.appliedMigrations = append(m.appliedMigrations, MigrationRecord{
		Version:     version,
		Description: description,
		Batch:       batch,
		ExecutedAt:  time.Now(),
	})
	if batch > m.lastBatch {
		m.lastBatch = batch
	}
	return nil
}

func (m *mockAdapter) RemoveMigration(ctx context.Context, version string) error {
	for i, rec := range m.appliedMigrations {
		if rec.Version == version {
			m.appliedMigrations = append(m.appliedMigrations[:i], m.appliedMigrations[i+1:]...)
			return nil
		}
	}
	return nil
}

func (m *mockAdapter) Lock(ctx context.Context) (Lock, error) {
	if m.lockAcquired {
		return nil, errors.New("lock already acquired")
	}
	m.lockAcquired = true
	return &mockLock{adapter: m}, nil
}

func (m *mockAdapter) GetLastBatch(ctx context.Context) (int, error) {
	return m.lastBatch, nil
}

// mockTransaction is a mock transaction for testing.
type mockTransaction struct {
	committed  bool
	rolledBack bool
}

func (m *mockTransaction) Commit() error {
	m.committed = true
	return nil
}

func (m *mockTransaction) Rollback() error {
	m.rolledBack = true
	return nil
}

func (m *mockTransaction) Exec(ctx context.Context, query string, args ...interface{}) error {
	return nil
}

func (m *mockTransaction) Query(ctx context.Context, query string, args ...interface{}) (Rows, error) {
	return nil, nil
}

// mockLock is a mock lock for testing.
type mockLock struct {
	adapter  *mockAdapter
	released bool
}

func (m *mockLock) Release() error {
	if !m.released {
		m.adapter.lockReleased = true
		m.adapter.lockAcquired = false
		m.released = true
	}
	return nil
}

func (m *mockLock) IsLocked() bool {
	return !m.released
}

func TestNewMigrator(t *testing.T) {
	config := DefaultConfig()
	config.DatabaseURL = "postgres://localhost/test"
	adapter := newMockAdapter()

	migrator, err := NewMigrator(config, adapter)
	if err != nil {
		t.Fatalf("NewMigrator() error = %v", err)
	}

	if migrator == nil {
		t.Fatal("NewMigrator() returned nil")
	}
}

func TestNewMigrator_NilConfig(t *testing.T) {
	adapter := newMockAdapter()

	_, err := NewMigrator(nil, adapter)
	if err == nil {
		t.Error("NewMigrator() with nil config should return error")
	}
}

func TestNewMigrator_NilAdapter(t *testing.T) {
	config := DefaultConfig()
	config.DatabaseURL = "postgres://localhost/test"

	_, err := NewMigrator(config, nil)
	if err == nil {
		t.Error("NewMigrator() with nil adapter should return error")
	}
}

func TestMigrator_Migrate(t *testing.T) {
	// Clear registry and add test migrations
	ClearRegisteredMigrations()
	defer ClearRegisteredMigrations()

	upCalled := 0
	RegisterMigration(NewBaseMigration("20241230100000", "first", func(adapter DatabaseAdapter) error {
		upCalled++
		return nil
	}, nil))
	RegisterMigration(NewBaseMigration("20241230110000", "second", func(adapter DatabaseAdapter) error {
		upCalled++
		return nil
	}, nil))

	config := DefaultConfig()
	config.DatabaseURL = "postgres://localhost/test"
	adapter := newMockAdapter()

	migrator, err := NewMigrator(config, adapter)
	if err != nil {
		t.Fatalf("NewMigrator() error = %v", err)
	}

	ctx := context.Background()
	if err := migrator.Migrate(ctx); err != nil {
		t.Fatalf("Migrate() error = %v", err)
	}

	if upCalled != 2 {
		t.Errorf("Expected 2 migrations to run, got %d", upCalled)
	}

	if len(adapter.appliedMigrations) != 2 {
		t.Errorf("Expected 2 recorded migrations, got %d", len(adapter.appliedMigrations))
	}
}

func TestMigrator_Migrate_NoPending(t *testing.T) {
	// Clear registry and add test migrations
	ClearRegisteredMigrations()
	defer ClearRegisteredMigrations()

	RegisterMigration(NewBaseMigration("20241230100000", "first", func(adapter DatabaseAdapter) error {
		return nil
	}, nil))

	config := DefaultConfig()
	config.DatabaseURL = "postgres://localhost/test"
	adapter := newMockAdapter()

	// Pre-apply migration
	adapter.appliedMigrations = []MigrationRecord{
		{Version: "20241230100000", Batch: 1},
	}
	adapter.lastBatch = 1

	migrator, err := NewMigrator(config, adapter)
	if err != nil {
		t.Fatalf("NewMigrator() error = %v", err)
	}

	ctx := context.Background()
	if err := migrator.Migrate(ctx); err != nil {
		t.Fatalf("Migrate() error = %v", err)
	}

	// Should still be 1 applied migration
	if len(adapter.appliedMigrations) != 1 {
		t.Errorf("Expected 1 recorded migration, got %d", len(adapter.appliedMigrations))
	}
}

func TestMigrator_MigrateUp(t *testing.T) {
	// Clear registry and add test migrations
	ClearRegisteredMigrations()
	defer ClearRegisteredMigrations()

	upCalled := 0
	RegisterMigration(NewBaseMigration("20241230100000", "first", func(adapter DatabaseAdapter) error {
		upCalled++
		return nil
	}, nil))
	RegisterMigration(NewBaseMigration("20241230110000", "second", func(adapter DatabaseAdapter) error {
		upCalled++
		return nil
	}, nil))
	RegisterMigration(NewBaseMigration("20241230120000", "third", func(adapter DatabaseAdapter) error {
		upCalled++
		return nil
	}, nil))

	config := DefaultConfig()
	config.DatabaseURL = "postgres://localhost/test"
	adapter := newMockAdapter()

	migrator, err := NewMigrator(config, adapter)
	if err != nil {
		t.Fatalf("NewMigrator() error = %v", err)
	}

	ctx := context.Background()
	if err := migrator.MigrateUp(ctx, 2); err != nil {
		t.Fatalf("MigrateUp() error = %v", err)
	}

	if upCalled != 2 {
		t.Errorf("Expected 2 migrations to run, got %d", upCalled)
	}

	if len(adapter.appliedMigrations) != 2 {
		t.Errorf("Expected 2 recorded migrations, got %d", len(adapter.appliedMigrations))
	}
}

func TestMigrator_MigrateDown(t *testing.T) {
	// Clear registry and add test migrations
	ClearRegisteredMigrations()
	defer ClearRegisteredMigrations()

	downCalled := 0
	RegisterMigration(NewBaseMigration("20241230100000", "first", nil, func(adapter DatabaseAdapter) error {
		downCalled++
		return nil
	}))
	RegisterMigration(NewBaseMigration("20241230110000", "second", nil, func(adapter DatabaseAdapter) error {
		downCalled++
		return nil
	}))

	config := DefaultConfig()
	config.DatabaseURL = "postgres://localhost/test"
	adapter := newMockAdapter()

	// Pre-apply migrations
	adapter.appliedMigrations = []MigrationRecord{
		{Version: "20241230100000", Batch: 1, ExecutedAt: time.Now()},
		{Version: "20241230110000", Batch: 1, ExecutedAt: time.Now()},
	}
	adapter.lastBatch = 1

	migrator, err := NewMigrator(config, adapter)
	if err != nil {
		t.Fatalf("NewMigrator() error = %v", err)
	}

	ctx := context.Background()
	if err := migrator.MigrateDown(ctx, 1); err != nil {
		t.Fatalf("MigrateDown() error = %v", err)
	}

	if downCalled != 1 {
		t.Errorf("Expected 1 migration to roll back, got %d", downCalled)
	}

	if len(adapter.appliedMigrations) != 1 {
		t.Errorf("Expected 1 remaining migration, got %d", len(adapter.appliedMigrations))
	}
}

func TestMigrator_Rollback(t *testing.T) {
	// Clear registry and add test migrations
	ClearRegisteredMigrations()
	defer ClearRegisteredMigrations()

	downCalled := 0
	RegisterMigration(NewBaseMigration("20241230100000", "first", nil, func(adapter DatabaseAdapter) error {
		downCalled++
		return nil
	}))
	RegisterMigration(NewBaseMigration("20241230110000", "second", nil, func(adapter DatabaseAdapter) error {
		downCalled++
		return nil
	}))

	config := DefaultConfig()
	config.DatabaseURL = "postgres://localhost/test"
	adapter := newMockAdapter()

	// Pre-apply migrations in same batch
	adapter.appliedMigrations = []MigrationRecord{
		{Version: "20241230100000", Batch: 1, ExecutedAt: time.Now()},
		{Version: "20241230110000", Batch: 1, ExecutedAt: time.Now()},
	}
	adapter.lastBatch = 1

	migrator, err := NewMigrator(config, adapter)
	if err != nil {
		t.Fatalf("NewMigrator() error = %v", err)
	}

	ctx := context.Background()
	if err := migrator.Rollback(ctx); err != nil {
		t.Fatalf("Rollback() error = %v", err)
	}

	if downCalled != 2 {
		t.Errorf("Expected 2 migrations to roll back, got %d", downCalled)
	}

	if len(adapter.appliedMigrations) != 0 {
		t.Errorf("Expected 0 remaining migrations, got %d", len(adapter.appliedMigrations))
	}
}

func TestMigrator_Status(t *testing.T) {
	// Clear registry and add test migrations
	ClearRegisteredMigrations()
	defer ClearRegisteredMigrations()

	RegisterMigration(NewBaseMigration("20241230100000", "first", nil, nil))
	RegisterMigration(NewBaseMigration("20241230110000", "second", nil, nil))

	config := DefaultConfig()
	config.DatabaseURL = "postgres://localhost/test"
	adapter := newMockAdapter()

	// Pre-apply first migration
	adapter.appliedMigrations = []MigrationRecord{
		{Version: "20241230100000", Batch: 1, ExecutedAt: time.Now()},
	}
	adapter.lastBatch = 1

	migrator, err := NewMigrator(config, adapter)
	if err != nil {
		t.Fatalf("NewMigrator() error = %v", err)
	}

	ctx := context.Background()
	statuses, err := migrator.Status(ctx)
	if err != nil {
		t.Fatalf("Status() error = %v", err)
	}

	if len(statuses) != 2 {
		t.Fatalf("Expected 2 statuses, got %d", len(statuses))
	}

	if !statuses[0].Applied {
		t.Error("Expected first migration to be applied")
	}

	if statuses[1].Applied {
		t.Error("Expected second migration to not be applied")
	}
}

func TestMigrator_Reset(t *testing.T) {
	// Clear registry and add test migrations
	ClearRegisteredMigrations()
	defer ClearRegisteredMigrations()

	downCalled := 0
	RegisterMigration(NewBaseMigration("20241230100000", "first", nil, func(adapter DatabaseAdapter) error {
		downCalled++
		return nil
	}))
	RegisterMigration(NewBaseMigration("20241230110000", "second", nil, func(adapter DatabaseAdapter) error {
		downCalled++
		return nil
	}))

	config := DefaultConfig()
	config.DatabaseURL = "postgres://localhost/test"
	adapter := newMockAdapter()

	// Pre-apply migrations
	adapter.appliedMigrations = []MigrationRecord{
		{Version: "20241230100000", Batch: 1, ExecutedAt: time.Now()},
		{Version: "20241230110000", Batch: 2, ExecutedAt: time.Now()},
	}
	adapter.lastBatch = 2

	migrator, err := NewMigrator(config, adapter)
	if err != nil {
		t.Fatalf("NewMigrator() error = %v", err)
	}

	ctx := context.Background()
	if err := migrator.Reset(ctx); err != nil {
		t.Fatalf("Reset() error = %v", err)
	}

	if downCalled != 2 {
		t.Errorf("Expected 2 migrations to roll back, got %d", downCalled)
	}

	if len(adapter.appliedMigrations) != 0 {
		t.Errorf("Expected 0 remaining migrations, got %d", len(adapter.appliedMigrations))
	}
}

func TestMigrator_Refresh(t *testing.T) {
// Clear registry and add test migration
ClearRegisteredMigrations()
defer ClearRegisteredMigrations()

upCalled := 0
downCalled := 0
RegisterMigration(NewBaseMigration("20241230100000", "first", 
func(adapter DatabaseAdapter) error {
upCalled++
return nil
}, 
func(adapter DatabaseAdapter) error {
downCalled++
return nil
}))

config := DefaultConfig()
config.DatabaseURL = "postgres://localhost/test"
adapter := newMockAdapter()

// Pre-apply migration
adapter.appliedMigrations = []MigrationRecord{
{Version: "20241230100000", Batch: 1, ExecutedAt: time.Now()},
}
adapter.lastBatch = 1

migrator, err := NewMigrator(config, adapter)
if err != nil {
t.Fatalf("NewMigrator() error = %v", err)
}

ctx := context.Background()
if err := migrator.Refresh(ctx); err != nil {
t.Fatalf("Refresh() error = %v", err)
}

if downCalled != 1 {
t.Errorf("Expected 1 down migration, got %d", downCalled)
}

if upCalled != 1 {
t.Errorf("Expected 1 up migration, got %d", upCalled)
}
}

func TestMigrator_SetBeforeMigrate(t *testing.T) {
config := DefaultConfig()
adapter := newMockAdapter()

migrator, err := NewMigrator(config, adapter)
if err != nil {
t.Fatalf("NewMigrator() error = %v", err)
}

called := false
migrator.SetBeforeMigrate(func(migration Migration, direction string) error {
called = true
return nil
})

// Just verify it doesn't crash - callback will be called during actual migration
if called {
t.Error("Expected callback not to be called yet")
}
}

func TestMigrator_SetAfterMigrate(t *testing.T) {
config := DefaultConfig()
adapter := newMockAdapter()

migrator, err := NewMigrator(config, adapter)
if err != nil {
t.Fatalf("NewMigrator() error = %v", err)
}

called := false
migrator.SetAfterMigrate(func(migration Migration, direction string) error {
called = true
return nil
})

if called {
t.Error("Expected callback not to be called yet")
}
}

func TestMigrator_SetOnError(t *testing.T) {
config := DefaultConfig()
adapter := newMockAdapter()

migrator, err := NewMigrator(config, adapter)
if err != nil {
t.Fatalf("NewMigrator() error = %v", err)
}

called := false
migrator.SetOnError(func(migration Migration, err error) {
called = true
})

if called {
t.Error("Expected callback not to be called yet")
}
}

func TestMigrator_SetLogger(t *testing.T) {
config := DefaultConfig()
adapter := newMockAdapter()

migrator, err := NewMigrator(config, adapter)
if err != nil {
t.Fatalf("NewMigrator() error = %v", err)
}

customLogger := NewNoopLogger()
migrator.SetLogger(customLogger)

// Just verify it doesn't crash
}

func TestMigrator_Migrate_WithCallbacks(t *testing.T) {
ClearRegisteredMigrations()
defer ClearRegisteredMigrations()

RegisterMigration(NewBaseMigration("20241230100000", "test", 
func(adapter DatabaseAdapter) error { return nil }, nil))

config := DefaultConfig()
config.DatabaseURL = "postgres://localhost/test"
adapter := newMockAdapter()

migrator, err := NewMigrator(config, adapter)
if err != nil {
t.Fatalf("NewMigrator() error = %v", err)
}

beforeCalled := false
afterCalled := false

migrator.SetBeforeMigrate(func(migration Migration, direction string) error {
beforeCalled = true
if direction != "up" {
t.Errorf("Expected direction 'up', got %s", direction)
}
return nil
})

migrator.SetAfterMigrate(func(migration Migration, direction string) error {
afterCalled = true
return nil
})

ctx := context.Background()
if err := migrator.Migrate(ctx); err != nil {
t.Fatalf("Migrate() error = %v", err)
}

if !beforeCalled {
t.Error("Expected before callback to be called")
}

if !afterCalled {
t.Error("Expected after callback to be called")
}
}

func TestMigrator_MigrateDown_WithCallbacks(t *testing.T) {
ClearRegisteredMigrations()
defer ClearRegisteredMigrations()

RegisterMigration(NewBaseMigration("20241230100000", "test", 
nil,
func(adapter DatabaseAdapter) error { return nil }))

config := DefaultConfig()
adapter := newMockAdapter()
adapter.appliedMigrations = []MigrationRecord{
{Version: "20241230100000", Batch: 1, ExecutedAt: time.Now()},
}
adapter.lastBatch = 1

migrator, err := NewMigrator(config, adapter)
if err != nil {
t.Fatalf("NewMigrator() error = %v", err)
}

beforeCalled := false
migrator.SetBeforeMigrate(func(migration Migration, direction string) error {
beforeCalled = true
if direction != "down" {
t.Errorf("Expected direction 'down', got %s", direction)
}
return nil
})

ctx := context.Background()
if err := migrator.MigrateDown(ctx, 1); err != nil {
t.Fatalf("MigrateDown() error = %v", err)
}

if !beforeCalled {
t.Error("Expected before callback to be called during rollback")
}
}

func TestMigrator_Status_Empty(t *testing.T) {
ClearRegisteredMigrations()
defer ClearRegisteredMigrations()

config := DefaultConfig()
adapter := newMockAdapter()

migrator, err := NewMigrator(config, adapter)
if err != nil {
t.Fatalf("NewMigrator() error = %v", err)
}

ctx := context.Background()
_, err = migrator.Status(ctx)

// Should return error for no migrations
if err == nil {
t.Error("Expected error for no migrations")
}
}

func TestMigrator_Rollback_NoMigrations(t *testing.T) {
ClearRegisteredMigrations()
defer ClearRegisteredMigrations()

config := DefaultConfig()
adapter := newMockAdapter()

migrator, err := NewMigrator(config, adapter)
if err != nil {
t.Fatalf("NewMigrator() error = %v", err)
}

ctx := context.Background()
err = migrator.Rollback(ctx)

if err == nil {
t.Error("Expected error when no migrations to rollback")
}
}

func TestMigrator_MigrateUp_ZeroSteps(t *testing.T) {
ClearRegisteredMigrations()
defer ClearRegisteredMigrations()

RegisterMigration(NewBaseMigration("20241230100000", "test", 
func(adapter DatabaseAdapter) error { return nil }, nil))

config := DefaultConfig()
adapter := newMockAdapter()

migrator, err := NewMigrator(config, adapter)
if err != nil {
t.Fatalf("NewMigrator() error = %v", err)
}

ctx := context.Background()
err = migrator.MigrateUp(ctx, 0)

// Should handle gracefully
if err != nil {
t.Errorf("Unexpected error for zero steps: %v", err)
}
}
