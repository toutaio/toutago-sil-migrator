package sil

import (
	"testing"
	"time"
)

func TestValidateVersion(t *testing.T) {
	tests := []struct {
		name    string
		version string
		wantErr bool
	}{
		{
			name:    "valid version",
			version: "20241230010000",
			wantErr: false,
		},
		{
			name:    "valid version with different time",
			version: "20240101120000",
			wantErr: false,
		},
		{
			name:    "too short",
			version: "2024123001",
			wantErr: true,
		},
		{
			name:    "too long",
			version: "202412300100000",
			wantErr: true,
		},
		{
			name:    "invalid characters",
			version: "20241230ABCD00",
			wantErr: true,
		},
		{
			name:    "invalid date",
			version: "20241332010000",
			wantErr: true,
		},
		{
			name:    "empty",
			version: "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateVersion(tt.version)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateVersion() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGenerateVersion(t *testing.T) {
	version := GenerateVersion()

	// Should be 14 digits
	if len(version) != 14 {
		t.Errorf("GenerateVersion() returned %s with length %d, want 14", version, len(version))
	}

	// Should be valid
	if err := ValidateVersion(version); err != nil {
		t.Errorf("GenerateVersion() returned invalid version: %v", err)
	}

	// Should be close to current time
	now := time.Now()
	expected := now.Format("20060102150405")

	// Allow 1 second difference
	if version[:12] != expected[:12] {
		t.Errorf("GenerateVersion() = %s, expected to start with %s", version, expected[:12])
	}
}

func TestSortMigrations(t *testing.T) {
	migrations := []Migration{
		NewBaseMigration("20241230120000", "third", nil, nil),
		NewBaseMigration("20241230100000", "first", nil, nil),
		NewBaseMigration("20241230110000", "second", nil, nil),
	}

	SortMigrations(migrations)

	expected := []string{"20241230100000", "20241230110000", "20241230120000"}
	for i, m := range migrations {
		if m.Version() != expected[i] {
			t.Errorf("SortMigrations() at index %d = %s, want %s", i, m.Version(), expected[i])
		}
	}
}

func TestSortMigrationsDescending(t *testing.T) {
	migrations := []Migration{
		NewBaseMigration("20241230100000", "first", nil, nil),
		NewBaseMigration("20241230120000", "third", nil, nil),
		NewBaseMigration("20241230110000", "second", nil, nil),
	}

	SortMigrationsDescending(migrations)

	expected := []string{"20241230120000", "20241230110000", "20241230100000"}
	for i, m := range migrations {
		if m.Version() != expected[i] {
			t.Errorf("SortMigrationsDescending() at index %d = %s, want %s", i, m.Version(), expected[i])
		}
	}
}

func TestFilterPendingMigrations(t *testing.T) {
	all := []Migration{
		NewBaseMigration("20241230100000", "first", nil, nil),
		NewBaseMigration("20241230110000", "second", nil, nil),
		NewBaseMigration("20241230120000", "third", nil, nil),
	}

	applied := []MigrationRecord{
		{Version: "20241230100000", Batch: 1},
	}

	pending := FilterPendingMigrations(all, applied)

	if len(pending) != 2 {
		t.Errorf("FilterPendingMigrations() returned %d migrations, want 2", len(pending))
	}

	expectedVersions := map[string]bool{
		"20241230110000": true,
		"20241230120000": true,
	}

	for _, m := range pending {
		if !expectedVersions[m.Version()] {
			t.Errorf("FilterPendingMigrations() included unexpected version %s", m.Version())
		}
	}
}

func TestFilterAppliedMigrations(t *testing.T) {
	all := []Migration{
		NewBaseMigration("20241230100000", "first", nil, nil),
		NewBaseMigration("20241230110000", "second", nil, nil),
		NewBaseMigration("20241230120000", "third", nil, nil),
	}

	applied := []MigrationRecord{
		{Version: "20241230100000", Batch: 1},
		{Version: "20241230120000", Batch: 2},
	}

	appliedMigrations := FilterAppliedMigrations(all, applied)

	if len(appliedMigrations) != 2 {
		t.Errorf("FilterAppliedMigrations() returned %d migrations, want 2", len(appliedMigrations))
	}

	expectedVersions := map[string]bool{
		"20241230100000": true,
		"20241230120000": true,
	}

	for _, m := range appliedMigrations {
		if !expectedVersions[m.Version()] {
			t.Errorf("FilterAppliedMigrations() included unexpected version %s", m.Version())
		}
	}
}

func TestGetMigrationsByBatch(t *testing.T) {
	all := []Migration{
		NewBaseMigration("20241230100000", "first", nil, nil),
		NewBaseMigration("20241230110000", "second", nil, nil),
		NewBaseMigration("20241230120000", "third", nil, nil),
	}

	applied := []MigrationRecord{
		{Version: "20241230100000", Batch: 1},
		{Version: "20241230110000", Batch: 1},
		{Version: "20241230120000", Batch: 2},
	}

	batch1 := GetMigrationsByBatch(all, applied, 1)
	if len(batch1) != 2 {
		t.Errorf("GetMigrationsByBatch(1) returned %d migrations, want 2", len(batch1))
	}

	batch2 := GetMigrationsByBatch(all, applied, 2)
	if len(batch2) != 1 {
		t.Errorf("GetMigrationsByBatch(2) returned %d migrations, want 1", len(batch2))
	}

	batch3 := GetMigrationsByBatch(all, applied, 3)
	if len(batch3) != 0 {
		t.Errorf("GetMigrationsByBatch(3) returned %d migrations, want 0", len(batch3))
	}
}

func TestGetMigrationStatus(t *testing.T) {
	now := time.Now()
	all := []Migration{
		NewBaseMigration("20241230100000", "first", nil, nil),
		NewBaseMigration("20241230110000", "second", nil, nil),
		NewBaseMigration("20241230120000", "third", nil, nil),
	}

	applied := []MigrationRecord{
		{Version: "20241230100000", Batch: 1, ExecutedAt: now},
		{Version: "20241230120000", Batch: 2, ExecutedAt: now},
	}

	statuses := GetMigrationStatus(all, applied)

	if len(statuses) != 3 {
		t.Errorf("GetMigrationStatus() returned %d statuses, want 3", len(statuses))
	}

	// Check first migration (applied)
	if !statuses[0].Applied {
		t.Error("Expected first migration to be applied")
	}
	if statuses[0].Batch != 1 {
		t.Errorf("Expected first migration batch = 1, got %d", statuses[0].Batch)
	}

	// Check second migration (not applied)
	if statuses[1].Applied {
		t.Error("Expected second migration to not be applied")
	}

	// Check third migration (applied)
	if !statuses[2].Applied {
		t.Error("Expected third migration to be applied")
	}
	if statuses[2].Batch != 2 {
		t.Errorf("Expected third migration batch = 2, got %d", statuses[2].Batch)
	}
}

func TestBaseMigration(t *testing.T) {
	upCalled := false
	downCalled := false

	upFunc := func(adapter DatabaseAdapter) error {
		upCalled = true
		return nil
	}

	downFunc := func(adapter DatabaseAdapter) error {
		downCalled = true
		return nil
	}

	migration := NewBaseMigration("20241230100000", "test migration", upFunc, downFunc)

	// Test Version()
	if migration.Version() != "20241230100000" {
		t.Errorf("Version() = %s, want 20241230100000", migration.Version())
	}

	// Test Description()
	if migration.Description() != "test migration" {
		t.Errorf("Description() = %s, want 'test migration'", migration.Description())
	}

	// Test Up()
	if err := migration.Up(nil); err != nil {
		t.Errorf("Up() returned error: %v", err)
	}
	if !upCalled {
		t.Error("Up() did not call upFunc")
	}

	// Test Down()
	if err := migration.Down(nil); err != nil {
		t.Errorf("Down() returned error: %v", err)
	}
	if !downCalled {
		t.Error("Down() did not call downFunc")
	}
}

func TestRegisterMigration(t *testing.T) {
	// Clear registry before test
	ClearRegisteredMigrations()

	migration := NewBaseMigration("20241230100000", "test", nil, nil)
	RegisterMigration(migration)

	migrations := GetRegisteredMigrations()
	if len(migrations) != 1 {
		t.Errorf("GetRegisteredMigrations() returned %d migrations, want 1", len(migrations))
	}

	if migrations[0].Version() != "20241230100000" {
		t.Errorf("Registered migration version = %s, want 20241230100000", migrations[0].Version())
	}

	// Clean up
	ClearRegisteredMigrations()
}

func TestRegisterMigration_Duplicate(t *testing.T) {
	// Clear registry before test
	ClearRegisteredMigrations()

	defer func() {
		if r := recover(); r == nil {
			t.Error("RegisterMigration() should panic on duplicate version")
		}
		ClearRegisteredMigrations()
	}()

	migration1 := NewBaseMigration("20241230100000", "test1", nil, nil)
	migration2 := NewBaseMigration("20241230100000", "test2", nil, nil)

	RegisterMigration(migration1)
	RegisterMigration(migration2) // Should panic
}

func TestNewBaseMigration(t *testing.T) {
upFunc := func(adapter DatabaseAdapter) error { return nil }
downFunc := func(adapter DatabaseAdapter) error { return nil }

migration := NewBaseMigration("20240101000000", "create users", upFunc, downFunc)

if migration.Version() != "20240101000000" {
t.Errorf("Expected version 20240101000000, got %s", migration.Version())
}

if migration.Description() != "create users" {
t.Errorf("Expected description 'create users', got %s", migration.Description())
}
}

func TestBaseMigration_Up(t *testing.T) {
called := false
upFunc := func(adapter DatabaseAdapter) error {
called = true
return nil
}

migration := NewBaseMigration("20240101000000", "test", upFunc, nil)

err := migration.Up(nil)
if err != nil {
t.Errorf("Expected no error, got %v", err)
}

if !called {
t.Error("Expected Up function to be called")
}
}

func TestBaseMigration_Down(t *testing.T) {
called := false
downFunc := func(adapter DatabaseAdapter) error {
called = true
return nil
}

migration := NewBaseMigration("20240101000000", "test", nil, downFunc)

err := migration.Down(nil)
if err != nil {
t.Errorf("Expected no error, got %v", err)
}

if !called {
t.Error("Expected Down function to be called")
}
}

func TestBaseMigration_UpNilFunc(t *testing.T) {
migration := NewBaseMigration("20240101000000", "test", nil, nil)

err := migration.Up(nil)
if err == nil {
t.Error("Expected error for nil Up function", err)
}
}

func TestBaseMigration_DownNilFunc(t *testing.T) {
migration := NewBaseMigration("20240101000000", "test", nil, nil)

err := migration.Down(nil)
if err == nil {
t.Errorf("Expected no error for nil Down function, got %v", err)
}
}

func TestClearRegisteredMigrations(t *testing.T) {
ClearRegisteredMigrations()

m1 := NewBaseMigration("20240101000000", "first", nil, nil)
m2 := NewBaseMigration("20240102000000", "second", nil, nil)
RegisterMigration(m1)
RegisterMigration(m2)

migrations := GetRegisteredMigrations()
if len(migrations) != 2 {
t.Errorf("Expected 2 registered migrations, got %d", len(migrations))
}

ClearRegisteredMigrations()

migrations = GetRegisteredMigrations()
if len(migrations) != 0 {
t.Errorf("Expected 0 migrations after clear, got %d", len(migrations))
}
}

func TestGetMigrationStatus_Multiple(t *testing.T) {
ClearRegisteredMigrations()
defer ClearRegisteredMigrations()

m1 := NewBaseMigration("20240101000000", "first", nil, nil)
m2 := NewBaseMigration("20240102000000", "second", nil, nil)
m3 := NewBaseMigration("20240103000000", "third", nil, nil)
RegisterMigration(m1)
RegisterMigration(m2)
RegisterMigration(m3)

applied := []MigrationRecord{
{Version: "20240101000000", Batch: 1},
{Version: "20240102000000", Batch: 1},
}

statuses := GetMigrationStatus(GetRegisteredMigrations(), applied)

if len(statuses) != 3 {
t.Errorf("Expected 3 statuses, got %d", len(statuses))
}

if !statuses[0].Applied {
t.Error("Expected first migration to be applied")
}

if !statuses[1].Applied {
t.Error("Expected second migration to be applied")
}

if statuses[2].Applied {
t.Error("Expected third migration to not be applied")
}
}

func TestValidateVersion_ValidFormats(t *testing.T) {
valid := []string{
"20240101000000",
"20991231235959",
"20240630120000",
}

for _, version := range valid {
t.Run(version, func(t *testing.T) {
err := ValidateVersion(version)
if err != nil {
t.Errorf("Expected %s to be valid, got error: %v", version, err)
}
})
}
}

func TestValidateVersion_InvalidFormats(t *testing.T) {
invalid := []string{
"",
"123",
"2024010100000",  // Too short
"202401010000000", // Too long
"abcd0101000000",  // Non-numeric
"20241301000000",  // Invalid month
"20240132000000",  // Invalid day
"20240101250000",  // Invalid hour
"20240101006000",  // Invalid minute
"20240101000060",  // Invalid second
}

for _, version := range invalid {
t.Run(version, func(t *testing.T) {
err := ValidateVersion(version)
if err == nil {
t.Errorf("Expected %s to be invalid", version)
}
})
}
}

func TestFilterPendingMigrations_Edge(t *testing.T) {
// Test with empty lists
pending := FilterPendingMigrations([]Migration{}, []MigrationRecord{})
if len(pending) != 0 {
t.Error("Expected empty result for empty inputs")
}
}

func TestFilterAppliedMigrations_Edge(t *testing.T) {
// Test with empty lists
applied := FilterAppliedMigrations([]Migration{}, []MigrationRecord{})
if len(applied) != 0 {
t.Error("Expected empty result for empty inputs")
}
}
