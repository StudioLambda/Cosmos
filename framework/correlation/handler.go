package correlation

import (
	"context"
	"log/slog"

	"github.com/studiolambda/cosmos/contract"
)

// Handler wraps a [slog.Handler] to automatically
// include the correlation ID in every log record whose context
// carries one. This enables transparent correlation ID propagation
// across all log calls without requiring manual attribute injection.
//
// Usage: wrap your existing handler when constructing the logger:
//
//	handler := slog.NewJSONHandler(os.Stdout, nil)
//	logger := slog.New(correlation.Handler(handler))
//
// Then use context-aware logging methods in your handlers:
//
//	logger.InfoContext(r.Context(), "processing request")
//	// Output automatically includes: correlation_id=<value>
//
// If no correlation ID is present in the context, the log record
// is passed through unmodified.
func Handler(next slog.Handler) slog.Handler {
	return handler{next: next}
}

// handler is a [slog.Handler] decorator that enriches
// log records with the correlation ID from context.
type handler struct {
	next slog.Handler
}

// Enabled delegates to the wrapped handler.
func (handler handler) Enabled(ctx context.Context, level slog.Level) bool {
	return handler.next.Enabled(ctx, level)
}

// Handle adds the correlation ID attribute to the record if present
// in the context, then delegates to the wrapped handler.
func (handler handler) Handle(ctx context.Context, record slog.Record) error {
	if id, ok := ctx.Value(contract.CorrelationIDKey).(string); ok && id != "" {
		record.AddAttrs(slog.String("correlation_id", id))
	}

	return handler.next.Handle(ctx, record)
}

// WithAttrs returns a new handler with the given attributes.
func (handler handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return Handler(handler.next.WithAttrs(attrs))
}

// WithGroup returns a new handler with the given group name.
func (handler handler) WithGroup(name string) slog.Handler {
	return Handler(handler.next.WithGroup(name))
}
