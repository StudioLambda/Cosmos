package logger

import (
	"context"

	"github.com/studiolambda/cosmos/contract"
	uberzap "go.uber.org/zap"
)

// Zap implements [contract.LoggerDriver] with [zap.SugaredLogger].
type Zap struct {
	logger *uberzap.SugaredLogger
}

// NewZapFrom creates a new Zap driver from an existing [zap.Logger].
// A nil logger produces a discard driver.
func NewZapFrom(logger *uberzap.Logger) *Zap {
	if logger == nil {
		logger = uberzap.NewNop()
	}

	return &Zap{logger: logger.Sugar()}
}

// DebugContext logs a debug-level message with the given context and attributes.
// Zap does not natively propagate contexts.
func (zap *Zap) DebugContext(_ context.Context, message string, args ...any) {
	zap.logger.Debugw(message, args...)
}

// InfoContext logs an info-level message with the given context and attributes.
// Zap does not natively propagate contexts.
func (zap *Zap) InfoContext(_ context.Context, message string, args ...any) {
	zap.logger.Infow(message, args...)
}

// WarnContext logs a warning-level message with the given context and attributes.
// Zap does not natively propagate contexts.
func (zap *Zap) WarnContext(_ context.Context, message string, args ...any) {
	zap.logger.Warnw(message, args...)
}

// ErrorContext logs an error-level message with the given context and attributes.
// Zap does not natively propagate contexts.
func (zap *Zap) ErrorContext(_ context.Context, message string, args ...any) {
	zap.logger.Errorw(message, args...)
}

// With returns a derived Zap driver with the given persistent attributes.
func (zap *Zap) With(args ...any) contract.LoggerDriver {
	return &Zap{logger: zap.logger.With(args...)}
}
