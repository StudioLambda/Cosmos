package framework_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/studiolambda/cosmos/contract"
	"github.com/studiolambda/cosmos/contract/request"
	"github.com/studiolambda/cosmos/framework"

	"github.com/stretchr/testify/require"
)

func TestHTTPAdaptsStandardHandler(t *testing.T) {
	t.Parallel()

	handler := framework.HTTP(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/health", r.URL.Path)
		w.WriteHeader(http.StatusNoContent)
	}))

	response := handler.Record(httptest.NewRequest(http.MethodGet, "/health", nil))

	require.Equal(t, http.StatusNoContent, response.StatusCode)
}

func TestHandlerDoesNotPanicWhenErrorFollowsResponseWrite(t *testing.T) {
	t.Parallel()

	handler := framework.Handler(func(w http.ResponseWriter, _ *http.Request) error {
		w.WriteHeader(http.StatusOK)

		return errors.New("handler failed")
	})

	require.NotPanics(t, func() {
		handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))
	})
}

func TestHandlerDoesNotPanicWhenAfterResponseHookPanics(t *testing.T) {
	t.Parallel()

	handler := framework.Handler(func(_ http.ResponseWriter, r *http.Request) error {
		request.Hooks(r).AfterResponse(func(error) {
			panic("hook failed")
		})

		return nil
	})

	require.NotPanics(t, func() {
		handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))
	})
}

func TestHandlerInitializesRequestLogger(t *testing.T) {
	t.Parallel()

	handler := framework.Handler(func(_ http.ResponseWriter, r *http.Request) error {
		logger := contract.NewLogger(nil)

		request.SetLogger(r, logger)

		require.Same(t, logger, request.Logger(r))

		return nil
	})

	handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))
}
