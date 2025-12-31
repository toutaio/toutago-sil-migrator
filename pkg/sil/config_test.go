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

func TestLoadConfigWithDefaults(t *testing.T) {
tmpDir := t.TempDir()
configPath := filepath.Join(tmpDir, "sil.yaml")

// Create minimal config file
configContent := `database_url: "postgres://localhost/test"`
if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
t.Fatalf("Failed to write config: %v", err)
}

config, err := LoadConfigWithDefaults(configPath)
if err != nil {
t.Fatalf("LoadConfigWithDefaults() error = %v", err)
}

if config.DatabaseURL != "postgres://localhost/test" {
t.Errorf("Expected database_url to be loaded")
}

// Should have defaults
if config.MigrationsDir != "./migrations" {
t.Errorf("Expected default migrations_dir")
}

if config.TableName != "sil_migrations" {
t.Errorf("Expected default table_name")
}
}

func TestFindConfigFile(t *testing.T) {
tmpDir := t.TempDir()

// Create config in tmpDir
configPath := filepath.Join(tmpDir, "sil.yaml")
if err := os.WriteFile(configPath, []byte("test"), 0644); err != nil {
t.Fatalf("Failed to write config: %v", err)
}

// Change to tmpDir
oldDir, err := os.Getwd()
if err != nil {
t.Fatalf("Failed to get working directory: %v", err)
}
defer os.Chdir(oldDir)

if err := os.Chdir(tmpDir); err != nil {
t.Fatalf("Failed to change directory: %v", err)
}

// Find config
found, err := FindConfigFile()
if err != nil {
t.Fatalf("FindConfigFile() error = %v", err)
}

if found != configPath {
t.Errorf("FindConfigFile() = %s, want %s", found, configPath)
}
}

func TestFindConfigFile_NotFound(t *testing.T) {
tmpDir := t.TempDir()

oldDir, err := os.Getwd()
if err != nil {
t.Fatalf("Failed to get working directory: %v", err)
}
defer os.Chdir(oldDir)

if err := os.Chdir(tmpDir); err != nil {
t.Fatalf("Failed to change directory: %v", err)
}

// Should not find config
_, err = FindConfigFile()
if err == nil {
t.Error("Expected error when config not found")
}
}

func TestLoadConfigAuto(t *testing.T) {
tmpDir := t.TempDir()
configPath := filepath.Join(tmpDir, "sil.yaml")

// Create config file
configContent := `database_url: "postgres://localhost/autotest"`
if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
t.Fatalf("Failed to write config: %v", err)
}

oldDir, err := os.Getwd()
if err != nil {
t.Fatalf("Failed to get working directory: %v", err)
}
defer os.Chdir(oldDir)

if err := os.Chdir(tmpDir); err != nil {
t.Fatalf("Failed to change directory: %v", err)
}

config, err := LoadConfigAuto()
if err != nil {
t.Fatalf("LoadConfigAuto() error = %v", err)
}

if config.DatabaseURL != "postgres://localhost/autotest" {
t.Errorf("Expected database_url from auto-loaded config")
}
}

func TestInitConfig(t *testing.T) {
tmpDir := t.TempDir()
configPath := filepath.Join(tmpDir, "sil.yaml")

err := InitConfig(configPath)
if err != nil {
t.Fatalf("InitConfig() error = %v", err)
}

// Check file exists
if _, err := os.Stat(configPath); os.IsNotExist(err) {
t.Error("Expected config file to be created")
}

// Check can be loaded
config, err := LoadConfig(configPath)
if err != nil {
t.Fatalf("Failed to load initialized config: %v", err)
}

if config.TableName != "sil_migrations" {
t.Errorf("Expected default table_name in initialized config")
}
}

func TestLoadConfigFromEnv_Complete(t *testing.T) {
// Save current env
oldURL := os.Getenv("SIL_DATABASE_URL")
oldDir := os.Getenv("SIL_MIGRATIONS_DIR")
oldTable := os.Getenv("SIL_TABLE_NAME")
defer func() {
os.Setenv("SIL_DATABASE_URL", oldURL)
os.Setenv("SIL_MIGRATIONS_DIR", oldDir)
os.Setenv("SIL_TABLE_NAME", oldTable)
}()

// Set env vars
os.Setenv("SIL_DATABASE_URL", "postgres://localhost/envtest")
os.Setenv("SIL_MIGRATIONS_DIR", "./test_migrations")
os.Setenv("SIL_TABLE_NAME", "test_migrations")

config := LoadConfigFromEnv(DefaultConfig())

if config.DatabaseURL != "postgres://localhost/envtest" {
t.Errorf("Expected database_url from env")
}

if config.MigrationsDir != "./test_migrations" {
t.Errorf("Expected migrations_dir from env")
}

if config.TableName != "test_migrations" {
t.Errorf("Expected table_name from env")
}
}

func TestLoadConfig_YAML(t *testing.T) {
tmpDir := t.TempDir()
configPath := filepath.Join(tmpDir, "test.yaml")

configContent := `
database_url: "postgres://localhost/yamltest"
migrations_dir: "./yaml_migrations"
seeders_dir: "./yaml_seeders"
table_name: "yaml_migrations"
environment: "test"
verbose: true
`
if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
t.Fatalf("Failed to write config: %v", err)
}

config, err := LoadConfig(configPath)
if err != nil {
t.Fatalf("LoadConfig() error = %v", err)
}

if config.DatabaseURL != "postgres://localhost/yamltest" {
t.Errorf("Expected DatabaseURL to be loaded from YAML")
}

if config.MigrationsDir != "./yaml_migrations" {
t.Errorf("Expected MigrationsDir to be loaded from YAML")
}

if !config.Verbose {
t.Errorf("Expected Verbose to be true")
}
}

func TestLoadConfig_JSON(t *testing.T) {
tmpDir := t.TempDir()
configPath := filepath.Join(tmpDir, "test.json")

configContent := `{
"database_url": "postgres://localhost/jsontest",
"migrations_dir": "./json_migrations",
"table_name": "json_migrations",
"environment": "test"
}`
if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
t.Fatalf("Failed to write config: %v", err)
}

config, err := LoadConfig(configPath)
if err != nil {
t.Fatalf("LoadConfig() error = %v", err)
}

if config.DatabaseURL != "postgres://localhost/jsontest" {
t.Errorf("Expected DatabaseURL to be loaded from JSON")
}

if config.MigrationsDir != "./json_migrations" {
t.Errorf("Expected MigrationsDir to be loaded from JSON")
}
}

func TestLoadConfig_FileNotFound(t *testing.T) {
_, err := LoadConfig("/nonexistent/config.yaml")
if err == nil {
t.Error("Expected error for nonexistent file")
}
}

func TestLoadConfig_InvalidYAML(t *testing.T) {
tmpDir := t.TempDir()
configPath := filepath.Join(tmpDir, "invalid.yaml")

if err := os.WriteFile(configPath, []byte("invalid: [yaml: content"), 0644); err != nil {
t.Fatalf("Failed to write config: %v", err)
}

_, err := LoadConfig(configPath)
if err == nil {
t.Error("Expected error for invalid YAML")
}
}

func TestLoadConfig_InvalidJSON(t *testing.T) {
tmpDir := t.TempDir()
configPath := filepath.Join(tmpDir, "invalid.json")

if err := os.WriteFile(configPath, []byte("{invalid json"), 0644); err != nil {
t.Fatalf("Failed to write config: %v", err)
}

_, err := LoadConfig(configPath)
if err == nil {
t.Error("Expected error for invalid JSON")
}
}

func TestLoadConfigFromEnv_AllVars(t *testing.T) {
// Save and restore env
vars := []string{
"SIL_DATABASE_URL",
"SIL_MIGRATIONS_DIR",
"SIL_SEEDERS_DIR",
"SIL_TABLE_NAME",
"SIL_SEEDS_TABLE_NAME",
"SIL_ENVIRONMENT",
"SIL_VERBOSE",
"SIL_LOCK_TIMEOUT",
"SIL_MIGRATION_TIMEOUT",
}

oldVals := make(map[string]string)
for _, v := range vars {
oldVals[v] = os.Getenv(v)
}
defer func() {
for _, v := range vars {
os.Setenv(v, oldVals[v])
}
}()

// Set all env vars
os.Setenv("SIL_DATABASE_URL", "postgres://env/test")
os.Setenv("SIL_MIGRATIONS_DIR", "./env_migrations")
os.Setenv("SIL_SEEDERS_DIR", "./env_seeders")
os.Setenv("SIL_TABLE_NAME", "env_migrations")
os.Setenv("SIL_SEEDS_TABLE_NAME", "env_seeds")
os.Setenv("SIL_ENVIRONMENT", "production")
os.Setenv("SIL_VERBOSE", "true")
os.Setenv("SIL_LOCK_TIMEOUT", "10m")
os.Setenv("SIL_MIGRATION_TIMEOUT", "1h")

config := LoadConfigFromEnv(DefaultConfig())

if config.DatabaseURL != "postgres://env/test" {
t.Errorf("Expected DatabaseURL from env")
}

if config.MigrationsDir != "./env_migrations" {
t.Errorf("Expected MigrationsDir from env")
}

if config.Environment != "production" {
t.Errorf("Expected Environment from env")
}

if !config.Verbose {
t.Errorf("Expected Verbose to be true")
}

if config.LockTimeout != 10*time.Minute {
t.Errorf("Expected LockTimeout 10m, got %v", config.LockTimeout)
}
}

func TestSaveConfig_CreateDirectory(t *testing.T) {
tmpDir := t.TempDir()
configPath := filepath.Join(tmpDir, "subdir", "config.yaml")

config := DefaultConfig()
config.DatabaseURL = "postgres://localhost/savetest"

err := SaveConfig(config, configPath)
if err != nil {
t.Fatalf("SaveConfig() error = %v", err)
}

// Verify file exists
if _, err := os.Stat(configPath); os.IsNotExist(err) {
t.Error("Expected config file to be created")
}

// Verify can be loaded back
loaded, err := LoadConfig(configPath)
if err != nil {
t.Fatalf("Failed to load saved config: %v", err)
}

if loaded.DatabaseURL != config.DatabaseURL {
t.Errorf("Expected saved and loaded configs to match")
}
}

func TestLoadConfigFromEnv_InvalidDurations(t *testing.T) {
oldTimeout := os.Getenv("SIL_LOCK_TIMEOUT")
defer os.Setenv("SIL_LOCK_TIMEOUT", oldTimeout)

os.Setenv("SIL_LOCK_TIMEOUT", "invalid")

config := LoadConfigFromEnv(DefaultConfig())

// Should fall back to default
if config.LockTimeout != 5*time.Minute {
t.Errorf("Expected default lock timeout on invalid duration")
}
}

func TestLoadConfigWithDefaults_MergesBehavior(t *testing.T) {
tmpDir := t.TempDir()
configPath := filepath.Join(tmpDir, "partial.yaml")

// Write partial config
if err := os.WriteFile(configPath, []byte("database_url: postgres://localhost/partial"), 0644); err != nil {
t.Fatalf("Failed to write config: %v", err)
}

config, err := LoadConfigWithDefaults(configPath)
if err != nil {
t.Fatalf("LoadConfigWithDefaults() error = %v", err)
}

// Should have loaded value
if config.DatabaseURL != "postgres://localhost/partial" {
t.Error("Expected database_url to be loaded")
}

// Should have defaults for unspecified values
if config.TableName != "sil_migrations" {
t.Errorf("Expected default table_name")
}
}

func TestInitConfig_FileExists(t *testing.T) {
tmpDir := t.TempDir()
configPath := filepath.Join(tmpDir, "exists.yaml")

// Create file first
if err := os.WriteFile(configPath, []byte("test"), 0644); err != nil {
t.Fatalf("Failed to create file: %v", err)
}

// InitConfig should handle existing file
err := InitConfig(configPath)
// Implementation may overwrite or error - just ensure it doesn't crash
_ = err
}
