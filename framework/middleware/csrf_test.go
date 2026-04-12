package middleware_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/studiolambda/cosmos/framework"
	"github.com/studiolambda/cosmos/framework/middleware"
	"github.com/studiolambda/cosmos/problem"

	"github.com/stretchr/testify/require"
)

func TestCSRFAllowsSameOriginRequest(t *testing.T) {
	t.Parallel()

	called := false
	handler := middleware.CSRF()(framework.Handler(func(
		w http.ResponseWriter,
		r *http.Request,
	) error {
		called = true

		return nil
	}))

	req := httptest.NewRequest(http.MethodPost, "/submit", nil)
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	res := handler.Record(req)

	require.True(t, called)
	require.Equal(t, http.StatusNoContent, res.StatusCode)
}

func TestCSRFBlocksCrossOriginPost(t *testing.T) {
	t.Parallel()

	handler := middleware.CSRF()(framework.Handler(func(
		w http.ResponseWriter,
		r *http.Request,
	) error {
		t.Fatal("handler should not be called for blocked request")

		return nil
	}))

	req := httptest.NewRequest(http.MethodPost, "/submit", nil)
	req.Header.Set("Sec-Fetch-Site", "cross-site")
	res := handler.Record(req)

	require.Equal(t, http.StatusForbidden, res.StatusCode)
}

func TestCSRFWithCustomError(t *testing.T) {
	t.Parallel()

	customErr := problem.Problem{
		Title:  "Custom CSRF Error",
		Detail: "Custom detail",
		Status: http.StatusUnauthorized,
	}

	csrf := http.NewCrossOriginProtection()
	handler := middleware.CSRFWith(
		csrf, customErr,
	)(framework.Handler(func(
		w http.ResponseWriter,
		r *http.Request,
	) error {
		t.Fatal("handler should not be called for blocked request")

		return nil
	}))

	req := httptest.NewRequest(http.MethodPost, "/submit", nil)
	req.Header.Set("Sec-Fetch-Site", "cross-site")
	res := handler.Record(req)

	require.Equal(t, http.StatusUnauthorized, res.StatusCode)
}

func TestCSRFWithTrustedOrigin(t *testing.T) {
	t.Parallel()

	called := false
	handler := middleware.CSRF(
		"https://trusted.example.com",
	)(framework.Handler(func(
		w http.ResponseWriter,
		r *http.Request,
	) error {
		called = true

		return nil
	}))

	req := httptest.NewRequest(http.MethodPost, "/submit", nil)
	req.Host = "myapp.example.com"
	req.Header.Set("Origin", "https://trusted.example.com")
	res := handler.Record(req)

	require.True(t, called)
	require.Equal(t, http.StatusNoContent, res.StatusCode)
}

func TestCSRFBlockedErrorWrapsOriginal(t *testing.T) {
	t.Parallel()

	var captured error
	handler := middleware.CSRF()(func(
		w http.ResponseWriter,
		r *http.Request,
	) error {
		t.Fatal("handler should not be called")

		return nil
	})

	req := httptest.NewRequest(http.MethodPost, "/submit", nil)
	req.Header.Set("Sec-Fetch-Site", "cross-site")
	rec := httptest.NewRecorder()

	captured = handler(rec, req)

	require.Error(t, captured)
	require.True(t, errors.As(captured, &problem.Problem{}))
}
