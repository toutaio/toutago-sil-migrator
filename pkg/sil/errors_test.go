package sil

import (
	"errors"
	"testing"
)

func TestMigrationError_Error(t *testing.T) {
	err := &MigrationError{
		Migration: "20240101000000_create_users",
		Operation: "up",
		Err:       errors.New("syntax error"),
	}

	errMsg := err.Error()
	if errMsg == "" {
		t.Error("Expected non-empty error message")
	}

	if err.Unwrap() == nil {
		t.Error("Expected Unwrap to return cause")
	}
}

func TestConfigurationError_Error(t *testing.T) {
	err := &ConfigurationError{
		Field:   "DatabaseURL",
		Message: "is required",
	}

	errMsg := err.Error()
	if errMsg == "" {
		t.Error("Expected non-empty error message")
	}
}

func TestLockError_Error(t *testing.T) {
	err := &LockError{
		Reason: "timeout",
	}

	errMsg := err.Error()
	if errMsg == "" {
		t.Error("Expected non-empty error message")
	}
}

func TestSeederError_Error(t *testing.T) {
	err := &SeederError{
		Seeder: "UserSeeder",
		Err:    errors.New("database error"),
	}

	errMsg := err.Error()
	if errMsg == "" {
		t.Error("Expected non-empty error message")
	}

	if err.Unwrap() == nil {
		t.Error("Expected Unwrap to return cause")
	}
}

func TestValidationError_Error(t *testing.T) {
	err := &ValidationError{
		Field:   "Version",
		Message: "invalid format",
	}

	errMsg := err.Error()
	if errMsg == "" {
		t.Error("Expected non-empty error message")
	}
}

func TestWrapMigrationError(t *testing.T) {
	baseErr := errors.New("database error")
	wrapped := WrapMigrationError("20240101000000_create_users", "up", baseErr)

	if wrapped == nil {
		t.Fatal("Expected wrapped error")
	}

	migErr, ok := wrapped.(*MigrationError)
	if !ok {
		t.Fatal("Expected MigrationError type")
	}

	if migErr.Migration != "20240101000000_create_users" {
		t.Errorf("Expected migration to be set")
	}

	if migErr.Err != baseErr {
		t.Errorf("Expected err to be preserved")
	}
}

func TestWrapSeederError(t *testing.T) {
	baseErr := errors.New("seed failed")
	wrapped := WrapSeederError("UserSeeder", baseErr)

	if wrapped == nil {
		t.Fatal("Expected wrapped error")
	}

	seedErr, ok := wrapped.(*SeederError)
	if !ok {
		t.Fatal("Expected SeederError type")
	}

	if seedErr.Seeder != "UserSeeder" {
		t.Errorf("Expected seeder to be set")
	}

	if seedErr.Err != baseErr {
		t.Errorf("Expected err to be preserved")
	}
}

func TestIsMigrationError(t *testing.T) {
	migErr := &MigrationError{
		Migration: "20240101000000",
		Err:   errors.New("test"),
	}

	if !IsMigrationError(migErr) {
		t.Error("Expected IsMigrationError to return true")
	}

	if IsMigrationError(errors.New("normal error")) {
		t.Error("Expected IsMigrationError to return false for non-migration error")
	}
}

func TestIsSeederError(t *testing.T) {
	seedErr := &SeederError{
		Seeder:  "TestSeeder",
		Err: errors.New("test"),
	}

	if !IsSeederError(seedErr) {
		t.Error("Expected IsSeederError to return true")
	}

	if IsSeederError(errors.New("normal error")) {
		t.Error("Expected IsSeederError to return false for non-seeder error")
	}
}

func TestIsConfigurationError(t *testing.T) {
	confErr := &ConfigurationError{
		Field:   "DatabaseURL",
		Message: "required",
	}

	if !IsConfigurationError(confErr) {
		t.Error("Expected IsConfigurationError to return true")
	}

	if IsConfigurationError(errors.New("normal error")) {
		t.Error("Expected IsConfigurationError to return false")
	}
}

func TestIsLockError(t *testing.T) {
	lockErr := &LockError{
		Reason: "timeout",
	}

	if !IsLockError(lockErr) {
		t.Error("Expected IsLockError to return true")
	}

	if IsLockError(errors.New("normal error")) {
		t.Error("Expected IsLockError to return false")
	}
}

func TestErrInvalidConfigurationField(t *testing.T) {
	err := ErrInvalidConfigurationField("DatabaseURL", "cannot be empty")

	if err == nil {
		t.Fatal("Expected error")
	}

	confErr, ok := err.(*ConfigurationError)
	if !ok {
		t.Fatal("Expected ConfigurationError type")
	}

	if confErr.Field != "DatabaseURL" {
		t.Errorf("Expected Field to be DatabaseURL")
	}

	if confErr.Message != "cannot be empty" {
		t.Errorf("Expected Message to be set")
	}
}
