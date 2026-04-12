package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/studiolambda/cosmos/framework"
	"github.com/studiolambda/cosmos/framework/middleware"
)

func TestCORSPreflightSetsHeaders(t *testing.T) {
	t.Parallel()

	handler := middleware.CORS(middleware.CORSOptions{
		AllowedOrigins:   []string{"https://example.com"},
		AllowedMethods:   []string{"GET", "POST"},
		AllowedHeaders:   []string{"Content-Type"},
		AllowCredentials: true,
		MaxAge:           600,
	})(framework.Handler(func(
		w http.ResponseWriter,
		r *http.Request,
	) error {
		t.Fatal("handler should not be called for preflight")

		return nil
	}))

	req := httptest.NewRequest(http.MethodOptions, "/api/data", nil)
	req.Header.Set("Origin", "https://example.com")
	res := handler.Record(req)

	require.Equal(t, http.StatusNoContent, res.StatusCode)
	require.Equal(
		t,
		"https://example.com",
		res.Header.Get("Access-Control-Allow-Origin"),
	)
	require.Equal(
		t,
		"GET, POST",
		res.Header.Get("Access-Control-Allow-Methods"),
	)
	require.Equal(
		t,
		"Content-Type",
		res.Header.Get("Access-Control-Allow-Headers"),
	)
	require.Equal(
		t,
		"true",
		res.Header.Get("Access-Control-Allow-Credentials"),
	)
	require.Equal(
		t,
		"600",
		res.Header.Get("Access-Control-Max-Age"),
	)
	require.Equal(
		t,
		"Origin",
		res.Header.Get("Vary"),
	)
}

func TestCORSDisallowedOriginSkipsHeaders(t *testing.T) {
	t.Parallel()

	called := false
	handler := middleware.CORS(middleware.CORSOptions{
		AllowedOrigins: []string{"https://trusted.com"},
		AllowedMethods: []string{"GET"},
	})(framework.Handler(func(
		w http.ResponseWriter,
		r *http.Request,
	) error {
		called = true

		return nil
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "https://evil.com")
	res := handler.Record(req)

	require.True(t, called)
	require.Empty(t, res.Header.Get("Access-Control-Allow-Origin"))
}

func TestCORSNoOriginHeaderPassesThrough(t *testing.T) {
	t.Parallel()

	called := false
	handler := middleware.CORS(
		middleware.DefaultCORSOptions,
	)(framework.Handler(func(
		w http.ResponseWriter,
		r *http.Request,
	) error {
		called = true

		return nil
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	res := handler.Record(req)

	require.True(t, called)
	require.Empty(t, res.Header.Get("Access-Control-Allow-Origin"))
}

func TestCORSWildcardOriginSetsStarHeader(t *testing.T) {
	t.Parallel()

	handler := middleware.CORS(middleware.CORSOptions{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET"},
	})(framework.Handler(func(
		w http.ResponseWriter,
		r *http.Request,
	) error {
		return nil
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "https://any.com")
	res := handler.Record(req)

	require.Equal(
		t,
		"*",
		res.Header.Get("Access-Control-Allow-Origin"),
	)
	require.Empty(t, res.Header.Get("Vary"))
}

func TestCORSExposedHeaders(t *testing.T) {
	t.Parallel()

	handler := middleware.CORS(middleware.CORSOptions{
		AllowedOrigins: []string{"*"},
		ExposedHeaders: []string{"X-Request-Id", "X-Total-Count"},
	})(framework.Handler(func(
		w http.ResponseWriter,
		r *http.Request,
	) error {
		return nil
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "https://any.com")
	res := handler.Record(req)

	require.Equal(
		t,
		"X-Request-Id, X-Total-Count",
		res.Header.Get("Access-Control-Expose-Headers"),
	)
}

func TestCORSPanicsOnCredentialsWithWildcard(t *testing.T) {
	t.Parallel()

	require.PanicsWithValue(
		t,
		"cors: AllowCredentials must not be used with wildcard AllowedOrigins",
		func() {
			middleware.CORS(middleware.CORSOptions{
				AllowedOrigins:   []string{"*"},
				AllowCredentials: true,
			})
		},
	)
}

func TestCORSDoesNotPanicOnCredentialsWithExplicitOrigins(t *testing.T) {
	t.Parallel()

	require.NotPanics(t, func() {
		middleware.CORS(middleware.CORSOptions{
			AllowedOrigins:   []string{"https://example.com"},
			AllowCredentials: true,
		})
	})
}
