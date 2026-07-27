package logger_test

import (
	"bytes"
	"context"
	"encoding/json/v2"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
	"github.com/studiolambda/cosmos/framework/logger"
)

type zerologContextHook struct {
	context context.Context
}

func (hook *zerologContextHook) Run(event *zerolog.Event, _ zerolog.Level, _ string) {
	hook.context = event.GetCtx()
}

func TestZerologLogsContextualRecord(t *testing.T) {
	t.Parallel()

	var output bytes.Buffer
	driver := logger.NewZerologFrom(zerolog.New(&output))
	ctx := context.WithValue(context.Background(), "request_id", "abc123")

	driver.InfoContext(ctx, "request complete", "status", 200)

	var record map[string]any
	err := json.Unmarshal(output.Bytes(), &record)
	require.NoError(t, err)
	require.Equal(t, "info", record["level"])
	require.Equal(t, "request complete", record["message"])
	require.Equal(t, float64(200), record["status"])
}

func TestZerologWithAddsPersistentAttributes(t *testing.T) {
	t.Parallel()

	var output bytes.Buffer
	driver := logger.NewZerologFrom(zerolog.New(&output))

	driver.With("component", "http").InfoContext(context.Background(), "started")

	var record map[string]any
	err := json.Unmarshal(output.Bytes(), &record)
	require.NoError(t, err)
	require.Equal(t, "http", record["component"])
	require.Equal(t, "started", record["message"])
}

func TestZerologDebugLogsDebugRecord(t *testing.T) {
	t.Parallel()

	var output bytes.Buffer
	driver := logger.NewZerologFrom(zerolog.New(&output))

	driver.DebugContext(context.Background(), "debug")

	var record map[string]any
	err := json.Unmarshal(output.Bytes(), &record)
	require.NoError(t, err)
	require.Equal(t, "debug", record["level"])
}

func TestZerologWarnLogsWarningRecord(t *testing.T) {
	t.Parallel()

	var output bytes.Buffer
	driver := logger.NewZerologFrom(zerolog.New(&output))

	driver.WarnContext(context.Background(), "warn")

	var record map[string]any
	err := json.Unmarshal(output.Bytes(), &record)
	require.NoError(t, err)
	require.Equal(t, "warn", record["level"])
}

func TestZerologErrorLogsErrorRecord(t *testing.T) {
	t.Parallel()

	var output bytes.Buffer
	driver := logger.NewZerologFrom(zerolog.New(&output))

	driver.ErrorContext(context.Background(), "error")

	var record map[string]any
	err := json.Unmarshal(output.Bytes(), &record)
	require.NoError(t, err)
	require.Equal(t, "error", record["level"])
}

func TestZerologPassesContextToHook(t *testing.T) {
	t.Parallel()

	var output bytes.Buffer
	hook := &zerologContextHook{}
	driver := logger.NewZerologFrom(zerolog.New(&output).Hook(hook))
	ctx := context.WithValue(context.Background(), "request_id", "abc123")

	driver.InfoContext(ctx, "request complete")

	require.Same(t, ctx, hook.context)
}
