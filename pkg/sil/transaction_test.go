package sil

import (
	"context"
	"errors"
	"testing"
)

func TestWithTransaction(t *testing.T) {
	tests := []struct {
		name         string
		setupMock    func(*mockAdapter)
		fn           func(Transaction) error
		wantErr      bool
		wantCommit   bool
		wantRollback bool
	}{
		{
			name: "successful transaction",
			setupMock: func(m *mockAdapter) {
				m.tx = &mockTx{}
			},
			fn: func(tx Transaction) error {
				return tx.Exec(context.Background(), "INSERT INTO test VALUES (1)")
			},
			wantErr:      false,
			wantCommit:   true,
			wantRollback: false,
		},
		{
			name: "transaction with error rolls back",
			setupMock: func(m *mockAdapter) {
				m.tx = &mockTx{}
			},
			fn: func(tx Transaction) error {
				return errors.New("operation failed")
			},
			wantErr:      true,
			wantCommit:   false,
			wantRollback: true,
		},
		{
			name: "begin transaction error",
			setupMock: func(m *mockAdapter) {
				m.beginTxErr = errors.New("begin failed")
			},
			fn: func(tx Transaction) error {
				return nil
			},
			wantErr:      true,
			wantCommit:   false,
			wantRollback: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := &mockAdapter{}
			tt.setupMock(adapter)

			ctx := context.Background()
			err := WithTransaction(ctx, adapter, tt.fn)

			if (err != nil) != tt.wantErr {
				t.Errorf("WithTransaction() error = %v, wantErr %v", err, tt.wantErr)
			}

			if adapter.tx != nil {
				if adapter.tx.committed != tt.wantCommit {
					t.Errorf("Transaction committed = %v, want %v", adapter.tx.committed, tt.wantCommit)
				}
				if adapter.tx.rolledBack != tt.wantRollback {
					t.Errorf("Transaction rolled back = %v, want %v", adapter.tx.rolledBack, tt.wantRollback)
				}
			}
		})
	}
}

func TestWithTransactionAndRecover(t *testing.T) {
	tests := []struct {
		name         string
		setupMock    func(*mockAdapter)
		fn           func(Transaction) error
		wantErr      bool
		wantPanic    bool
		wantRollback bool
	}{
		{
			name: "successful transaction",
			setupMock: func(m *mockAdapter) {
				m.tx = &mockTx{}
			},
			fn: func(tx Transaction) error {
				return nil
			},
			wantErr:      false,
			wantPanic:    false,
			wantRollback: false,
		},
		{
			name: "panic is recovered and rolled back",
			setupMock: func(m *mockAdapter) {
				m.tx = &mockTx{}
			},
			fn: func(tx Transaction) error {
				panic("something went wrong")
			},
			wantErr:      true,
			wantPanic:    true,
			wantRollback: true,
		},
		{
			name: "error without panic",
			setupMock: func(m *mockAdapter) {
				m.tx = &mockTx{}
			},
			fn: func(tx Transaction) error {
				return errors.New("normal error")
			},
			wantErr:      true,
			wantPanic:    false,
			wantRollback: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := &mockAdapter{}
			tt.setupMock(adapter)

			ctx := context.Background()
			err := WithTransactionAndRecover(ctx, adapter, tt.fn)

			if (err != nil) != tt.wantErr {
				t.Errorf("WithTransactionAndRecover() error = %v, wantErr %v", err, tt.wantErr)
			}

			if adapter.tx != nil && adapter.tx.rolledBack != tt.wantRollback {
				t.Errorf("Transaction rolled back = %v, want %v", adapter.tx.rolledBack, tt.wantRollback)
			}
		})
	}
}

func TestTransactionExec(t *testing.T) {
	adapter := &mockAdapter{
		tx: &mockTx{},
	}

	ctx := context.Background()
	err := WithTransaction(ctx, adapter, func(tx Transaction) error {
		err := tx.Exec(ctx, "CREATE TABLE test (id INT)")
		if err != nil {
			return err
		}

		err = tx.Exec(ctx, "INSERT INTO test VALUES (1)")
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		t.Errorf("Transaction failed: %v", err)
	}

	if !adapter.tx.committed {
		t.Error("Transaction should have been committed")
	}
}

func TestTransactionRollbackOnError(t *testing.T) {
	adapter := &mockAdapter{
		tx: &mockTx{},
	}

	ctx := context.Background()
	testErr := errors.New("exec failed")

	err := WithTransaction(ctx, adapter, func(tx Transaction) error {
		return testErr
	})

	if err == nil {
		t.Error("Expected error from failed exec")
	}

	if !adapter.tx.rolledBack {
		t.Error("Transaction should have been rolled back")
	}

	if adapter.tx.committed {
		t.Error("Transaction should not have been committed")
	}
}
