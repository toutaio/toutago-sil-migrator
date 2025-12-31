package sil

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoader_Load(t *testing.T) {
	// Clear and register test migrations
	ClearRegisteredMigrations()
	defer ClearRegisteredMigrations()

	m1 := NewBaseMigration("20240101000000", "first", nil, nil)
	m2 := NewBaseMigration("20240102000000", "second", nil, nil)
	RegisterMigration(m1)
	RegisterMigration(m2)

	loader := NewLoader("./migrations", NewNoopLogger())

	migrations, err := loader.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if len(migrations) != 2 {
		t.Errorf("Expected 2 migrations, got %d", len(migrations))
	}
}

func TestLoader_Load_Empty(t *testing.T) {
	// Clear registry
	ClearRegisteredMigrations()
	defer ClearRegisteredMigrations()

	loader := NewLoader("./migrations", NewNoopLogger())

	_, err := loader.Load()
	if err != ErrNoMigrationsFound {
		t.Errorf("Expected ErrNoMigrationsFound, got %v", err)
	}
}

func TestLoader_LoadPending(t *testing.T) {
	// Clear and register test migrations
	ClearRegisteredMigrations()
	defer ClearRegisteredMigrations()

	m1 := NewBaseMigration("20240101000000", "first", nil, nil)
	m2 := NewBaseMigration("20240102000000", "second", nil, nil)
	RegisterMigration(m1)
	RegisterMigration(m2)

	loader := NewLoader("./migrations", NewNoopLogger())

	applied := []MigrationRecord{
		{Version: "20240101000000", Batch: 1},
	}

	pending, err := loader.LoadPending(applied)
	if err != nil {
		t.Fatalf("LoadPending() error = %v", err)
	}

	if len(pending) != 1 {
		t.Errorf("Expected 1 pending migration, got %d", len(pending))
	}

	if pending[0].Version() != "20240102000000" {
		t.Errorf("Expected version 20240102000000, got %s", pending[0].Version())
	}
}

func TestLoader_LoadApplied(t *testing.T) {
	// Clear and register test migrations
	ClearRegisteredMigrations()
	defer ClearRegisteredMigrations()

	m1 := NewBaseMigration("20240101000000", "first", nil, nil)
	m2 := NewBaseMigration("20240102000000", "second", nil, nil)
	RegisterMigration(m1)
	RegisterMigration(m2)

	loader := NewLoader("./migrations", NewNoopLogger())

	records := []MigrationRecord{
		{Version: "20240101000000", Batch: 1},
	}

	applied, err := loader.LoadApplied(records)
	if err != nil {
		t.Fatalf("LoadApplied() error = %v", err)
	}

	if len(applied) != 1 {
		t.Errorf("Expected 1 applied migration, got %d", len(applied))
	}

	if applied[0].Version() != "20240101000000" {
		t.Errorf("Expected version 20240101000000, got %s", applied[0].Version())
	}
}

func TestDiscoverMigrationFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test files
	files := []string{
		"20240101000000_create_users.go",
		"20240102000000_create_posts.go",
		"README.md", // Should be ignored
		"helper.go", // Should be ignored
	}

	for _, file := range files {
		path := filepath.Join(tmpDir, file)
		if err := os.WriteFile(path, []byte("package migrations"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	loader := NewLoader(tmpDir, NewNoopLogger())
	discovered, err := loader.DiscoverMigrationFiles()
	if err != nil {
		t.Fatalf("DiscoverMigrationFiles() error = %v", err)
	}

	// Should include all .go files (migrations + helper)
	if len(discovered) != 3 {
		t.Errorf("Expected 3 .go files, got %d", len(discovered))
	}
}

func TestDiscoverMigrationFiles_NonexistentDir(t *testing.T) {
	loader := NewLoader("/nonexistent/directory", NewNoopLogger())
	_, err := loader.DiscoverMigrationFiles()
	if err == nil {
		t.Error("Expected error for nonexistent directory")
	}
}

func TestValidateMigrationFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create valid files
	files := map[string]string{
		"20240101000000_create_users.go": "package migrations",
		"20240102000000_create_posts.go": "package migrations",
	}

	for filename, content := range files {
		path := filepath.Join(tmpDir, filename)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	loader := NewLoader(tmpDir, NewNoopLogger())
	err := loader.ValidateMigrationFiles()
	if err != nil {
		t.Errorf("ValidateMigrationFiles() error = %v, want nil", err)
	}
}

func TestValidateMigrationFiles_Invalid(t *testing.T) {
	tmpDir := t.TempDir()

	// Create invalid file
	invalidFile := filepath.Join(tmpDir, "invalid_name.go")
	if err := os.WriteFile(invalidFile, []byte("package migrations"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	loader := NewLoader(tmpDir, NewNoopLogger())
	err := loader.ValidateMigrationFiles()
	if err == nil {
		t.Error("Expected error for invalid migration file")
	}
}

func TestCheckForOrphanedFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create migration file
	filename := "20240101000000_create_users.go"
	path := filepath.Join(tmpDir, filename)
	if err := os.WriteFile(path, []byte("package migrations"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// No migrations registered, so file is orphaned
	ClearRegisteredMigrations()
	defer ClearRegisteredMigrations()

	loader := NewLoader(tmpDir, NewNoopLogger())
	orphaned, err := loader.CheckForOrphanedFiles()
	if err != nil {
		t.Fatalf("CheckForOrphanedFiles() error = %v", err)
	}

	if len(orphaned) != 1 {
		t.Errorf("Expected 1 orphaned file, got %d", len(orphaned))
	}
}

func TestCheckForOrphanedFiles_NoOrphans(t *testing.T) {
	tmpDir := t.TempDir()

	// Create migration file
	filename := "20240101000000_create_users.go"
	path := filepath.Join(tmpDir, filename)
	if err := os.WriteFile(path, []byte("package migrations"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Register matching migration
	ClearRegisteredMigrations()
	defer ClearRegisteredMigrations()

	m := NewBaseMigration("20240101000000", "create users", nil, nil)
	RegisterMigration(m)

	loader := NewLoader(tmpDir, NewNoopLogger())
	orphaned, err := loader.CheckForOrphanedFiles()
	if err != nil {
		t.Fatalf("CheckForOrphanedFiles() error = %v", err)
	}

	if len(orphaned) != 0 {
		t.Errorf("Expected 0 orphaned files, got %d", len(orphaned))
	}
}

func TestLoader_LoadPending_AllApplied(t *testing.T) {
	ClearRegisteredMigrations()
	defer ClearRegisteredMigrations()

	m1 := NewBaseMigration("20240101000000", "first", nil, nil)
	RegisterMigration(m1)

	loader := NewLoader("./migrations", NewNoopLogger())

	// All migrations are applied
	applied := []MigrationRecord{
		{Version: "20240101000000", Batch: 1},
	}

	pending, err := loader.LoadPending(applied)
	if err != nil {
		t.Fatalf("LoadPending() error = %v", err)
	}

	if len(pending) != 0 {
		t.Errorf("Expected 0 pending migrations, got %d", len(pending))
	}
}

func TestLoader_GetMigrationPath(t *testing.T) {
	tmpDir := t.TempDir()

	// Create migration file
	filename := "20240101000000_create_users.go"
	path := filepath.Join(tmpDir, filename)
	if err := os.WriteFile(path, []byte("package migrations"), 0644); err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}

	loader := NewLoader(tmpDir, NewNoopLogger())

	foundPath, err := loader.GetMigrationPath("20240101000000")
	if err != nil {
		t.Fatalf("GetMigrationPath() error = %v", err)
	}

	if foundPath != path {
		t.Errorf("Expected path %s, got %s", path, foundPath)
	}
}

func TestLoader_GetMigrationPath_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	loader := NewLoader(tmpDir, NewNoopLogger())

	_, err := loader.GetMigrationPath("20240101000000")
	if err == nil {
		t.Error("Expected error for nonexistent migration")
	}
}

func TestParseMigrationFileName_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		filename    string
		wantVersion string
		wantDesc    string
		wantErr     bool
	}{
		{
			name:        "Standard format",
			filename:    "20240101000000_create_users.go",
			wantVersion: "20240101000000",
			wantDesc:    "create users",
			wantErr:     false,
		},
		{
			name:        "With underscores in description",
			filename:    "20240101000000_create_user_profiles.go",
			wantVersion: "20240101000000",
			wantDesc:    "create user profiles",
			wantErr:     false,
		},
		{
			name:     "Invalid format",
			filename: "invalid.go",
			wantErr:  true,
		},
		{
			name:     "Missing extension",
			filename: "20240101000000_create_users",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			version, desc, err := ParseMigrationFileName(tt.filename)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if version != tt.wantVersion {
				t.Errorf("Expected version %s, got %s", tt.wantVersion, version)
			}

			if desc != tt.wantDesc {
				t.Errorf("Expected description %s, got %s", tt.wantDesc, desc)
			}
		})
	}
}

func TestGenerateMigrationFileName_Consistency(t *testing.T) {
	desc := "create users"

	filename1 := GenerateMigrationFileName(desc)
	filename2 := GenerateMigrationFileName(desc)

	// Should generate different filenames due to timestamp
	if filename1 == filename2 {
		// This might happen if called in same microsecond, which is fine
	}

	// Both should be valid
	_, _, err1 := ParseMigrationFileName(filename1)
	_, _, err2 := ParseMigrationFileName(filename2)

	if err1 != nil {
		t.Errorf("Generated invalid filename: %s", filename1)
	}

	if err2 != nil {
		t.Errorf("Generated invalid filename: %s", filename2)
	}
}
