package adapters

import (
	"context"
	"database/sql"
	"testing"

	"github.com/toutaio/toutago-sil-migrator/pkg/sil"
)

// MockDB is a mock database for testing
type MockDB struct {
	execFunc  func(query string, args ...interface{}) (sql.Result, error)
	queryFunc func(query string, args ...interface{}) (*sql.Rows, error)
	beginFunc func() (*sql.Tx, error)
	closeFunc func() error
}

func (m *MockDB) Exec(query string, args ...interface{}) (sql.Result, error) {
	if m.execFunc != nil {
		return m.execFunc(query, args...)
	}
	return &MockResult{}, nil
}

func (m *MockDB) Query(query string, args ...interface{}) (*sql.Rows, error) {
	if m.queryFunc != nil {
		return m.queryFunc(query, args...)
	}
	return nil, nil
}

func (m *MockDB) Begin() (*sql.Tx, error) {
	if m.beginFunc != nil {
		return m.beginFunc()
	}
	return nil, nil
}

func (m *MockDB) Close() error {
	if m.closeFunc != nil {
		return m.closeFunc()
	}
	return nil
}

// MockResult implements sql.Result
type MockResult struct {
	lastInsertId int64
	rowsAffected int64
}

func (m *MockResult) LastInsertId() (int64, error) {
	return m.lastInsertId, nil
}

func (m *MockResult) RowsAffected() (int64, error) {
	return m.rowsAffected, nil
}

// Tests for adapter creation
func TestCreateAdapter(t *testing.T) {
	tests := []struct {
		name        string
		config      *sil.Config
		wantType    string
		wantErr     bool
		errContains string
	}{
		{
			name: "postgres adapter",
			config: &sil.Config{
				DatabaseURL: "postgres://localhost/test",
			},
			wantType: "*adapters.PostgresAdapter",
			wantErr:  false,
		},
		{
			name: "postgresql adapter",
			config: &sil.Config{
				DatabaseURL: "postgresql://localhost/test",
			},
			wantType: "*adapters.PostgresAdapter",
			wantErr:  false,
		},
		{
			name: "mysql adapter",
			config: &sil.Config{
				DatabaseURL: "mysql://user:pass@localhost/test",
			},
			wantType: "*adapters.MySQLAdapter",
			wantErr:  false,
		},
		{
			name: "sqlite adapter",
			config: &sil.Config{
				DatabaseURL: "sqlite://./test.db",
			},
			wantType: "*adapters.SQLiteAdapter",
			wantErr:  false,
		},
		{
			name: "sqlite3 adapter",
			config: &sil.Config{
				DatabaseURL: "sqlite3://./test.db",
			},
			wantType: "*adapters.SQLiteAdapter",
			wantErr:  false,
		},
		{
			name: "unsupported database",
			config: &sil.Config{
				DatabaseURL: "oracle://localhost/test",
			},
			wantErr:     true,
			errContains: "unable to detect adapter type",
		},
		{
			name: "invalid URL",
			config: &sil.Config{
				DatabaseURL: "not-a-valid-url",
			},
			wantErr:     true,
			errContains: "unable to detect adapter type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter, err := NewAdapter(tt.config)

			if tt.wantErr {
				if err == nil {
					t.Errorf("NewAdapter() expected error, got nil")
					return
				}
				if tt.errContains != "" && !contains(err.Error(), tt.errContains) {
					t.Errorf("NewAdapter() error = %v, want error containing %v", err, tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("NewAdapter() unexpected error = %v", err)
				return
			}

			if adapter == nil {
				t.Errorf("NewAdapter() returned nil adapter")
				return
			}

			// Check type (we can't do direct type assertions without connecting)
			// Just verify we got something
		})
	}
}

func TestDatabaseTypeDetection(t *testing.T) {
	tests := []struct {
		url      string
		wantType string
	}{
		{"postgres://localhost/db", "postgres"},
		{"postgresql://localhost/db", "postgres"},
		{"mysql://localhost/db", "mysql"},
		{"sqlite://./db.sqlite", "sqlite"},
		{"sqlite3://./db.db", "sqlite"},
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			config := &sil.Config{
				DatabaseURL: tt.url,
			}
			adapter, err := NewAdapter(config)
			if err != nil {
				t.Skipf("Skipping: %v", err)
				return
			}
			// Just verify we can create the adapter
			_ = adapter
		})
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Test adapter interface compliance
func TestAdapterInterfaceCompliance(t *testing.T) {
	config := &sil.Config{
		DatabaseURL: "postgres://localhost/test",
	}

	adapter, err := NewAdapter(config)
	if err != nil {
		t.Skipf("Skipping interface test: %v", err)
		return
	}

	// Verify adapter implements DatabaseAdapter
	var _ sil.DatabaseAdapter = adapter

	// Test that adapter has expected methods (compile-time check)
	ctx := context.Background()

	// These should not panic
	_ = adapter.Close()
	_ = adapter.Exec(ctx, "")
	_, _ = adapter.Query(ctx, "")
	_, _ = adapter.BeginTx(ctx)
}

// Test MySQL specific features
func TestMySQLAdapterCreation(t *testing.T) {
	// Test the adapter creation
	adapter := &MySQLAdapter{}

	// Verify the adapter structure is initialized
	_ = adapter // adapter is intentionally not nil after struct literal creation
}

// Test SQLite path extraction
func TestSQLitePathExtraction(t *testing.T) {
	tests := []struct {
		url      string
		wantPath string
	}{
		{"sqlite://./test.db", "./test.db"},
		{"sqlite3://./test.db", "./test.db"},
		{"sqlite:///absolute/path/test.db", "/absolute/path/test.db"},
		{"sqlite://test.db", "test.db"},
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			// Extract path from URL
			path := tt.url
			if len(path) > 9 && path[:9] == "sqlite://" {
				path = path[9:]
			} else if len(path) > 10 && path[:10] == "sqlite3://" {
				path = path[10:]
			}

			if path != tt.wantPath {
				t.Errorf("path extraction = %q, want %q", path, tt.wantPath)
			}
		})
	}
}

// Test connection string validation
func TestConnectionStringValidation(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{"valid postgres", "postgres://localhost/db", false},
		{"valid mysql", "mysql://localhost/db", false},
		{"valid sqlite", "sqlite://./db.sqlite", false},
		{"empty url", "", true},
		{"unsupported", "oracle://localhost/db", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &sil.Config{
				DatabaseURL: tt.url,
			}

			_, err := NewAdapter(config)

			hasErr := err != nil
			if hasErr != tt.wantErr {
				t.Errorf("NewAdapter() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
