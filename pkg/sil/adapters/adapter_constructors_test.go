package adapters

import (
	"testing"

	"github.com/toutaio/toutago-sil-migrator/pkg/sil"
)

func TestNewMySQLAdapter(t *testing.T) {
	tests := []struct {
		name    string
		config  *sil.Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: &sil.Config{
				DatabaseURL: "mysql://localhost/test",
			},
			wantErr: false,
		},
		{
			name:    "nil config",
			config:  nil,
			wantErr: true,
		},
		{
			name:    "empty database URL",
			config:  &sil.Config{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter, err := NewMySQLAdapter(tt.config)

			if (err != nil) != tt.wantErr {
				t.Errorf("NewMySQLAdapter() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && adapter == nil {
				t.Error("NewMySQLAdapter() returned nil adapter")
			}
		})
	}
}

func TestNewPostgresAdapter(t *testing.T) {
	tests := []struct {
		name    string
		config  *sil.Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: &sil.Config{
				DatabaseURL: "postgres://localhost/test",
			},
			wantErr: false,
		},
		{
			name:    "nil config",
			config:  nil,
			wantErr: true,
		},
		{
			name: "empty database URL",
			config: &sil.Config{
				// Postgres adapter doesn't validate URL in constructor
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter, err := NewPostgresAdapter(tt.config)

			if (err != nil) != tt.wantErr {
				t.Errorf("NewPostgresAdapter() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && adapter == nil {
				t.Error("NewPostgresAdapter() returned nil adapter")
			}
		})
	}
}

func TestNewSQLiteAdapter(t *testing.T) {
	tests := []struct {
		name    string
		config  *sil.Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: &sil.Config{
				DatabaseURL: "sqlite://./test.db",
			},
			wantErr: false,
		},
		{
			name:    "nil config",
			config:  nil,
			wantErr: true,
		},
		{
			name:    "empty database URL",
			config:  &sil.Config{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter, err := NewSQLiteAdapter(tt.config)

			if (err != nil) != tt.wantErr {
				t.Errorf("NewSQLiteAdapter() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && adapter == nil {
				t.Error("NewSQLiteAdapter() returned nil adapter")
			}
		})
	}
}
