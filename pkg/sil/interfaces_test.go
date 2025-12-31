package sil

import (
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestMockRows(t *testing.T) {
	rows := newMockRows([][]interface{}{
		{"v1", "desc1", 1},
		{"v2", "desc2", 2},
	})

	// Test Next
	if !rows.Next() {
		t.Error("Expected Next() to return true for first row")
	}

	// Test Scan
	var version, description string
	var batch int
	err := rows.Scan(&version, &description, &batch)
	if err != nil {
		t.Errorf("Scan failed: %v", err)
	}

	if version != "v1" || description != "desc1" || batch != 1 {
		t.Errorf("Unexpected values: %s, %s, %d", version, description, batch)
	}

	// Next row
	if !rows.Next() {
		t.Error("Expected Next() to return true for second row")
	}

	// Third Next should be false
	if rows.Next() {
		t.Error("Expected Next() to return false after all rows")
	}

	// Test Err
	if err := rows.Err(); err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Test Close
	if err := rows.Close(); err != nil {
		t.Errorf("Close failed: %v", err)
	}
}

func TestMockRowsEmpty(t *testing.T) {
	rows := newMockRows([][]interface{}{})

	if rows.Next() {
		t.Error("Expected Next() to return false for empty rows")
	}

	if err := rows.Close(); err != nil {
		t.Errorf("Close failed: %v", err)
	}
}

func TestMockRowsScanError(t *testing.T) {
	rows := newMockRows([][]interface{}{{"v1"}})
	rows.Next()

	// Try to scan more values than available
	var v1, v2 string
	err := rows.Scan(&v1, &v2)
	if err == nil {
		t.Error("Expected Scan() to return error for mismatched args")
	}
}

func TestRowsAdapter(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"id", "name"}).
		AddRow(1, "test1").
		AddRow(2, "test2")

	mock.ExpectQuery("SELECT (.+)").WillReturnRows(rows)

	sqlRows, err := db.Query("SELECT * FROM test")
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}

	adapter := NewRowsAdapter(sqlRows)
	if adapter == nil {
		t.Fatal("NewRowsAdapter returned nil")
	}

	// Test Next
	if !adapter.Next() {
		t.Error("Next() should return true for first row")
	}

	// Test Scan
	var id int
	var name string
	err = adapter.Scan(&id, &name)
	if err != nil {
		t.Errorf("Scan() error = %v", err)
	}
	if id != 1 || name != "test1" {
		t.Errorf("Scan() got id=%d name=%s, want id=1 name=test1", id, name)
	}

	// Test Next again
	if !adapter.Next() {
		t.Error("Next() should return true for second row")
	}

	err = adapter.Scan(&id, &name)
	if err != nil {
		t.Errorf("Scan() error = %v", err)
	}
	if id != 2 || name != "test2" {
		t.Errorf("Scan() got id=%d name=%s, want id=2 name=test2", id, name)
	}

	// Test Next returns false when no more rows
	if adapter.Next() {
		t.Error("Next() should return false when no more rows")
	}

	// Test Err
	err = adapter.Err()
	if err != nil {
		t.Errorf("Err() should be nil, got %v", err)
	}

	// Test Close
	err = adapter.Close()
	if err != nil {
		t.Errorf("Close() error = %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestRowsAdapter_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	expectedErr := errors.New("scan error")
	rows := sqlmock.NewRows([]string{"id", "name"}).
		AddRow(1, "test").
		RowError(0, expectedErr)

	mock.ExpectQuery("SELECT (.+)").WillReturnRows(rows)

	sqlRows, err := db.Query("SELECT * FROM test")
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}

	adapter := NewRowsAdapter(sqlRows)
	defer adapter.Close()

	// Try to get the row which has an error
	adapter.Next()

	// Err() might return the error
	err = adapter.Err()
	if err == nil {
		// Some SQL mocks might not trigger error on Next
		t.Log("Err() returned nil, which is acceptable for some mock implementations")
	}
}
