package middleware_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/studiolambda/cosmos/framework"
	"github.com/studiolambda/cosmos/framework/middleware"

	"github.com/stretchr/testify/require"
)

func TestHTTPAdapterCallsMiddleware(t *testing.T) {
	t.Parallel()

	middlewareCalled := false
	stdMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			middlewareCalled = true
			next.ServeHTTP(w, r)
		})
	}

	handler := middleware.HTTP(stdMiddleware)(framework.Handler(func(
		w http.ResponseWriter,
		r *http.Request,
	) error {
		w.WriteHeader(http.StatusOK)

		return nil
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	res := handler.Record(req)

	require.True(t, middlewareCalled)
	require.Equal(t, http.StatusOK, res.StatusCode)
}

func TestHTTPAdapterCapturesError(t *testing.T) {
	t.Parallel()

	expected := errors.New("handler error")
	stdMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r)
		})
	}

	handler := middleware.HTTP(stdMiddleware)(func(
		w http.ResponseWriter,
		r *http.Request,
	) error {
		return expected
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	err := handler(rec, req)

	require.ErrorIs(t, err, expected)
}

func TestHTTPAdapterPassesThrough(t *testing.T) {
	t.Parallel()

	headerSet := false
	stdMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Custom", "added")
			headerSet = true
			next.ServeHTTP(w, r)
		})
	}

	handler := middleware.HTTP(stdMiddleware)(framework.Handler(func(
		w http.ResponseWriter,
		r *http.Request,
	) error {
		w.WriteHeader(http.StatusOK)

		return nil
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	res := handler.Record(req)

	require.True(t, headerSet)
	require.Equal(t, "added", res.Header.Get("X-Custom"))
	require.Equal(t, http.StatusOK, res.StatusCode)
}

func TestHTTPAdapterReturnsNilOnSuccess(t *testing.T) {
	t.Parallel()

	stdMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r)
		})
	}

	handler := middleware.HTTP(stdMiddleware)(func(
		w http.ResponseWriter,
		r *http.Request,
	) error {
		return nil
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	err := handler(rec, req)

	require.NoError(t, err)
}
