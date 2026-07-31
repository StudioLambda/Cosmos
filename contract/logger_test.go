package contract_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/studiolambda/cosmos/contract"
	contractmock "github.com/studiolambda/cosmos/contract/mock"
)

func TestLoggerMethodsUseBackgroundContext(t *testing.T) {
	t.Parallel()

	driver := contractmock.NewLoggerDriverMock(t)
	logger := contract.NewLogger(driver)

	driver.On("DebugContext", context.Background(), "debug", []any{"id", 1}).Return()
	driver.On("InfoContext", context.Background(), "info", []any{"id", 2}).Return()
	driver.On("WarnContext", context.Background(), "warn", []any{"id", 3}).Return()
	driver.On("ErrorContext", context.Background(), "error", []any{"id", 4}).Return()

	logger.Debug("debug", "id", 1)
	logger.Info("info", "id", 2)
	logger.Warn("warn", "id", 3)
	logger.Error("error", "id", 4)
}

func TestLoggerContextMethodsUseProvidedContext(t *testing.T) {
	t.Parallel()

	driver := contractmock.NewLoggerDriverMock(t)
	logger := contract.NewLogger(driver)
	ctx := context.WithValue(context.Background(), "key", "value")

	driver.On("DebugContext", ctx, "debug", []any{"id", 1}).Return()
	driver.On("InfoContext", ctx, "info", []any{"id", 2}).Return()
	driver.On("WarnContext", ctx, "warn", []any{"id", 3}).Return()
	driver.On("ErrorContext", ctx, "error", []any{"id", 4}).Return()

	logger.DebugContext(ctx, "debug", "id", 1)
	logger.InfoContext(ctx, "info", "id", 2)
	logger.WarnContext(ctx, "warn", "id", 3)
	logger.ErrorContext(ctx, "error", "id", 4)
}

func TestLoggerWithReturnsDerivedLogger(t *testing.T) {
	t.Parallel()

	driver := contractmock.NewLoggerDriverMock(t)
	child := contractmock.NewLoggerDriverMock(t)
	logger := contract.NewLogger(driver)

	driver.On("With", []any{"component", "http"}).Return(child)
	child.On("InfoContext", context.Background(), "started").Return()

	derived := logger.With("component", "http")
	derived.Info("started")

	require.NotSame(t, logger, derived)
	driver.AssertNotCalled(t, "InfoContext", mock.Anything, mock.Anything)
}

func TestLoggerWithErrorReturnsDerivedLogger(t *testing.T) {
	t.Parallel()

	driver := contractmock.NewLoggerDriverMock(t)
	child := contractmock.NewLoggerDriverMock(t)
	logger := contract.NewLogger(driver)
	errExpected := errors.New("failed")

	driver.On("With", []any{"err", errExpected}).Return(child)
	child.On("ErrorContext", context.Background(), "save failed").Return()

	logger.WithError(errExpected).Error("save failed")
}

func TestLoggerNamedReturnsDerivedLogger(t *testing.T) {
	t.Parallel()

	driver := contractmock.NewLoggerDriverMock(t)
	child := contractmock.NewLoggerDriverMock(t)
	logger := contract.NewLogger(driver)

	driver.On("With", []any{"component", "database"}).Return(child)
	child.On("InfoContext", context.Background(), "connected").Return()

	logger.Named("database").Info("connected")
}

func TestLoggerLogContextDispatchesLevel(t *testing.T) {
	t.Parallel()

	driver := contractmock.NewLoggerDriverMock(t)
	logger := contract.NewLogger(driver)
	ctx := context.Background()

	driver.On("DebugContext", ctx, "debug").Return()
	driver.On("InfoContext", ctx, "info").Return()
	driver.On("WarnContext", ctx, "warn").Return()
	driver.On("ErrorContext", ctx, "error").Return()
	driver.On("ErrorContext", ctx, "unknown").Return()

	logger.LogContext(ctx, contract.LogLevelDebug, "debug")
	logger.LogContext(ctx, contract.LogLevelInfo, "info")
	logger.LogContext(ctx, contract.LogLevelWarn, "warn")
	logger.LogContext(ctx, contract.LogLevelError, "error")
	logger.LogContext(ctx, contract.LogLevel("invalid"), "unknown")
}

func TestNewLoggerWithNilDriverDiscardsRecords(t *testing.T) {
	t.Parallel()

	logger := contract.NewLogger(nil)

	require.NotPanics(t, func() {
		logger.Info("discarded")
		logger.With("component", "test").Error("also discarded")
	})
}
