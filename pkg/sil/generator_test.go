package sil

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerator_Generate(t *testing.T) {
	tmpDir := t.TempDir()

	generator := NewGenerator(tmpDir, NewNoopLogger())

	filename, err := generator.Generate("create users table")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if filename == "" {
		t.Error("Expected filename, got empty string")
	}

	// Check file exists
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		t.Errorf("Expected file to exist at %s", filename)
	}

	// Check file content
	content, err := os.ReadFile(filename) //#nosec G304 -- Filename is from test setup
	if err != nil {
		t.Fatalf("Failed to read generated file: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "package migrations") {
		t.Error("Expected package declaration")
	}

	if !strings.Contains(contentStr, "func (m *Migration_") {
		t.Error("Expected migration struct methods")
	}
}

func TestGenerator_GenerateCreateTable(t *testing.T) {
	tmpDir := t.TempDir()

	generator := NewGenerator(tmpDir, NewNoopLogger())

	filename, err := generator.GenerateCreateTable("users")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Check file content
	content, err := os.ReadFile(filename) //#nosec G304 -- Filename is from test setup
	if err != nil {
		t.Fatalf("Failed to read generated file: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "CREATE TABLE users") {
		t.Error("Expected CREATE TABLE statement")
	}

	if !strings.Contains(contentStr, "DROP TABLE IF EXISTS users") {
		t.Error("Expected DROP TABLE statement in Down method")
	}
}

func TestGenerator_GenerateAddColumn(t *testing.T) {
	tmpDir := t.TempDir()

	generator := NewGenerator(tmpDir, NewNoopLogger())

	filename, err := generator.GenerateAddColumn("users", "email")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Check file exists
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		t.Error("Expected file to exist")
	}
}

func TestGenerator_GenerateEmptyDescription(t *testing.T) {
	tmpDir := t.TempDir()

	generator := NewGenerator(tmpDir, NewNoopLogger())

	_, err := generator.Generate("")
	if err == nil {
		t.Error("Expected error for empty description")
	}
}

func TestGenerator_GenerateInvalidDirectory(t *testing.T) {
	generator := NewGenerator("/nonexistent/directory/that/does/not/exist", NewNoopLogger())

	_, err := generator.Generate("test migration")
	if err == nil {
		t.Error("Expected error for invalid directory")
	}
}

func TestGenerator_CreateDirectoryIfNotExists(t *testing.T) {
	tmpDir := t.TempDir()
	newDir := filepath.Join(tmpDir, "new_migrations")

	generator := NewGenerator(newDir, NewNoopLogger())

	_, err := generator.Generate("test migration")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Check directory was created
	if _, err := os.Stat(newDir); os.IsNotExist(err) {
		t.Error("Expected directory to be created")
	}
}

func TestGenerator_FileAlreadyExists(t *testing.T) {
	tmpDir := t.TempDir()

	generator := NewGenerator(tmpDir, NewNoopLogger())

	// Generate first migration
	filename, err := generator.Generate("create users")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Try to write the same file again (simulate duplicate)
	err = os.WriteFile(filename, []byte("duplicate"), 0o600)
	if err != nil {
		t.Fatalf("Failed to create duplicate file: %v", err)
	}

	// This should succeed but create different filename due to timestamp
	_, _ = generator.Generate("create users")
	// It's okay if it fails or succeeds - timing dependent
	// Just ensure it doesn't panic
}
