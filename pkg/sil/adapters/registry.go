package adapters

import (
	"fmt"
	"strings"

	"github.com/toutaio/toutago-sil-migrator/pkg/sil"
)

// AdapterType represents the type of database adapter.
type AdapterType string

const (
	AdapterPostgreSQL AdapterType = "postgres"
	AdapterMySQL      AdapterType = "mysql"
	AdapterSQLite     AdapterType = "sqlite"
)

// DetectAdapterType detects the adapter type from a database URL.
func DetectAdapterType(databaseURL string) (AdapterType, error) {
	if databaseURL == "" {
		return "", fmt.Errorf("database URL is empty")
	}

	// Check for common prefixes
	switch {
	case strings.HasPrefix(databaseURL, "postgres://"),
		strings.HasPrefix(databaseURL, "postgresql://"):
		return AdapterPostgreSQL, nil

	case strings.HasPrefix(databaseURL, "mysql://"),
		strings.Contains(databaseURL, "@tcp("):
		return AdapterMySQL, nil

	case strings.HasPrefix(databaseURL, "sqlite://"),
		strings.HasPrefix(databaseURL, "file://"),
		strings.HasSuffix(databaseURL, ".db"),
		strings.HasSuffix(databaseURL, ".sqlite"),
		strings.HasSuffix(databaseURL, ".sqlite3"):
		return AdapterSQLite, nil

	default:
		return "", fmt.Errorf("unable to detect adapter type from URL: %s", databaseURL)
	}
}

// NewAdapter creates a new database adapter based on the configuration.
// It automatically detects the adapter type from the database URL.
func NewAdapter(config *sil.Config) (sil.DatabaseAdapter, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	adapterType, err := DetectAdapterType(config.DatabaseURL)
	if err != nil {
		return nil, err
	}

	return NewAdapterOfType(adapterType, config)
}

// NewAdapterOfType creates a new database adapter of the specified type.
func NewAdapterOfType(adapterType AdapterType, config *sil.Config) (sil.DatabaseAdapter, error) {
	switch adapterType {
	case AdapterPostgreSQL:
		return NewPostgresAdapter(config)

	case AdapterMySQL:
		return NewMySQLAdapter(config)

	case AdapterSQLite:
		return NewSQLiteAdapter(config)

	default:
		return nil, fmt.Errorf("unsupported adapter type: %s", adapterType)
	}
}

// SupportedAdapters returns a list of supported database adapters.
func SupportedAdapters() []AdapterType {
	return []AdapterType{
		AdapterPostgreSQL,
		AdapterMySQL,
		AdapterSQLite,
	}
}

// IsAdapterSupported checks if an adapter type is supported.
func IsAdapterSupported(adapterType AdapterType) bool {
	for _, supported := range SupportedAdapters() {
		if supported == adapterType {
			return true
		}
	}
	return false
}
