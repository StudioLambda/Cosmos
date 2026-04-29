package correlation_test

import (
	"bytes"
	"context"
	"log/slog"
	"testing"

	"github.com/studiolambda/cosmos/contract"
	"github.com/studiolambda/cosmos/framework/correlation"

	"github.com/stretchr/testify/require"
)

func TestHandlerAddsIDFromContext(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	handler := correlation.Handler(slog.NewTextHandler(&buf, nil))
	logger := slog.New(handler)

	ctx := context.WithValue(context.Background(), contract.CorrelationIDKey, "trace-abc-123")
	logger.InfoContext(ctx, "test message")

	output := buf.String()
	require.Contains(t, output, "correlation_id")
	require.Contains(t, output, "trace-abc-123")
}

func TestHandlerOmitsWhenMissing(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	handler := correlation.Handler(slog.NewTextHandler(&buf, nil))
	logger := slog.New(handler)

	logger.InfoContext(context.Background(), "test message")

	output := buf.String()
	require.NotContains(t, output, "correlation_id")
}

func TestHandlerOmitsWhenEmpty(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	handler := correlation.Handler(slog.NewTextHandler(&buf, nil))
	logger := slog.New(handler)

	ctx := context.WithValue(context.Background(), contract.CorrelationIDKey, "")
	logger.InfoContext(ctx, "test message")

	output := buf.String()
	require.NotContains(t, output, "correlation_id")
}

func TestHandlerPreservesExistingAttrs(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	handler := correlation.Handler(slog.NewTextHandler(&buf, nil))
	logger := slog.New(handler)

	ctx := context.WithValue(context.Background(), contract.CorrelationIDKey, "id-456")
	logger.InfoContext(ctx, "test", "extra", "value")

	output := buf.String()
	require.Contains(t, output, "correlation_id")
	require.Contains(t, output, "id-456")
	require.Contains(t, output, "extra")
	require.Contains(t, output, "value")
}

func TestHandlerWithAttrsPreservesWrapping(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	handler := correlation.Handler(slog.NewTextHandler(&buf, nil))
	logger := slog.New(handler).With("service", "api")

	ctx := context.WithValue(context.Background(), contract.CorrelationIDKey, "id-789")
	logger.InfoContext(ctx, "test")

	output := buf.String()
	require.Contains(t, output, "service")
	require.Contains(t, output, "api")
	require.Contains(t, output, "correlation_id")
	require.Contains(t, output, "id-789")
}

func TestHandlerWithGroupPreservesWrapping(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	handler := correlation.Handler(slog.NewTextHandler(&buf, nil))
	logger := slog.New(handler).WithGroup("request")

	ctx := context.WithValue(context.Background(), contract.CorrelationIDKey, "id-group")
	logger.InfoContext(ctx, "test")

	output := buf.String()
	require.Contains(t, output, "id-group")
}

func TestHandlerRespectsLevel(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	handler := correlation.Handler(
		slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelError}),
	)
	logger := slog.New(handler)

	ctx := context.WithValue(context.Background(), contract.CorrelationIDKey, "id-level")
	logger.InfoContext(ctx, "should not appear")

	require.Empty(t, buf.String())
}
