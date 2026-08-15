package request_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/studiolambda/cosmos/contract"
	"github.com/studiolambda/cosmos/contract/request"

	"github.com/stretchr/testify/require"
)

func TestLoggerReturnsValueFromContext(t *testing.T) {
	t.Parallel()

	pointer := new(atomic.Pointer[contract.Logger])
	logger := contract.NewLogger(nil)
	pointer.Store(logger)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = req.WithContext(context.WithValue(req.Context(), contract.LoggerKey, pointer))

	require.Same(t, logger, request.Logger(req))
}

func TestLoggerReturnsDiscardLoggerWhenContextMissing(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	logger := request.Logger(req)

	require.NotNil(t, logger)
	require.NotPanics(t, func() {
		logger.InfoContext(req.Context(), "ignored")
	})
}

func TestSetLoggerReplacesContextLogger(t *testing.T) {
	t.Parallel()

	pointer := new(atomic.Pointer[contract.Logger])
	pointer.Store(contract.NewLogger(nil))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = req.WithContext(context.WithValue(req.Context(), contract.LoggerKey, pointer))
	logger := contract.NewLogger(nil)

	request.SetLogger(req, logger)

	require.Same(t, logger, request.Logger(req))
}
