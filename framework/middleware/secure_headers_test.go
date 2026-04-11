package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/studiolambda/cosmos/framework"
	"github.com/studiolambda/cosmos/framework/middleware"
)

func TestSecureHeadersDefault(t *testing.T) {
	t.Parallel()

	handler := middleware.SecureHeaders()(framework.Handler(func(
		w http.ResponseWriter,
		r *http.Request,
	) error {
		w.WriteHeader(http.StatusOK)

		return nil
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	res := handler.Record(req)

	require.Equal(t, http.StatusOK, res.StatusCode)
	require.Equal(
		t, "nosniff", res.Header.Get("X-Content-Type-Options"),
	)
	require.Equal(
		t, "DENY", res.Header.Get("X-Frame-Options"),
	)
	require.Equal(
		t,
		"strict-origin-when-cross-origin",
		res.Header.Get("Referrer-Policy"),
	)
	require.Equal(
		t, "1; mode=block", res.Header.Get("X-XSS-Protection"),
	)
	require.Equal(
		t,
		"max-age=63072000; includeSubDomains",
		res.Header.Get("Strict-Transport-Security"),
	)
	require.Empty(t, res.Header.Get("Content-Security-Policy"))
	require.Empty(t, res.Header.Get("Permissions-Policy"))
}

func TestSecureHeadersWithCustomOptions(t *testing.T) {
	t.Parallel()

	handler := middleware.SecureHeadersWith(
		middleware.SecureHeadersOptions{
			ContentTypeOptions: "nosniff",
			FrameOptions:       "SAMEORIGIN",
			ReferrerPolicy:     "no-referrer",
		},
	)(framework.Handler(func(
		w http.ResponseWriter,
		r *http.Request,
	) error {
		w.WriteHeader(http.StatusOK)

		return nil
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	res := handler.Record(req)

	require.Equal(t, http.StatusOK, res.StatusCode)
	require.Equal(
		t, "nosniff", res.Header.Get("X-Content-Type-Options"),
	)
	require.Equal(
		t, "SAMEORIGIN", res.Header.Get("X-Frame-Options"),
	)
	require.Equal(
		t, "no-referrer", res.Header.Get("Referrer-Policy"),
	)
	require.Empty(t, res.Header.Get("X-XSS-Protection"))
	require.Empty(t, res.Header.Get("Strict-Transport-Security"))
}

func TestSecureHeadersEmptyOptionsSkipsHeaders(t *testing.T) {
	t.Parallel()

	handler := middleware.SecureHeadersWith(
		middleware.SecureHeadersOptions{},
	)(framework.Handler(func(
		w http.ResponseWriter,
		r *http.Request,
	) error {
		w.WriteHeader(http.StatusOK)

		return nil
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	res := handler.Record(req)

	require.Equal(t, http.StatusOK, res.StatusCode)
	require.Empty(t, res.Header.Get("X-Content-Type-Options"))
	require.Empty(t, res.Header.Get("X-Frame-Options"))
	require.Empty(t, res.Header.Get("Referrer-Policy"))
	require.Empty(t, res.Header.Get("X-XSS-Protection"))
	require.Empty(t, res.Header.Get("Strict-Transport-Security"))
	require.Empty(t, res.Header.Get("Content-Security-Policy"))
	require.Empty(t, res.Header.Get("Permissions-Policy"))
}

func TestSecureHeadersAllOptions(t *testing.T) {
	t.Parallel()

	handler := middleware.SecureHeadersWith(
		middleware.SecureHeadersOptions{
			ContentTypeOptions:      "nosniff",
			FrameOptions:            "DENY",
			ReferrerPolicy:          "no-referrer",
			XSSProtection:           "0",
			StrictTransportSecurity: "max-age=31536000",
			ContentSecurityPolicy:   "default-src 'self'",
			PermissionsPolicy:       "camera=(), microphone=()",
		},
	)(framework.Handler(func(
		w http.ResponseWriter,
		r *http.Request,
	) error {
		w.WriteHeader(http.StatusOK)

		return nil
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	res := handler.Record(req)

	require.Equal(t, http.StatusOK, res.StatusCode)
	require.Equal(
		t, "nosniff", res.Header.Get("X-Content-Type-Options"),
	)
	require.Equal(t, "DENY", res.Header.Get("X-Frame-Options"))
	require.Equal(
		t, "no-referrer", res.Header.Get("Referrer-Policy"),
	)
	require.Equal(t, "0", res.Header.Get("X-XSS-Protection"))
	require.Equal(
		t,
		"max-age=31536000",
		res.Header.Get("Strict-Transport-Security"),
	)
	require.Equal(
		t,
		"default-src 'self'",
		res.Header.Get("Content-Security-Policy"),
	)
	require.Equal(
		t,
		"camera=(), microphone=()",
		res.Header.Get("Permissions-Policy"),
	)
}

func TestSecureHeadersCallsNextHandler(t *testing.T) {
	t.Parallel()

	called := false
	handler := middleware.SecureHeaders()(framework.Handler(func(
		w http.ResponseWriter,
		r *http.Request,
	) error {
		called = true

		return nil
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	handler.Record(req)

	require.True(t, called)
}
