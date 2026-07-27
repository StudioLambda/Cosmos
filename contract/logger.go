package contract

import "context"

// LoggerDriver defines the contract implemented by structured logging backends.
// Arguments use alternating string key and value pairs.
type LoggerDriver interface {
	// DebugContext logs a debug-level message with the given context and attributes.
	DebugContext(ctx context.Context, message string, args ...any)

	// InfoContext logs an info-level message with the given context and attributes.
	InfoContext(ctx context.Context, message string, args ...any)

	// WarnContext logs a warning-level message with the given context and attributes.
	WarnContext(ctx context.Context, message string, args ...any)

	// ErrorContext logs an error-level message with the given context and attributes.
	ErrorContext(ctx context.Context, message string, args ...any)

	// With returns a derived driver with the given persistent attributes.
	With(args ...any) LoggerDriver
}

// Logger provides a backend-agnostic structured logging API over a
// [LoggerDriver].
type Logger struct {
	driver LoggerDriver
}

// NewLogger creates a new [Logger] that delegates to the given driver.
// A nil driver produces a discard logger.
func NewLogger(driver LoggerDriver) *Logger {
	if driver == nil {
		driver = discardLogger{}
	}

	return &Logger{driver: driver}
}

// Driver returns the underlying [LoggerDriver].
func (logger *Logger) Driver() LoggerDriver {
	return logger.driver
}

// Debug logs a debug-level message with a background context and attributes.
func (logger *Logger) Debug(message string, args ...any) {
	logger.driver.DebugContext(context.Background(), message, args...)
}

// Info logs an info-level message with a background context and attributes.
func (logger *Logger) Info(message string, args ...any) {
	logger.driver.InfoContext(context.Background(), message, args...)
}

// Warn logs a warning-level message with a background context and attributes.
func (logger *Logger) Warn(message string, args ...any) {
	logger.driver.WarnContext(context.Background(), message, args...)
}

// Error logs an error-level message with a background context and attributes.
func (logger *Logger) Error(message string, args ...any) {
	logger.driver.ErrorContext(context.Background(), message, args...)
}

// With returns a derived [Logger] with the given persistent attributes.
func (logger *Logger) With(args ...any) *Logger {
	return NewLogger(logger.driver.With(args...))
}

// discardLogger silently discards every log record.
type discardLogger struct{}

func (discardLogger) DebugContext(context.Context, string, ...any) {}

func (discardLogger) InfoContext(context.Context, string, ...any) {}

func (discardLogger) WarnContext(context.Context, string, ...any) {}

func (discardLogger) ErrorContext(context.Context, string, ...any) {}

func (discardLogger) With(...any) LoggerDriver {
	return discardLogger{}
}
