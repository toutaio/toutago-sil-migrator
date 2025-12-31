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
