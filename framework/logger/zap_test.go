package logger_test

import (
	"bytes"
	"context"
	"encoding/json/v2"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/studiolambda/cosmos/framework/logger"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestZapLogsContextualRecord(t *testing.T) {
	t.Parallel()

	var output bytes.Buffer
	driver := logger.NewZapFrom(newZapLogger(&output))
	ctx := context.WithValue(context.Background(), "request_id", "abc123")

	driver.InfoContext(ctx, "request complete", "status", 200)

	var record map[string]any
	err := json.Unmarshal(output.Bytes(), &record)
	require.NoError(t, err)
	require.Equal(t, "info", record["level"])
	require.Equal(t, "request complete", record["msg"])
	require.Equal(t, float64(200), record["status"])
}

func TestZapWithAddsPersistentAttributes(t *testing.T) {
	t.Parallel()

	var output bytes.Buffer
	driver := logger.NewZapFrom(newZapLogger(&output))

	driver.With("component", "http").InfoContext(context.Background(), "started")

	var record map[string]any
	err := json.Unmarshal(output.Bytes(), &record)
	require.NoError(t, err)
	require.Equal(t, "http", record["component"])
	require.Equal(t, "started", record["msg"])
}

func TestZapDebugLogsDebugRecord(t *testing.T) {
	t.Parallel()

	var output bytes.Buffer
	driver := logger.NewZapFrom(newZapLogger(&output))

	driver.DebugContext(context.Background(), "debug")

	var record map[string]any
	err := json.Unmarshal(output.Bytes(), &record)
	require.NoError(t, err)
	require.Equal(t, "debug", record["level"])
}

func TestZapWarnLogsWarningRecord(t *testing.T) {
	t.Parallel()

	var output bytes.Buffer
	driver := logger.NewZapFrom(newZapLogger(&output))

	driver.WarnContext(context.Background(), "warn")

	var record map[string]any
	err := json.Unmarshal(output.Bytes(), &record)
	require.NoError(t, err)
	require.Equal(t, "warn", record["level"])
}

func TestZapErrorLogsErrorRecord(t *testing.T) {
	t.Parallel()

	var output bytes.Buffer
	driver := logger.NewZapFrom(newZapLogger(&output))

	driver.ErrorContext(context.Background(), "error")

	var record map[string]any
	err := json.Unmarshal(output.Bytes(), &record)
	require.NoError(t, err)
	require.Equal(t, "error", record["level"])
}

func TestNewZapFromNilDiscardsRecords(t *testing.T) {
	t.Parallel()

	driver := logger.NewZapFrom(nil)

	require.NotPanics(t, func() {
		driver.ErrorContext(context.Background(), "discarded")
	})
}

func newZapLogger(output *bytes.Buffer) *zap.Logger {
	encoder := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
	core := zapcore.NewCore(encoder, zapcore.AddSync(output), zapcore.DebugLevel)

	return zap.New(core)
}
