package sil

import (
	"fmt"
	"log"
	"os"
)

// defaultLogger is a simple logger implementation that logs to stdout.
type defaultLogger struct {
	verbose bool
}

// NewDefaultLogger creates a new default logger.
func NewDefaultLogger(verbose bool) Logger {
	return &defaultLogger{verbose: verbose}
}

// Info logs an informational message.
func (l *defaultLogger) Info(msg string, args ...interface{}) {
	log.Printf("[INFO] "+msg, args...)
}

// Warn logs a warning message.
func (l *defaultLogger) Warn(msg string, args ...interface{}) {
	log.Printf("[WARN] "+msg, args...)
}

// Error logs an error message.
func (l *defaultLogger) Error(msg string, args ...interface{}) {
	log.Printf("[ERROR] "+msg, args...)
}

// Debug logs a debug message (only if verbose is enabled).
func (l *defaultLogger) Debug(msg string, args ...interface{}) {
	if l.verbose {
		log.Printf("[DEBUG] "+msg, args...)
	}
}

// noopLogger is a logger that does nothing.
type noopLogger struct{}

// NewNoopLogger creates a logger that discards all messages.
func NewNoopLogger() Logger {
	return &noopLogger{}
}

// Info does nothing.
func (l *noopLogger) Info(msg string, args ...interface{}) {}

// Warn does nothing.
func (l *noopLogger) Warn(msg string, args ...interface{}) {}

// Error does nothing.
func (l *noopLogger) Error(msg string, args ...interface{}) {}

// Debug does nothing.
func (l *noopLogger) Debug(msg string, args ...interface{}) {}

// colorLogger adds color to log messages for better readability.
type colorLogger struct {
	verbose bool
}

// NewColorLogger creates a logger with colored output.
func NewColorLogger(verbose bool) Logger {
	return &colorLogger{verbose: verbose}
}

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorPurple = "\033[35m"
	colorCyan   = "\033[36m"
	colorWhite  = "\033[37m"
)

// Info logs an informational message in green.
func (l *colorLogger) Info(msg string, args ...interface{}) {
	message := fmt.Sprintf(msg, args...)
	_, _ = fmt.Fprintf(os.Stdout, "%s[INFO]%s %s\n", colorGreen, colorReset, message)
}

// Warn logs a warning message in yellow.
func (l *colorLogger) Warn(msg string, args ...interface{}) {
	message := fmt.Sprintf(msg, args...)
	_, _ = fmt.Fprintf(os.Stdout, "%s[WARN]%s %s\n", colorYellow, colorReset, message)
}

// Error logs an error message in red.
func (l *colorLogger) Error(msg string, args ...interface{}) {
	message := fmt.Sprintf(msg, args...)
	_, _ = fmt.Fprintf(os.Stderr, "%s[ERROR]%s %s\n", colorRed, colorReset, message)
}

// Debug logs a debug message in cyan (only if verbose is enabled).
func (l *colorLogger) Debug(msg string, args ...interface{}) {
	if l.verbose {
		message := fmt.Sprintf(msg, args...)
		_, _ = fmt.Fprintf(os.Stdout, "%s[DEBUG]%s %s\n", colorCyan, colorReset, message)
	}
}
