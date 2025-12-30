package sil

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// LoadConfig loads configuration from a file (YAML or JSON).
func LoadConfig(path string) (*Config, error) {
	if path == "" {
		return nil, ErrInvalidConfiguration("config path is empty")
	}

	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file not found: %s", path)
	}

	// Read file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse based on file extension
	ext := strings.ToLower(filepath.Ext(path))
	var config Config

	switch ext {
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(data, &config); err != nil {
			return nil, fmt.Errorf("failed to parse YAML config: %w", err)
		}
	case ".json":
		if err := json.Unmarshal(data, &config); err != nil {
			return nil, fmt.Errorf("failed to parse JSON config: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported config file format: %s (use .yaml, .yml, or .json)", ext)
	}

	return &config, nil
}

// LoadConfigWithDefaults loads configuration from a file and merges with defaults.
func LoadConfigWithDefaults(path string) (*Config, error) {
	defaults := DefaultConfig()

	if path == "" {
		return defaults, nil
	}

	config, err := LoadConfig(path)
	if err != nil {
		return nil, err
	}

	return defaults.Merge(config), nil
}

// LoadConfigFromEnv loads configuration from environment variables.
// Environment variables override config file values.
func LoadConfigFromEnv(config *Config) *Config {
	if config == nil {
		config = DefaultConfig()
	}

	// Database URL
	if v := os.Getenv("SIL_DATABASE_URL"); v != "" {
		config.DatabaseURL = v
	}

	// Migrations directory
	if v := os.Getenv("SIL_MIGRATIONS_DIR"); v != "" {
		config.MigrationsDir = v
	}

	// Seeders directory
	if v := os.Getenv("SIL_SEEDERS_DIR"); v != "" {
		config.SeedersDir = v
	}

	// Lock timeout
	if v := os.Getenv("SIL_LOCK_TIMEOUT"); v != "" {
		if duration, err := time.ParseDuration(v); err == nil {
			config.LockTimeout = duration
		}
	}

	// Migration timeout
	if v := os.Getenv("SIL_MIGRATION_TIMEOUT"); v != "" {
		if duration, err := time.ParseDuration(v); err == nil {
			config.MigrationTimeout = duration
		}
	}

	// Max connections
	if v := os.Getenv("SIL_MAX_CONNECTIONS"); v != "" {
		if num, err := strconv.Atoi(v); err == nil && num > 0 {
			config.MaxConnections = num
		}
	}

	// Max idle connections
	if v := os.Getenv("SIL_MAX_IDLE_CONNECTIONS"); v != "" {
		if num, err := strconv.Atoi(v); err == nil && num > 0 {
			config.MaxIdleConnections = num
		}
	}

	// Connection max lifetime
	if v := os.Getenv("SIL_CONNECTION_MAX_LIFETIME"); v != "" {
		if duration, err := time.ParseDuration(v); err == nil {
			config.ConnectionMaxLifetime = duration
		}
	}

	// Environment
	if v := os.Getenv("SIL_ENVIRONMENT"); v != "" {
		config.Environment = v
	}

	// Table name
	if v := os.Getenv("SIL_TABLE_NAME"); v != "" {
		config.TableName = v
	}

	// Seeds table name
	if v := os.Getenv("SIL_SEEDS_TABLE_NAME"); v != "" {
		config.SeedsTableName = v
	}

	// Verbose
	if v := os.Getenv("SIL_VERBOSE"); v != "" {
		config.Verbose = v == "true" || v == "1" || v == "yes"
	}

	return config
}

// FindConfigFile searches for a config file in common locations.
// It looks for sil.yaml, sil.yml, or sil.json in the current directory
// and parent directories up to the root.
func FindConfigFile() (string, error) {
	configNames := []string{"sil.yaml", "sil.yml", "sil.json"}

	// Start from current directory
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get working directory: %w", err)
	}

	// Search up to root
	for {
		for _, name := range configNames {
			path := filepath.Join(dir, name)
			if _, err := os.Stat(path); err == nil {
				return path, nil
			}
		}

		// Move to parent directory
		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached root
			break
		}
		dir = parent
	}

	return "", fmt.Errorf("no config file found (looking for: %s)", strings.Join(configNames, ", "))
}

// LoadConfigAuto automatically finds and loads configuration.
// It searches for config files, loads environment variables, and applies defaults.
func LoadConfigAuto() (*Config, error) {
	// Start with defaults
	config := DefaultConfig()

	// Try to find config file
	configPath, err := FindConfigFile()
	if err == nil {
		// Config file found, load it
		fileConfig, err := LoadConfig(configPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load config file %s: %w", configPath, err)
		}
		config = config.Merge(fileConfig)
	}
	// If no config file found, continue with defaults

	// Override with environment variables
	config = LoadConfigFromEnv(config)

	// Validate final configuration
	if err := config.Validate(); err != nil {
		return nil, err
	}

	return config, nil
}

// SaveConfig saves configuration to a file.
func SaveConfig(config *Config, path string) error {
	if config == nil {
		return ErrInvalidConfiguration("config is nil")
	}

	if path == "" {
		return ErrInvalidConfiguration("config path is empty")
	}

	// Determine format from extension
	ext := strings.ToLower(filepath.Ext(path))

	var data []byte
	var err error

	switch ext {
	case ".yaml", ".yml":
		data, err = yaml.Marshal(config)
		if err != nil {
			return fmt.Errorf("failed to marshal config to YAML: %w", err)
		}
	case ".json":
		data, err = json.MarshalIndent(config, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal config to JSON: %w", err)
		}
	default:
		return fmt.Errorf("unsupported config file format: %s (use .yaml, .yml, or .json)", ext)
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Write file
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// InitConfig creates a new config file with defaults.
func InitConfig(path string) error {
	// Check if file already exists
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("config file already exists: %s", path)
	}

	config := DefaultConfig()
	return SaveConfig(config, path)
}
