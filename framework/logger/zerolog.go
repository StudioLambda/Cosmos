package logger

import (
	"context"

	rszerolog "github.com/rs/zerolog"
	"github.com/studiolambda/cosmos/contract"
)

// Zerolog implements [contract.LoggerDriver] with [zerolog.Logger].
type Zerolog struct {
	logger rszerolog.Logger
}

// NewZerologFrom creates a new Zerolog driver from an existing [zerolog.Logger].
func NewZerologFrom(logger rszerolog.Logger) *Zerolog {
	return &Zerolog{logger: logger}
}

// DebugContext logs a debug-level message with the given context and attributes.
func (zerolog *Zerolog) DebugContext(ctx context.Context, message string, args ...any) {
	zerolog.logger.Debug().Ctx(ctx).Fields(args).Msg(message)
}

// InfoContext logs an info-level message with the given context and attributes.
func (zerolog *Zerolog) InfoContext(ctx context.Context, message string, args ...any) {
	zerolog.logger.Info().Ctx(ctx).Fields(args).Msg(message)
}

// WarnContext logs a warning-level message with the given context and attributes.
func (zerolog *Zerolog) WarnContext(ctx context.Context, message string, args ...any) {
	zerolog.logger.Warn().Ctx(ctx).Fields(args).Msg(message)
}

// ErrorContext logs an error-level message with the given context and attributes.
func (zerolog *Zerolog) ErrorContext(ctx context.Context, message string, args ...any) {
	zerolog.logger.Error().Ctx(ctx).Fields(args).Msg(message)
}

// With returns a derived Zerolog driver with the given persistent attributes.
func (zerolog *Zerolog) With(args ...any) contract.LoggerDriver {
	return NewZerologFrom(zerolog.logger.With().Fields(args).Logger())
}
