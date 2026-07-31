package contract

import "context"

// LogLevel identifies a structured logging severity.
type LogLevel string

const (
	// LogLevelDebug identifies debug-level records.
	LogLevelDebug LogLevel = "debug"

	// LogLevelInfo identifies informational records.
	LogLevelInfo LogLevel = "info"

	// LogLevelWarn identifies warning-level records.
	LogLevelWarn LogLevel = "warn"

	// LogLevelError identifies error-level records.
	LogLevelError LogLevel = "error"
)

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
	logger.DebugContext(context.Background(), message, args...)
}

// DebugContext logs a debug-level message with the given context and attributes.
func (logger *Logger) DebugContext(ctx context.Context, message string, args ...any) {
	logger.driver.DebugContext(ctx, message, args...)
}

// Info logs an info-level message with a background context and attributes.
func (logger *Logger) Info(message string, args ...any) {
	logger.InfoContext(context.Background(), message, args...)
}

// InfoContext logs an info-level message with the given context and attributes.
func (logger *Logger) InfoContext(ctx context.Context, message string, args ...any) {
	logger.driver.InfoContext(ctx, message, args...)
}

// Warn logs a warning-level message with a background context and attributes.
func (logger *Logger) Warn(message string, args ...any) {
	logger.WarnContext(context.Background(), message, args...)
}

// WarnContext logs a warning-level message with the given context and attributes.
func (logger *Logger) WarnContext(ctx context.Context, message string, args ...any) {
	logger.driver.WarnContext(ctx, message, args...)
}

// Error logs an error-level message with a background context and attributes.
func (logger *Logger) Error(message string, args ...any) {
	logger.ErrorContext(context.Background(), message, args...)
}

// ErrorContext logs an error-level message with the given context and attributes.
func (logger *Logger) ErrorContext(ctx context.Context, message string, args ...any) {
	logger.driver.ErrorContext(ctx, message, args...)
}

// Log logs a message at level with a background context and attributes.
func (logger *Logger) Log(level LogLevel, message string, args ...any) {
	logger.LogContext(context.Background(), level, message, args...)
}

// LogContext logs a message at level with the given context and attributes.
// Unknown levels are logged as errors to avoid silently dropping records.
func (logger *Logger) LogContext(ctx context.Context, level LogLevel, message string, args ...any) {
	switch level {
	case LogLevelDebug:
		logger.DebugContext(ctx, message, args...)
	case LogLevelInfo:
		logger.InfoContext(ctx, message, args...)
	case LogLevelWarn:
		logger.WarnContext(ctx, message, args...)
	default:
		logger.ErrorContext(ctx, message, args...)
	}
}

// With returns a derived [Logger] with the given persistent attributes.
func (logger *Logger) With(args ...any) *Logger {
	return NewLogger(logger.driver.With(args...))
}

// WithError returns a derived [Logger] with err as a persistent "err" attribute.
func (logger *Logger) WithError(err error) *Logger {
	return logger.With("err", err)
}

// Named returns a derived [Logger] with name as a persistent "component" attribute.
func (logger *Logger) Named(name string) *Logger {
	return logger.With("component", name)
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
