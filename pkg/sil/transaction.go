package sil

import (
	"context"
	"fmt"
)

// WithTransaction executes a function within a transaction.
// If the function returns an error, the transaction is rolled back.
// Otherwise, it is committed.
func WithTransaction(ctx context.Context, adapter DatabaseAdapter, fn func(tx Transaction) error) error {
	tx, err := adapter.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Track whether we need to rollback
	needsRollback := true
	defer func() {
		if needsRollback {
			_ = tx.Rollback() // Ignore rollback errors in defer
			// Log but don't return error from defer
		}
	}()

	// Execute function
	if err := fn(tx); err != nil {
		return err
	}

	// Commit
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	needsRollback = false
	return nil
}

// WithTransactionAndRecover executes a function within a transaction with panic recovery.
// If the function panics or returns an error, the transaction is rolled back.
func WithTransactionAndRecover(ctx context.Context, adapter DatabaseAdapter, fn func(tx Transaction) error) (err error) {
	tx, err := adapter.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Track whether we need to rollback
	needsRollback := true
	defer func() {
		if needsRollback {
			_ = tx.Rollback() // Ignore rollback errors in defer
			// Log but don't return error from defer
		}
	}()

	// Execute function with panic recovery
	func() {
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("transaction panicked: %v", r)
			}
		}()

		err = fn(tx)
	}()

	if err != nil {
		return err
	}

	// Commit
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	needsRollback = false
	return nil
}
