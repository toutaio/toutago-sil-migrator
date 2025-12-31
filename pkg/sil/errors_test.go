package sil

import (
	"errors"
	"testing"
	"time"
)

func TestLockErrorFormatting(t *testing.T) {
	err := &LockError{
		Reason:  "timeout",
		LockID:  "migration_lock",
		Timeout: 30 * time.Second,
	}

	errStr := err.Error()
	if errStr == "" {
		t.Error("Error string should not be empty")
	}
}

func TestSeederErrorFormatting(t *testing.T) {
	err := &SeederError{
		Seeder: "test_seeder",
		Err:    errors.New("failed"),
	}

	errStr := err.Error()
	if errStr == "" {
		t.Error("Error string should not be empty")
	}
}

func TestConfigurationError_Error(t *testing.T) {
	err := &ConfigurationError{
		Field:   "DatabaseURL",
		Message: "invalid URL format",
	}

	expected := "configuration error: DatabaseURL: invalid URL format"
	if err.Error() != expected {
		t.Errorf("Error() = %v, want %v", err.Error(), expected)
	}
}

func TestConfigurationError_Error_NoField(t *testing.T) {
	err := &ConfigurationError{
		Message: "invalid configuration",
	}

	expected := "configuration error: invalid configuration"
	if err.Error() != expected {
		t.Errorf("Error() = %v, want %v", err.Error(), expected)
	}
}

func TestValidationError_Error(t *testing.T) {
	err := &ValidationError{
		Message: "invalid migration name",
	}

	expected := "validation error: invalid migration name"
	if err.Error() != expected {
		t.Errorf("Error() = %v, want %v", err.Error(), expected)
	}
}

func TestValidationError_Error_WithField(t *testing.T) {
	err := &ValidationError{
		Field:   "Name",
		Message: "cannot be empty",
	}

	expected := "validation error: Name: cannot be empty"
	if err.Error() != expected {
		t.Errorf("Error() = %v, want %v", err.Error(), expected)
	}
}

func TestErrInvalidConfigurationField(t *testing.T) {
	err := ErrInvalidConfigurationField("DatabaseURL", "cannot be empty")

	if err == nil {
		t.Error("ErrInvalidConfigurationField() returned nil")
	}

	expected := "configuration error: DatabaseURL: cannot be empty"
	if err.Error() != expected {
		t.Errorf("Error() = %v, want %v", err.Error(), expected)
	}
}

func TestMigrationError_Error(t *testing.T) {
	err := &MigrationError{
		Migration: "20240101_create_users",
		Operation: "up",
		Err:       errors.New("table exists"),
	}

	expected := "migration 20240101_create_users failed during up: table exists"
	if err.Error() != expected {
		t.Errorf("Error() = %v, want %v", err.Error(), expected)
	}
}

func TestMigrationError_Unwrap(t *testing.T) {
	innerErr := errors.New("SQL error")
	err := &MigrationError{
		Migration: "20240101_create_users",
		Operation: "up",
		Err:       innerErr,
	}

	unwrapped := err.Unwrap()
	if unwrapped != innerErr {
		t.Errorf("Unwrap() = %v, want %v", unwrapped, innerErr)
	}
}

func TestSeederError_Error(t *testing.T) {
	err := &SeederError{
		Seeder: "UserSeeder",
		Err:    errors.New("duplicate key"),
	}

	expected := "seeder UserSeeder failed: duplicate key"
	if err.Error() != expected {
		t.Errorf("Error() = %v, want %v", err.Error(), expected)
	}
}

func TestSeederError_Unwrap(t *testing.T) {
	innerErr := errors.New("duplicate key")
	err := &SeederError{
		Seeder: "UserSeeder",
		Err:    innerErr,
	}

	unwrapped := err.Unwrap()
	if unwrapped != innerErr {
		t.Errorf("Unwrap() = %v, want %v", unwrapped, innerErr)
	}
}

func TestWrapMigrationError(t *testing.T) {
	innerErr := errors.New("table already exists")
	err := WrapMigrationError("20240101_create_users", "up", innerErr)

	if err == nil {
		t.Error("WrapMigrationError() returned nil")
	}

	expected := "migration 20240101_create_users failed during up: table already exists"
	if err.Error() != expected {
		t.Errorf("Error() = %v, want %v", err.Error(), expected)
	}

	migErr, ok := err.(*MigrationError)
	if !ok {
		t.Error("WrapMigrationError() did not return *MigrationError")
	}

	if migErr.Unwrap() != innerErr {
		t.Error("WrapMigrationError() did not wrap inner error correctly")
	}
}

func TestWrapSeederError(t *testing.T) {
	innerErr := errors.New("duplicate key violation")
	err := WrapSeederError("UserSeeder", innerErr)

	if err == nil {
		t.Error("WrapSeederError() returned nil")
	}

	expected := "seeder UserSeeder failed: duplicate key violation"
	if err.Error() != expected {
		t.Errorf("Error() = %v, want %v", err.Error(), expected)
	}

	seederErr, ok := err.(*SeederError)
	if !ok {
		t.Error("WrapSeederError() did not return *SeederError")
	}

	if seederErr.Err != innerErr {
		t.Error("WrapSeederError() did not wrap inner error correctly")
	}
}

func TestIsMigrationError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "migration error",
			err:  &MigrationError{Migration: "test", Operation: "up", Err: errors.New("error")},
			want: true,
		},
		{
			name: "wrapped migration error",
			err:  WrapMigrationError("test", "up", errors.New("inner")),
			want: true,
		},
		{
			name: "other error",
			err:  errors.New("generic error"),
			want: false,
		},
		{
			name: "nil error",
			err:  nil,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsMigrationError(tt.err)
			if got != tt.want {
				t.Errorf("IsMigrationError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsSeederError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "seeder error",
			err:  &SeederError{Seeder: "test", Err: errors.New("error")},
			want: true,
		},
		{
			name: "wrapped seeder error",
			err:  WrapSeederError("test", errors.New("inner")),
			want: true,
		},
		{
			name: "other error",
			err:  errors.New("generic error"),
			want: false,
		},
		{
			name: "nil error",
			err:  nil,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsSeederError(tt.err)
			if got != tt.want {
				t.Errorf("IsSeederError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsConfigurationError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "configuration error",
			err:  &ConfigurationError{Field: "test", Message: "error"},
			want: true,
		},
		{
			name: "other error",
			err:  errors.New("generic error"),
			want: false,
		},
		{
			name: "nil error",
			err:  nil,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsConfigurationError(tt.err)
			if got != tt.want {
				t.Errorf("IsConfigurationError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsLockError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "lock error",
			err:  &LockError{Reason: "timeout"},
			want: true,
		},
		{
			name: "lock timeout error",
			err:  ErrLockTimeout,
			want: false,
		},
		{
			name: "other error",
			err:  errors.New("generic error"),
			want: false,
		},
		{
			name: "nil error",
			err:  nil,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsLockError(tt.err)
			if got != tt.want {
				t.Errorf("IsLockError() = %v, want %v", got, tt.want)
			}
		})
	}
}
