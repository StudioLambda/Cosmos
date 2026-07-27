package logger

import (
	"context"
	stdslog "log/slog"

	"github.com/studiolambda/cosmos/contract"
)

// Slog implements [contract.LoggerDriver] with [slog.Logger].
type Slog struct {
	logger *stdslog.Logger
}

// NewSlogFrom creates a new Slog driver from an existing [slog.Logger].
// A nil logger produces a discard driver.
func NewSlogFrom(logger *stdslog.Logger) *Slog {
	if logger == nil {
		logger = stdslog.New(stdslog.DiscardHandler)
	}

	return &Slog{logger: logger}
}

// DebugContext logs a debug-level message with the given context and attributes.
func (slog *Slog) DebugContext(ctx context.Context, message string, args ...any) {
	slog.logger.DebugContext(ctx, message, args...)
}

// InfoContext logs an info-level message with the given context and attributes.
func (slog *Slog) InfoContext(ctx context.Context, message string, args ...any) {
	slog.logger.InfoContext(ctx, message, args...)
}

// WarnContext logs a warning-level message with the given context and attributes.
func (slog *Slog) WarnContext(ctx context.Context, message string, args ...any) {
	slog.logger.WarnContext(ctx, message, args...)
}

// ErrorContext logs an error-level message with the given context and attributes.
func (slog *Slog) ErrorContext(ctx context.Context, message string, args ...any) {
	slog.logger.ErrorContext(ctx, message, args...)
}

// With returns a derived Slog driver with the given persistent attributes.
func (slog *Slog) With(args ...any) contract.LoggerDriver {
	return NewSlogFrom(slog.logger.With(args...))
}
