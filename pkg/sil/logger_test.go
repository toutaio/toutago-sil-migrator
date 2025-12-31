package sil

import (
	"testing"
)

func TestDefaultLogger(t *testing.T) {
	logger := NewDefaultLogger(true)

	// Should not panic
	logger.Info("test info")
	logger.Warn("test warn")
	logger.Error("test error")
	logger.Debug("test debug")
}

func TestNoopLogger(t *testing.T) {
	logger := NewNoopLogger()

	// Should not panic and do nothing
	logger.Info("test info")
	logger.Warn("test warn")
	logger.Error("test error")
	logger.Debug("test debug")
}

func TestColorLogger(t *testing.T) {
	// Test with color enabled
	logger := NewColorLogger(true)

	// Should not panic
	logger.Info("test info")
	logger.Warn("test warn")
	logger.Error("test error")
	logger.Debug("test debug")

	// Test with color disabled
	logger = NewColorLogger(false)
	logger.Info("test info")
	logger.Warn("test warn")
	logger.Error("test error")
	logger.Debug("test debug")
}

func TestColorLogger_Output(t *testing.T) {
	// We can't directly test the output since it goes to stdout,
	// but we can at least verify it doesn't panic
	logger := NewColorLogger(true)
	logger.Info("info message with %s", "formatting")
	logger.Warn("warn message")
	logger.Error("error message")
	logger.Debug("debug message")
}

func TestColorLogger_FormatString(t *testing.T) {
	logger := NewColorLogger(false)

	// Test formatting
	logger.Info("User %s logged in from %s", "john", "192.168.1.1")
	logger.Warn("Disk usage at %d%%", 85)
	logger.Error("Failed to connect: %v", "timeout")
	logger.Debug("Request took %dms", 150)
}

func TestDefaultLogger_Output(t *testing.T) {
	logger := NewDefaultLogger(false)

	// These should not panic
	logger.Info("message")
	logger.Warn("message")
	logger.Error("message")
	logger.Debug("message")
}

func TestLoggerInterface(t *testing.T) {
	// Ensure all logger types implement the Logger interface
	var _ Logger = NewDefaultLogger(true)
	var _ Logger = NewNoopLogger()
	var _ Logger = NewColorLogger(true)
	var _ Logger = NewColorLogger(false)
}

func TestColorOutput(t *testing.T) {
	// Test that color codes are used when enabled
	logger := NewColorLogger(true)

	// Just verify it doesn't crash
	logger.Info("colored info")
	logger.Warn("colored warn")
	logger.Error("colored error")
	logger.Debug("colored debug")
}

func TestNoColorOutput(t *testing.T) {
	// Test that no color codes are used when disabled
	logger := NewColorLogger(false)

	// Just verify it doesn't crash
	logger.Info("plain info")
	logger.Warn("plain warn")
	logger.Error("plain error")
	logger.Debug("plain debug")
}
