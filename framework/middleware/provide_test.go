package middleware_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/studiolambda/cosmos/framework"
	"github.com/studiolambda/cosmos/framework/middleware"

	"github.com/stretchr/testify/require"
)

type contextKey string

func TestProvideAddsValueToContext(t *testing.T) {
	t.Parallel()

	key := contextKey("db")
	val := "my-database"

	var captured any
	handler := middleware.Provide(key, val)(framework.Handler(func(
		w http.ResponseWriter,
		r *http.Request,
	) error {
		captured = r.Context().Value(key)

		return nil
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	res := handler.Record(req)

	require.Equal(t, http.StatusNoContent, res.StatusCode)
	require.Equal(t, "my-database", captured)
}

func TestProvideWithDynamicValue(t *testing.T) {
	t.Parallel()

	key := contextKey("request-id")

	var captured any
	handler := middleware.ProvideWith(func(
		w http.ResponseWriter,
		r *http.Request,
	) (context.Context, error) {
		return context.WithValue(
			r.Context(), key, r.Header.Get("X-Request-ID"),
		), nil
	})(framework.Handler(func(
		w http.ResponseWriter,
		r *http.Request,
	) error {
		captured = r.Context().Value(key)

		return nil
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Request-ID", "abc-123")
	res := handler.Record(req)

	require.Equal(t, http.StatusNoContent, res.StatusCode)
	require.Equal(t, "abc-123", captured)
}

func TestProvideWithReturnsErrorOnFailure(t *testing.T) {
	t.Parallel()

	expected := errors.New("provider failed")

	handler := middleware.ProvideWith(func(
		w http.ResponseWriter,
		r *http.Request,
	) (context.Context, error) {
		return nil, expected
	})(func(
		w http.ResponseWriter,
		r *http.Request,
	) error {
		t.Fatal("handler should not be called when provider fails")

		return nil
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	err := handler(rec, req)

	require.ErrorIs(t, err, expected)
}
