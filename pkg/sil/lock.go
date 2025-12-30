package sil

import (
	"context"
	"fmt"
	"time"
)

// lock implements the Lock interface.
type lock struct {
	releaseFunc func() error
	locked      bool
	logger      Logger
}

// newLock creates a new lock.
func newLock(releaseFunc func() error, logger Logger) Lock {
	if logger == nil {
		logger = NewNoopLogger()
	}

	return &lock{
		releaseFunc: releaseFunc,
		locked:      true,
		logger:      logger,
	}
}

// Release releases the lock.
func (l *lock) Release() error {
	if !l.locked {
		return nil // Already released
	}

	l.logger.Debug("Releasing lock")

	if err := l.releaseFunc(); err != nil {
		return fmt.Errorf("failed to release lock: %w", err)
	}

	l.locked = false
	return nil
}

// IsLocked returns whether the lock is currently held.
func (l *lock) IsLocked() bool {
	return l.locked
}

// AcquireLockWithTimeout attempts to acquire a lock with timeout.
// It will retry acquiring the lock until timeout is reached.
func AcquireLockWithTimeout(ctx context.Context, acquireFunc func() (Lock, error), timeout time.Duration) (Lock, error) {
	deadline := time.Now().Add(timeout)
	attempts := 0

	for {
		attempts++

		// Try to acquire lock
		lock, err := acquireFunc()
		if err == nil {
			return lock, nil
		}

		// Check if we've exceeded the timeout
		if time.Now().After(deadline) {
			return nil, &LockError{
				Reason:  fmt.Sprintf("timeout after %d attempts", attempts),
				Timeout: timeout,
			}
		}

		// Check if context is cancelled
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("lock acquisition cancelled: %w", ctx.Err())
		default:
		}

		// Wait before retrying (exponential backoff)
		waitTime := time.Duration(attempts) * 100 * time.Millisecond
		if waitTime > 5*time.Second {
			waitTime = 5 * time.Second
		}

		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("lock acquisition cancelled: %w", ctx.Err())
		case <-time.After(waitTime):
			// Continue to next attempt
		}
	}
}
