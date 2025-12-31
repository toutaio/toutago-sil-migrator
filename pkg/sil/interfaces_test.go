package sil

import "testing"

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
