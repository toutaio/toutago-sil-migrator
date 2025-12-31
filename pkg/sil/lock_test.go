package sil

import (
	"errors"
	"testing"
)

func TestNewLock(t *testing.T) {
	called := false
	releaseFunc := func() error {
		called = true
		return nil
	}

	lock := newLock(releaseFunc, NewDefaultLogger(false))
	if lock == nil {
		t.Fatal("newLock returned nil")
	}

	if !lock.IsLocked() {
		t.Error("new lock should be locked")
	}

	err := lock.Release()
	if err != nil {
		t.Errorf("Release() error = %v", err)
	}

	if !called {
		t.Error("Release() did not call releaseFunc")
	}

	if lock.IsLocked() {
		t.Error("lock should not be locked after release")
	}
}

func TestLock_ReleaseError(t *testing.T) {
	expectedErr := errors.New("release failed")
	releaseFunc := func() error {
		return expectedErr
	}

	lock := newLock(releaseFunc, NewDefaultLogger(false))
	err := lock.Release()

	if err == nil {
		t.Error("Release() should return error")
	}

	// Check if the expected error is wrapped
	if !errors.Is(err, expectedErr) {
		t.Errorf("Release() error does not contain expected error: %v", err)
	}
}

func TestLock_DoubleRelease(t *testing.T) {
	callCount := 0
	releaseFunc := func() error {
		callCount++
		return nil
	}

	lock := newLock(releaseFunc, NewDefaultLogger(false))

	err := lock.Release()
	if err != nil {
		t.Errorf("first Release() error = %v", err)
	}

	// Second release should be a no-op
	err = lock.Release()
	if err != nil {
		t.Errorf("second Release() error = %v", err)
	}

	if callCount != 1 {
		t.Errorf("releaseFunc called %d times, want 1", callCount)
	}
}
