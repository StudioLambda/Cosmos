package framework_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/studiolambda/cosmos/contract"
	"github.com/studiolambda/cosmos/contract/response"
	"github.com/studiolambda/cosmos/framework"

	"github.com/stretchr/testify/require"
)

func handler(w http.ResponseWriter, r *http.Request) error {
	return response.Status(w, http.StatusOK)
}

func TestHandlerRecordReturnsResponse(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)
	res := framework.Handler(handler).Record(req)

	require.Equal(t, http.StatusOK, res.StatusCode)
}

func TestServeHTTPNoContent(t *testing.T) {
	t.Parallel()

	h := framework.Handler(func(w http.ResponseWriter, r *http.Request) error {
		return nil
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNoContent, rec.Code)
}

func TestServeHTTPHandlerReturnsError(t *testing.T) {
	t.Parallel()

	h := framework.Handler(func(w http.ResponseWriter, r *http.Request) error {
		return errors.New("something went wrong")
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	require.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestServeHTTPContextCanceled(t *testing.T) {
	t.Parallel()

	h := framework.Handler(func(w http.ResponseWriter, r *http.Request) error {
		return context.Canceled
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	require.Equal(t, 499, rec.Code)
}

func TestServeHTTPContextDeadlineExceeded(t *testing.T) {
	t.Parallel()

	h := framework.Handler(func(w http.ResponseWriter, r *http.Request) error {
		return context.DeadlineExceeded
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	require.Equal(t, 499, rec.Code)
}

type statusError struct {
	status int
}

func (err statusError) Error() string {
	return fmt.Sprintf("status: %d", err.status)
}

func (err statusError) HTTPStatus() int {
	return err.status
}

func TestServeHTTPCustomHTTPStatus(t *testing.T) {
	t.Parallel()

	h := framework.Handler(func(w http.ResponseWriter, r *http.Request) error {
		return statusError{status: http.StatusTeapot}
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	require.Equal(t, http.StatusTeapot, rec.Code)
}

type handlerError struct{}

func (err handlerError) Error() string {
	return "handler error"
}

func (err handlerError) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusConflict)
}

func TestServeHTTPErrorImplementsHandler(t *testing.T) {
	t.Parallel()

	h := framework.Handler(func(w http.ResponseWriter, r *http.Request) error {
		return handlerError{}
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	require.Equal(t, http.StatusConflict, rec.Code)
}

func TestServeHTTPErrorAfterPartialWrite(t *testing.T) {
	t.Parallel()

	h := framework.Handler(func(w http.ResponseWriter, r *http.Request) error {
		w.WriteHeader(http.StatusOK)
		return errors.New("late error")
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	// The status should remain 200 since WriteHeader was already called.
	// The error is logged but not written.
	require.Equal(t, http.StatusOK, rec.Code)
}

func TestServeHTTPAfterResponseHooksRun(t *testing.T) {
	t.Parallel()

	var hookCalled atomic.Bool
	var receivedErr atomic.Value

	h := framework.Handler(func(w http.ResponseWriter, r *http.Request) error {
		hooks := r.Context().Value(contract.HooksKey).(contract.Hooks)
		hooks.AfterResponse(func(err error) {
			hookCalled.Store(true)
			receivedErr.Store(err)
		})

		return errors.New("test error")
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	require.True(t, hookCalled.Load())
	require.EqualError(t, receivedErr.Load().(error), "test error")
}

func TestServeHTTPAfterResponseHookPanicRecovered(t *testing.T) {
	t.Parallel()

	h := framework.Handler(func(w http.ResponseWriter, r *http.Request) error {
		hooks := r.Context().Value(contract.HooksKey).(contract.Hooks)
		hooks.AfterResponse(func(err error) {
			panic("hook panic")
		})

		return nil
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	// Should not panic.
	require.NotPanics(t, func() {
		h.ServeHTTP(rec, req)
	})
}

func TestServeHTTPHooksInContext(t *testing.T) {
	t.Parallel()

	var foundHooks atomic.Bool

	h := framework.Handler(func(w http.ResponseWriter, r *http.Request) error {
		hooks, ok := r.Context().Value(contract.HooksKey).(contract.Hooks)
		foundHooks.Store(ok && hooks != nil)

		return nil
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	require.True(t, foundHooks.Load())
}

func TestServeHTTPAfterResponseHookReceivesNilOnSuccess(t *testing.T) {
	t.Parallel()

	var hookCalled atomic.Bool
	errChan := make(chan error, 1)

	h := framework.Handler(func(w http.ResponseWriter, r *http.Request) error {
		hooks := r.Context().Value(contract.HooksKey).(contract.Hooks)
		hooks.AfterResponse(func(err error) {
			hookCalled.Store(true)
			errChan <- err
		})

		return response.Status(w, http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	require.True(t, hookCalled.Load())
	require.Nil(t, <-errChan)
}

func TestServeHTTPWrappedContextCanceled(t *testing.T) {
	t.Parallel()

	h := framework.Handler(func(w http.ResponseWriter, r *http.Request) error {
		return fmt.Errorf("request failed: %w", context.Canceled)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	require.Equal(t, 499, rec.Code)
}
