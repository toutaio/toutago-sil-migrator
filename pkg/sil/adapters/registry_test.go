package adapters

import (
	// imported for potential use
	"testing"

	"github.com/toutaio/toutago-sil-migrator/pkg/sil"
)

func TestDetectAdapterType(t *testing.T) {
	tests := []struct {
		name        string
		databaseURL string
		want        AdapterType
		wantErr     bool
	}{
		{"PostgreSQL", "postgres://localhost/db", AdapterPostgreSQL, false},
		{"MySQL", "mysql://localhost/db", AdapterMySQL, false},
		{"SQLite", "sqlite://./db.sqlite", AdapterSQLite, false},
		{"Empty URL", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DetectAdapterType(tt.databaseURL)
			if (err != nil) != tt.wantErr {
				t.Errorf("DetectAdapterType() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("DetectAdapterType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewAdapter(t *testing.T) {
	config := &sil.Config{
		DatabaseURL: "postgres://localhost/test",
		TableName:   "migrations",
	}

	adapter, err := NewAdapter(config)
	if err != nil {
		t.Fatalf("NewAdapter() error = %v", err)
	}

	if adapter == nil {
		t.Error("NewAdapter() returned nil")
	}
}

func TestSupportedAdapters(t *testing.T) {
	adapters := SupportedAdapters()
	if len(adapters) != 3 {
		t.Errorf("Expected 3 adapters, got %d", len(adapters))
	}
}

func TestIsAdapterSupported(t *testing.T) {
	if !IsAdapterSupported(AdapterPostgreSQL) {
		t.Error("PostgreSQL should be supported")
	}

	if IsAdapterSupported(AdapterType("oracle")) {
		t.Error("Oracle should not be supported")
	}
}
