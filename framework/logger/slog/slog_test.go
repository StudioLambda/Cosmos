package slog_test

import (
	"bytes"
	"context"
	"encoding/json/v2"
	"errors"
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/studiolambda/cosmos/contract"
	logger "github.com/studiolambda/cosmos/framework/logger/slog"
)

type slogContextHandler struct {
	context context.Context
}

type configurationDriver struct {
	output string
}

func (driver configurationDriver) Unmarshal(key string, dest any) error {
	if key != "logger.output" {
		return errors.New("configuration key not found")
	}

	value, ok := dest.(*string)
	if !ok {
		return errors.New("invalid configuration destination")
	}

	*value = driver.output

	return nil
}

func (configurationDriver) Has(string) bool {
	return true
}

func (configurationDriver) Delimiter() string {
	return "."
}

func (configurationDriver) Extend(...contract.ConfigurationProvider) error {
	return nil
}

func TestConfigFromMapsStdoutOutput(t *testing.T) {
	t.Parallel()

	configuration := contract.NewConfiguration(configurationDriver{output: "stdout"})
	config := configuration.From[logger.Config]("logger")

	require.Same(t, os.Stdout, config.Output.(*os.File))
}

func TestConfigFromMapsStderrOutput(t *testing.T) {
	t.Parallel()

	configuration := contract.NewConfiguration(configurationDriver{output: "stderr"})
	config := configuration.From[logger.Config]("logger")

	require.Same(t, os.Stderr, config.Output.(*os.File))
}

func (handler *slogContextHandler) Enabled(context.Context, slog.Level) bool {
	return true
}

func (handler *slogContextHandler) Handle(ctx context.Context, _ slog.Record) error {
	handler.context = ctx

	return nil
}

func (handler *slogContextHandler) WithAttrs([]slog.Attr) slog.Handler {
	return handler
}

func (handler *slogContextHandler) WithGroup(string) slog.Handler {
	return handler
}

func TestSlogLogsContextualRecord(t *testing.T) {
	t.Parallel()

	var output bytes.Buffer
	driver := logger.NewSlogFrom(slog.New(slog.NewJSONHandler(&output, &slog.HandlerOptions{Level: slog.LevelDebug})))
	ctx := context.WithValue(context.Background(), "request_id", "abc123")

	driver.InfoContext(ctx, "request complete", "status", 200)

	var record map[string]any
	err := json.Unmarshal(output.Bytes(), &record)
	require.NoError(t, err)
	require.Equal(t, "INFO", record["level"])
	require.Equal(t, "request complete", record["msg"])
	require.Equal(t, float64(200), record["status"])
}

func TestSlogWithAddsPersistentAttributes(t *testing.T) {
	t.Parallel()

	var output bytes.Buffer
	driver := logger.NewSlogFrom(slog.New(slog.NewJSONHandler(&output, nil)))

	driver.With("component", "http").InfoContext(context.Background(), "started")

	var record map[string]any
	err := json.Unmarshal(output.Bytes(), &record)
	require.NoError(t, err)
	require.Equal(t, "http", record["component"])
	require.Equal(t, "started", record["msg"])
}

func TestSlogDebugLogsDebugRecord(t *testing.T) {
	t.Parallel()

	var output bytes.Buffer
	driver := logger.NewSlogFrom(slog.New(slog.NewJSONHandler(&output, &slog.HandlerOptions{Level: slog.LevelDebug})))

	driver.DebugContext(context.Background(), "debug")

	var record map[string]any
	err := json.Unmarshal(output.Bytes(), &record)
	require.NoError(t, err)
	require.Equal(t, "DEBUG", record["level"])
}

func TestSlogWarnLogsWarningRecord(t *testing.T) {
	t.Parallel()

	var output bytes.Buffer
	driver := logger.NewSlogFrom(slog.New(slog.NewJSONHandler(&output, nil)))

	driver.WarnContext(context.Background(), "warn")

	var record map[string]any
	err := json.Unmarshal(output.Bytes(), &record)
	require.NoError(t, err)
	require.Equal(t, "WARN", record["level"])
}

func TestSlogErrorLogsErrorRecord(t *testing.T) {
	t.Parallel()

	var output bytes.Buffer
	driver := logger.NewSlogFrom(slog.New(slog.NewJSONHandler(&output, nil)))

	driver.ErrorContext(context.Background(), "error")

	var record map[string]any
	err := json.Unmarshal(output.Bytes(), &record)
	require.NoError(t, err)
	require.Equal(t, "ERROR", record["level"])
}

func TestSlogPassesContextToHandler(t *testing.T) {
	t.Parallel()

	handler := &slogContextHandler{}
	driver := logger.NewSlogFrom(slog.New(handler))
	ctx := context.WithValue(context.Background(), "request_id", "abc123")

	driver.InfoContext(ctx, "request complete")

	require.Same(t, ctx, handler.context)
}

func TestNewSlogFromNilDiscardsRecords(t *testing.T) {
	t.Parallel()

	driver := logger.NewSlogFrom(nil)

	require.NotPanics(t, func() {
		driver.ErrorContext(context.Background(), "discarded")
	})
}
