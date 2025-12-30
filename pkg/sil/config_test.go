package sil

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.MigrationsDir != "./migrations" {
		t.Errorf("DefaultConfig().MigrationsDir = %s, want ./migrations", config.MigrationsDir)
	}

	if config.Environment != "development" {
		t.Errorf("DefaultConfig().Environment = %s, want development", config.Environment)
	}

	if config.LockTimeout != 5*time.Minute {
		t.Errorf("DefaultConfig().LockTimeout = %v, want 5m", config.LockTimeout)
	}

	if config.TableName != "sil_migrations" {
		t.Errorf("DefaultConfig().TableName = %s, want sil_migrations", config.TableName)
	}
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: func() *Config {
				c := DefaultConfig()
				c.DatabaseURL = "postgres://localhost/test"
				return c
			}(),
			wantErr: false,
		},
		{
			name: "missing database URL",
			config: &Config{
				MigrationsDir: "./migrations",
				TableName:     "sil_migrations",
			},
			wantErr: true,
		},
		{
			name: "missing migrations dir",
			config: &Config{
				DatabaseURL: "postgres://localhost/test",
				TableName:   "sil_migrations",
			},
			wantErr: true,
		},
		{
			name: "invalid lock timeout",
			config: &Config{
				DatabaseURL:   "postgres://localhost/test",
				MigrationsDir: "./migrations",
				LockTimeout:   0,
				TableName:     "sil_migrations",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConfig_Merge(t *testing.T) {
	base := DefaultConfig()
	base.DatabaseURL = "postgres://localhost/base"

	override := &Config{
		DatabaseURL: "postgres://localhost/override",
		Environment: "production",
	}

	result := base.Merge(override)

	if result.DatabaseURL != "postgres://localhost/override" {
		t.Errorf("Merge() DatabaseURL = %s, want postgres://localhost/override", result.DatabaseURL)
	}

	if result.Environment != "production" {
		t.Errorf("Merge() Environment = %s, want production", result.Environment)
	}

	// Check that default values are preserved
	if result.MigrationsDir != "./migrations" {
		t.Errorf("Merge() MigrationsDir = %s, want ./migrations", result.MigrationsDir)
	}
}

func TestSaveAndLoadConfig(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()

	config := DefaultConfig()
	config.DatabaseURL = "postgres://localhost/test"
	config.Environment = "test"

	// Test YAML
	yamlPath := filepath.Join(tmpDir, "test.yaml")
	if err := SaveConfig(config, yamlPath); err != nil {
		t.Fatalf("SaveConfig() error = %v", err)
	}

	loaded, err := LoadConfig(yamlPath)
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	if loaded.DatabaseURL != config.DatabaseURL {
		t.Errorf("LoadConfig() DatabaseURL = %s, want %s", loaded.DatabaseURL, config.DatabaseURL)
	}

	// Test JSON
	jsonPath := filepath.Join(tmpDir, "test.json")
	if err := SaveConfig(config, jsonPath); err != nil {
		t.Fatalf("SaveConfig() JSON error = %v", err)
	}

	loadedJSON, err := LoadConfig(jsonPath)
	if err != nil {
		t.Fatalf("LoadConfig() JSON error = %v", err)
	}

	if loadedJSON.DatabaseURL != config.DatabaseURL {
		t.Errorf("LoadConfig() JSON DatabaseURL = %s, want %s", loadedJSON.DatabaseURL, config.DatabaseURL)
	}
}

func TestLoadConfigFromEnv(t *testing.T) {
	// Save original env vars
	originalURL := os.Getenv("SIL_DATABASE_URL")
	originalEnv := os.Getenv("SIL_ENVIRONMENT")

	// Set test env vars
	os.Setenv("SIL_DATABASE_URL", "postgres://localhost/envtest")
	os.Setenv("SIL_ENVIRONMENT", "testing")
	os.Setenv("SIL_VERBOSE", "true")

	defer func() {
		// Restore original env vars
		os.Setenv("SIL_DATABASE_URL", originalURL)
		os.Setenv("SIL_ENVIRONMENT", originalEnv)
		os.Unsetenv("SIL_VERBOSE")
	}()

	config := DefaultConfig()
	config = LoadConfigFromEnv(config)

	if config.DatabaseURL != "postgres://localhost/envtest" {
		t.Errorf("LoadConfigFromEnv() DatabaseURL = %s, want postgres://localhost/envtest", config.DatabaseURL)
	}

	if config.Environment != "testing" {
		t.Errorf("LoadConfigFromEnv() Environment = %s, want testing", config.Environment)
	}

	if !config.Verbose {
		t.Error("LoadConfigFromEnv() Verbose = false, want true")
	}
}

func TestGenerateMigrationFileName(t *testing.T) {
	tests := []struct {
		name        string
		description string
		wantPattern string
	}{
		{
			name:        "simple description",
			description: "create users table",
			wantPattern: "_create_users_table.go",
		},
		{
			name:        "with special characters",
			description: "add email-column to users",
			wantPattern: "_add_emailcolumn_to_users.go",
		},
		{
			name:        "uppercase",
			description: "CREATE USERS",
			wantPattern: "_create_users.go",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filename := GenerateMigrationFileName(tt.description)

			// Check it ends with .go
			if filepath.Ext(filename) != ".go" {
				t.Errorf("GenerateMigrationFileName() = %s, should end with .go", filename)
			}

			// Check it contains the pattern
			if !contains(filename, tt.wantPattern) {
				t.Errorf("GenerateMigrationFileName() = %s, should contain %s", filename, tt.wantPattern)
			}

			// Check version part is valid
			version := filename[:14]
			if err := ValidateVersion(version); err != nil {
				t.Errorf("GenerateMigrationFileName() version %s is invalid: %v", version, err)
			}
		})
	}
}

func TestParseMigrationFileName(t *testing.T) {
	tests := []struct {
		name            string
		filename        string
		wantVersion     string
		wantDescription string
		wantErr         bool
	}{
		{
			name:            "valid filename",
			filename:        "20241230100000_create_users_table.go",
			wantVersion:     "20241230100000",
			wantDescription: "create users table",
			wantErr:         false,
		},
		{
			name:            "with path",
			filename:        "/path/to/20241230100000_create_users.go",
			wantVersion:     "20241230100000",
			wantDescription: "create users",
			wantErr:         false,
		},
		{
			name:     "invalid format - no underscore",
			filename: "20241230100000.go",
			wantErr:  true,
		},
		{
			name:     "invalid format - short version",
			filename: "202412301_create_users.go",
			wantErr:  true,
		},
		{
			name:     "not a go file",
			filename: "20241230100000_create_users.txt",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			version, description, err := ParseMigrationFileName(tt.filename)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseMigrationFileName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if version != tt.wantVersion {
					t.Errorf("ParseMigrationFileName() version = %s, want %s", version, tt.wantVersion)
				}
				if description != tt.wantDescription {
					t.Errorf("ParseMigrationFileName() description = %s, want %s", description, tt.wantDescription)
				}
			}
		})
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[len(s)-len(substr):] == substr ||
		len(s) > len(substr) && findSubstring(s, substr)
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
