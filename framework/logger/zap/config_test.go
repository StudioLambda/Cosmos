package zap_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	logger "github.com/studiolambda/cosmos/framework/logger/zap"
)

func TestNewWritesJSON(t *testing.T) {
	t.Parallel()

	var output bytes.Buffer
	driver, err := logger.New(logger.Config{Output: &output})
	require.NoError(t, err)

	driver.InfoContext(context.Background(), "started")

	require.Contains(t, output.String(), `"msg":"started"`)
}

func TestNewWritesPlainText(t *testing.T) {
	t.Parallel()

	var output bytes.Buffer
	driver, err := logger.New(logger.Config{Format: logger.FormatText, Output: &output})
	require.NoError(t, err)

	driver.InfoContext(context.Background(), "started")

	require.Contains(t, output.String(), "started")
	require.NotContains(t, output.String(), "\x1b[")
}

func TestNewWritesColoredText(t *testing.T) {
	t.Parallel()

	var output bytes.Buffer
	driver, err := logger.New(logger.Config{Format: logger.FormatColored, Output: &output})
	require.NoError(t, err)

	driver.InfoContext(context.Background(), "started")

	require.Contains(t, output.String(), "\x1b[")
}

func TestNewRejectsInvalidFormat(t *testing.T) {
	t.Parallel()

	_, err := logger.New(logger.Config{Format: "xml"})

	require.Error(t, err)
}

func TestNewRejectsInvalidLevel(t *testing.T) {
	t.Parallel()

	_, err := logger.New(logger.Config{Level: "trace"})

	require.Error(t, err)
}
