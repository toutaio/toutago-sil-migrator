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
if len(errStr) == 0 {
t.Error("Error string should not be empty")
}
}

func TestSeederErrorFormatting(t *testing.T) {
err := &SeederError{
Seeder: "test_seeder",
Err:    errors.New("failed"),
}

errStr := err.Error()
if len(errStr) == 0 {
t.Error("Error string should not be empty")
}
}
