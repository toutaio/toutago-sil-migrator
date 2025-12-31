package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/toutaio/toutago-sil-migrator/pkg/sil"
	"github.com/toutaio/toutago-sil-migrator/pkg/sil/adapters"
)

// loadConfig loads the configuration from file, environment, or defaults.
func loadConfig() (*sil.Config, error) {
	var config *sil.Config
	var err error

	// Try to load from config file if specified
	if configFile != "" {
		config, err = sil.LoadConfig(configFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load config: %w", err)
		}
	} else {
		// Try to auto-find config file
		configPath, err := sil.FindConfigFile()
		if err == nil {
			config, err = sil.LoadConfig(configPath)
			if err != nil {
				return nil, fmt.Errorf("failed to load config: %w", err)
			}
		} else {
			// Use defaults
			config = sil.DefaultConfig()
		}
	}

	// Override with environment variables
	config = sil.LoadConfigFromEnv(config)

	// Validate
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return config, nil
}

// createAdapter creates a database adapter based on the database URL.
func createAdapter(config *sil.Config) (sil.DatabaseAdapter, error) {
	// Determine database type from URL
	dbType := detectDatabaseType(config.DatabaseURL)

	switch dbType {
	case "postgres", "postgresql":
		return adapters.NewPostgresAdapter(config)
	default:
		return nil, fmt.Errorf("unsupported database type: %s (only PostgreSQL is currently supported)", dbType)
	}
}

// detectDatabaseType detects the database type from the connection URL.
func detectDatabaseType(url string) string {
	if strings.HasPrefix(url, "postgres://") || strings.HasPrefix(url, "postgresql://") {
		return "postgres"
	}
	if strings.HasPrefix(url, "mysql://") {
		return "mysql"
	}
	if strings.HasPrefix(url, "sqlite://") || strings.HasSuffix(url, ".db") {
		return "sqlite"
	}
	return "unknown"
}

// printSuccess prints a success message in green.
func printSuccess(format string, args ...interface{}) {
	const colorGreen = "\033[32m"
	const colorReset = "\033[0m"

	if isColorSupported() {
		fmt.Printf("%s✓ "+format+"%s\n", colorGreen, args, colorReset)
	} else {
		fmt.Printf("✓ "+format+"\n", args...)
	}
}

// isColorSupported checks if the terminal supports color output.
func isColorSupported() bool {
	// Check if NO_COLOR environment variable is set
	if os.Getenv("NO_COLOR") != "" {
		return false
	}

	// Check if running in a terminal
	if os.Getenv("TERM") == "dumb" {
		return false
	}

	// Assume color is supported on Unix-like systems
	return true
}

func printBanner(title string) {
	width := 60
	fmt.Println()
	fmt.Println(strings.Repeat("=", width))
	padding := (width - len(title) - 2) / 2
	fmt.Printf("%s %s %s\n", strings.Repeat("=", padding), title, strings.Repeat("=", padding))
	fmt.Println(strings.Repeat("=", width))
}

func toSnakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune('_')
		}
		result.WriteRune(r)
	}
	return strings.ToLower(result.String())
}

func toPascalCase(s string) string {
	parts := strings.Split(s, "_")
	for i, part := range parts {
		if part != "" {
			parts[i] = strings.ToUpper(part[:1]) + part[1:]
		}
	}
	return strings.Join(parts, "")
}
