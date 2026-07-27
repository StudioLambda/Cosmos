package contract_test

import (
	"context"
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

func TestNewLoggerWithNilDriverDiscardsRecords(t *testing.T) {
	t.Parallel()

	logger := contract.NewLogger(nil)

	require.NotPanics(t, func() {
		logger.Info("discarded")
		logger.With("component", "test").Error("also discarded")
	})
}
