package middleware_test

import (
	"bytes"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/studiolambda/cosmos/framework"
	"github.com/studiolambda/cosmos/framework/middleware"

	"github.com/stretchr/testify/require"
)

func TestLoggerNilLoggerDoesNotPanic(t *testing.T) {
	t.Parallel()

	handler := middleware.Logger(nil)(framework.Handler(func(
		w http.ResponseWriter,
		r *http.Request,
	) error {
		w.WriteHeader(http.StatusOK)

		return nil
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	res := handler.Record(req)

	require.Equal(t, http.StatusOK, res.StatusCode)
}

func TestLoggerLogsOnHandlerError(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, nil))

	handler := middleware.Logger(logger)(framework.Handler(func(
		w http.ResponseWriter,
		r *http.Request,
	) error {
		return errors.New("handler failure")
	}))

	req := httptest.NewRequest(http.MethodGet, "/test-path", nil)
	handler.Record(req)

	output := buf.String()
	require.Contains(t, output, "request failed")
	require.Contains(t, output, "handler failure")
	require.Contains(t, output, "GET")
}

func TestLoggerLogsOn5xxStatus(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, nil))

	handler := middleware.Logger(logger)(framework.Handler(func(
		w http.ResponseWriter,
		r *http.Request,
	) error {
		w.WriteHeader(http.StatusInternalServerError)

		return nil
	}))

	req := httptest.NewRequest(http.MethodGet, "/server-error", nil)
	handler.Record(req)

	output := buf.String()
	require.Contains(t, output, "request failed")
	require.Contains(t, output, "500")
}

func TestLoggerDoesNotLogOnSuccess(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, nil))

	handler := middleware.Logger(logger)(framework.Handler(func(
		w http.ResponseWriter,
		r *http.Request,
	) error {
		w.WriteHeader(http.StatusOK)

		return nil
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	handler.Record(req)

	require.Empty(t, buf.String())
}

func TestLoggerDoesNotLogOn4xxStatus(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, nil))

	handler := middleware.Logger(logger)(framework.Handler(func(
		w http.ResponseWriter,
		r *http.Request,
	) error {
		w.WriteHeader(http.StatusNotFound)

		return nil
	}))

	req := httptest.NewRequest(http.MethodGet, "/missing", nil)
	handler.Record(req)

	require.Empty(t, buf.String())
}

func TestLoggerIncludesURLInLog(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, nil))

	handler := middleware.Logger(logger)(framework.Handler(func(
		w http.ResponseWriter,
		r *http.Request,
	) error {
		return errors.New("some error")
	}))

	req := httptest.NewRequest(
		http.MethodPost, "/api/resource?key=val", nil,
	)
	handler.Record(req)

	output := buf.String()
	require.Contains(t, output, "/api/resource")
	require.Contains(t, output, "POST")
}
