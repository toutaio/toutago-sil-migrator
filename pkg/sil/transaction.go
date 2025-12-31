package sil

import (
	"context"
	"database/sql"
	"fmt"
)

// transaction implements the Transaction interface.
// This type and its methods are part of the public Transaction API,
// though currently used internally only via the interface.
//
//nolint:unused // Part of public Transaction API
type transaction struct {
	tx     *sql.Tx
	logger Logger
}

// newTransaction creates a new transaction wrapper.
//
//nolint:unused // Part of public Transaction API
func newTransaction(tx *sql.Tx, logger Logger) Transaction {
	if logger == nil {
		logger = NewNoopLogger()
	}

	return &transaction{
		tx:     tx,
		logger: logger,
	}
}

// Commit commits the transaction.
func (t *transaction) Commit() error {
	t.logger.Debug("Committing transaction")

	if err := t.tx.Commit(); err != nil {
		return fmt.Errorf("transaction commit failed: %w", err)
	}

	return nil
}

// Rollback rolls back the transaction.
func (t *transaction) Rollback() error {
	t.logger.Debug("Rolling back transaction")

	if err := t.tx.Rollback(); err != nil {
		// Ignore "sql: transaction has already been committed or rolled back"
		if err != sql.ErrTxDone {
			return fmt.Errorf("transaction rollback failed: %w", err)
		}
	}

	return nil
}

// Exec executes a SQL statement within the transaction.
func (t *transaction) Exec(ctx context.Context, query string, args ...interface{}) error {
	t.logger.Debug("Executing query in transaction: %s", truncateQuery(query))

	_, err := t.tx.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("transaction exec failed: %w", err)
	}

	return nil
}

// Query executes a SQL query within the transaction.
func (t *transaction) Query(ctx context.Context, query string, args ...interface{}) (Rows, error) {
	t.logger.Debug("Querying in transaction: %s", truncateQuery(query))

	rows, err := t.tx.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("transaction query failed: %w", err)
	}

	return NewRowsAdapter(rows), nil
}

// truncateQuery truncates a query for logging (first 100 chars).
func truncateQuery(query string) string {
	const maxLen = 100
	if len(query) <= maxLen {
		return query
	}
	return query[:maxLen] + "..."
}

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
			if rbErr := tx.Rollback(); rbErr != nil {
				// Log but don't return error from defer
			}
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
			if rbErr := tx.Rollback(); rbErr != nil {
				// Log but don't return error from defer
			}
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
