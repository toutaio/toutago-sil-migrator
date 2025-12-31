package sil

import (
	"testing"
	"time"
)

func TestMigrationRecord_Fields(t *testing.T) {
	record := MigrationRecord{
		Version:     "20240101000000",
		Description: "create users",
		Batch:       1,
		ExecutedAt:  time.Now(),
	}

	if record.Version != "20240101000000" {
		t.Error("Expected version to match")
	}

	if record.Batch != 1 {
		t.Errorf("Expected batch 1, got %d", record.Batch)
	}
}

func TestMigrationStatus_Fields(t *testing.T) {
	now := time.Now()
	status := MigrationStatus{
		Version:     "20240101000000",
		Description: "create users",
		Applied:     true,
		Batch:       1,
		ExecutedAt:  &now,
	}

	if status.Version != "20240101000000" {
		t.Errorf("Expected version to match")
	}

	if !status.Applied {
		t.Error("Expected applied to be true")
	}

	if status.Batch != 1 {
		t.Errorf("Expected batch 1, got %d", status.Batch)
	}
}

func TestConfig_MergeOverrides(t *testing.T) {
	base := DefaultConfig()
	base.DatabaseURL = "postgres://localhost/base"
	base.MigrationsDir = "./base_migrations"

	override := &Config{
		DatabaseURL: "postgres://localhost/override",
		TableName:   "custom_migrations",
	}

	result := base.Merge(override)

	if result.DatabaseURL != "postgres://localhost/override" {
		t.Errorf("Expected DatabaseURL to be overridden")
	}

	if result.MigrationsDir != "./base_migrations" {
		t.Errorf("Expected MigrationsDir to remain from base")
	}

	if result.TableName != "custom_migrations" {
		t.Errorf("Expected TableName to be overridden")
	}
}


func TestConfig_ValidateTableName(t *testing.T) {
	config := DefaultConfig()
	config.DatabaseURL = "postgres://localhost/test"
	config.TableName = ""

	err := config.Validate()
	if err == nil {
		t.Error("Expected error for empty table name")
	}
}

func TestConfig_ValidateMigrationsDir(t *testing.T) {
	config := DefaultConfig()
	config.DatabaseURL = "postgres://localhost/test"
	config.MigrationsDir = ""

	err := config.Validate()
	if err == nil {
		t.Error("Expected error for empty migrations directory")
	}
}

func TestDefaultConfig_Values(t *testing.T) {
	config := DefaultConfig()

	if config.MigrationsDir != "./migrations" {
		t.Errorf("Expected default migrations dir './migrations', got %s", config.MigrationsDir)
	}

	if config.SeedersDir != "./seeders" {
		t.Errorf("Expected default seeders dir './seeders', got %s", config.SeedersDir)
	}

	if config.TableName != "sil_migrations" {
		t.Errorf("Expected default table name 'sil_migrations', got %s", config.TableName)
	}

	if config.Environment != "development" {
		t.Errorf("Expected default environment 'development', got %s", config.Environment)
	}

	if config.LockTimeout != 5*time.Minute {
		t.Errorf("Expected default lock timeout 5m, got %v", config.LockTimeout)
	}
}

func TestSeedRecord_Fields(t *testing.T) {
	record := SeedRecord{
		Name:        "UserSeeder",
		Environment: "development",
		ExecutedAt:  time.Now(),
	}

	if record.Name != "UserSeeder" {
		t.Errorf("Expected name to match")
	}

	if record.Environment != "development" {
		t.Errorf("Expected environment to match")
	}
}

func TestSeedStatus_Fields(t *testing.T) {
	status := SeedStatus{
		Name:     "UserSeeder",
		Executed: true,
		Skipped:  false,
		Reason:   "",
	}

	if status.Name != "UserSeeder" {
		t.Errorf("Expected name to match")
	}

	if !status.Executed {
		t.Error("Expected executed to be true")
	}

	if status.Skipped {
		t.Error("Expected skipped to be false")
	}
}

func TestConfig_TimeoutFields(t *testing.T) {
	config := DefaultConfig()

	if config.LockTimeout <= 0 {
		t.Error("Expected positive lock timeout")
	}

	if config.MigrationTimeout <= 0 {
		t.Error("Expected positive migration timeout")
	}
}

func TestConfig_BooleanFlags(t *testing.T) {
	config := DefaultConfig()

	// Defaults should be false
	if config.Verbose {
		t.Error("Expected verbose to default to false")
	}

	// Test setting
	config.Verbose = true
	if !config.Verbose {
		t.Error("Expected verbose to be true after setting")
	}
}

func TestConfig_String(t *testing.T) {
config := DefaultConfig()
config.DatabaseURL = "postgres://user:pass@localhost/testdb"

str := config.String()

if str == "" {
t.Error("Expected non-empty string representation")
}

// Should contain key information
if !stringContains(str, "DatabaseURL") {
t.Error("Expected string to contain DatabaseURL")
}
}

func TestConfig_MergeNilOverride(t *testing.T) {
base := DefaultConfig()
base.DatabaseURL = "postgres://localhost/base"

result := base.Merge(nil)

if result.DatabaseURL != base.DatabaseURL {
t.Error("Expected database URL to remain unchanged with nil override")
}
}

func TestConfig_MergeWithOverrides(t *testing.T) {
base := DefaultConfig()
base.DatabaseURL = "postgres://localhost/base"
base.MigrationsDir = "./migrations"
base.Verbose = false

override := &Config{
DatabaseURL: "postgres://localhost/override",
Verbose:     true,
}

result := base.Merge(override)

if result.DatabaseURL != override.DatabaseURL {
t.Errorf("Expected database URL %s, got %s", override.DatabaseURL, result.DatabaseURL)
}

if result.Verbose != override.Verbose {
t.Error("Expected verbose to be overridden")
}

// Non-overridden field should keep base value
if result.MigrationsDir != base.MigrationsDir {
t.Error("Expected migrations dir to keep base value")
}
}

func TestConfig_MergeAllFields(t *testing.T) {
base := DefaultConfig()

override := &Config{
DatabaseURL:      "postgres://new/db",
MigrationsDir:    "./new-migrations",
SeedersDir:       "./new-seeders",
TableName:        "new_migrations",
LockTimeout:      60,
MigrationTimeout: 120,
Verbose:          true,
Environment:      "production",
}

result := base.Merge(override)

if result.DatabaseURL != override.DatabaseURL {
t.Error("DatabaseURL not merged correctly")
}
if result.MigrationsDir != override.MigrationsDir {
t.Error("MigrationsDir not merged correctly")
}
if result.SeedersDir != override.SeedersDir {
t.Error("SeedersDir not merged correctly")
}
if result.TableName != override.TableName {
t.Error("TableName not merged correctly")
}
if result.LockTimeout != override.LockTimeout {
t.Error("LockTimeout not merged correctly")
}
if result.MigrationTimeout != override.MigrationTimeout {
t.Error("MigrationTimeout not merged correctly")
}
if result.Verbose != override.Verbose {
t.Error("Verbose not merged correctly")
}
if result.Environment != override.Environment {
t.Error("Environment not merged correctly")
}
}

func TestConfig_ValidateErrors(t *testing.T) {
tests := []struct {
name        string
config      *Config
wantErr     bool
errContains string
}{
{
name: "empty database URL",
config: &Config{
DatabaseURL: "",
},
wantErr:     true,
errContains: "database_url is required",
},
{
name: "valid config",
config: func() *Config {
c := DefaultConfig()
c.DatabaseURL = "postgres://localhost/db"
return c
}(),
wantErr: false,
},
}

for _, tt := range tests {
t.Run(tt.name, func(t *testing.T) {
err := tt.config.Validate()

if tt.wantErr && err == nil {
t.Error("Expected validation error, got nil")
}

if !tt.wantErr && err != nil {
t.Errorf("Expected no error, got %v", err)
}

if tt.wantErr && tt.errContains != "" {
if err == nil || !stringContains(err.Error(), tt.errContains) {
t.Errorf("Expected error containing %q, got %v", tt.errContains, err)
}
}
})
}
}


// Helper function
func stringContains(s, substr string) bool {
return len(s) >= len(substr) && findInString(s, substr)
}

func findInString(s, substr string) bool {
if s == substr {
return true
}
for i := 0; i <= len(s)-len(substr); i++ {
if s[i:i+len(substr)] == substr {
return true
}
}
return false
}
