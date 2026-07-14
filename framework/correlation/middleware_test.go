package correlation_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/studiolambda/cosmos/contract/request"
	"github.com/studiolambda/cosmos/framework"
	"github.com/studiolambda/cosmos/framework/correlation"

	"github.com/stretchr/testify/require"
)

func TestMiddlewareGeneratesNewID(t *testing.T) {
	t.Parallel()

	var captured string

	handler := correlation.Middleware()(framework.Handler(func(
		w http.ResponseWriter,
		r *http.Request,
	) error {
		captured = request.CorrelationID(r)
		w.WriteHeader(http.StatusOK)

		return nil
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	res := handler.Record(req)

	require.Len(t, captured, 32)
	require.Equal(t, http.StatusOK, res.StatusCode)
	require.Equal(t, captured, res.Header.Get("X-Correlation-ID"))
}

func TestMiddlewareUsesExistingHeader(t *testing.T) {
	t.Parallel()

	var captured string

	handler := correlation.Middleware()(framework.Handler(func(
		w http.ResponseWriter,
		r *http.Request,
	) error {
		captured = request.CorrelationID(r)
		w.WriteHeader(http.StatusOK)

		return nil
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Correlation-ID", "existing-id")
	res := handler.Record(req)

	require.Equal(t, "existing-id", captured)
	require.Equal(t, "existing-id", res.Header.Get("X-Correlation-ID"))
}

func TestMiddlewareExtractsFromTraceparent(t *testing.T) {
	t.Parallel()

	var captured string

	handler := correlation.Middleware()(framework.Handler(func(
		w http.ResponseWriter,
		r *http.Request,
	) error {
		captured = request.CorrelationID(r)
		w.WriteHeader(http.StatusOK)

		return nil
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Traceparent", "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01")
	res := handler.Record(req)

	require.Equal(t, "4bf92f3577b34da6a3ce929d0e0e4736", captured)
	require.Equal(t, "4bf92f3577b34da6a3ce929d0e0e4736", res.Header.Get("X-Correlation-ID"))
}

func TestMiddlewareTraceparentTakesPrecedenceOverHeader(t *testing.T) {
	t.Parallel()

	var captured string

	handler := correlation.Middleware()(framework.Handler(func(
		w http.ResponseWriter,
		r *http.Request,
	) error {
		captured = request.CorrelationID(r)
		w.WriteHeader(http.StatusOK)

		return nil
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Traceparent", "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01")
	req.Header.Set("X-Correlation-ID", "should-be-ignored")
	handler.Record(req)

	require.Equal(t, "4bf92f3577b34da6a3ce929d0e0e4736", captured)
}

func TestMiddlewareIgnoresInvalidTraceparent(t *testing.T) {
	t.Parallel()

	var captured string

	handler := correlation.Middleware()(framework.Handler(func(
		w http.ResponseWriter,
		r *http.Request,
	) error {
		captured = request.CorrelationID(r)
		w.WriteHeader(http.StatusOK)

		return nil
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Traceparent", "invalid-format")
	req.Header.Set("X-Correlation-ID", "fallback-id")
	handler.Record(req)

	require.Equal(t, "fallback-id", captured)
}

func TestMiddlewareIgnoresAllZerosTraceID(t *testing.T) {
	t.Parallel()

	var captured string

	handler := correlation.Middleware()(framework.Handler(func(
		w http.ResponseWriter,
		r *http.Request,
	) error {
		captured = request.CorrelationID(r)
		w.WriteHeader(http.StatusOK)

		return nil
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Traceparent", "00-00000000000000000000000000000000-00f067aa0ba902b7-01")
	handler.Record(req)

	// Should generate a new ID since all-zeros is invalid
	require.Len(t, captured, 32)
	require.NotEqual(t, "00000000000000000000000000000000", captured)
}

func TestMiddlewareWithCustomHeader(t *testing.T) {
	t.Parallel()

	var captured string

	handler := correlation.MiddlewareWith(correlation.MiddlewareConfig{
		Header: "X-Request-ID",
	})(framework.Handler(func(
		w http.ResponseWriter,
		r *http.Request,
	) error {
		captured = request.CorrelationID(r)
		w.WriteHeader(http.StatusOK)

		return nil
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Request-ID", "custom-id")
	res := handler.Record(req)

	require.Equal(t, "custom-id", captured)
	require.Equal(t, "custom-id", res.Header.Get("X-Request-ID"))
}

func TestMiddlewareRejectsUnsafeHeaderAndGenerates(t *testing.T) {
	t.Parallel()

	var captured string

	handler := correlation.MiddlewareWith(correlation.MiddlewareConfig{
		Generate: func() (string, error) {
			return "generated-safe-id", nil
		},
	})(framework.Handler(func(
		w http.ResponseWriter,
		r *http.Request,
	) error {
		captured = request.CorrelationID(r)
		w.WriteHeader(http.StatusOK)

		return nil
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Correlation-ID", "bad\nvalue")
	res := handler.Record(req)

	require.Equal(t, "generated-safe-id", captured)
	require.Equal(t, "generated-safe-id", res.Header.Get("X-Correlation-ID"))
}

func TestMiddlewareRejectsOverlongHeaderAndGenerates(t *testing.T) {
	t.Parallel()

	var captured string

	handler := correlation.MiddlewareWith(correlation.MiddlewareConfig{
		Generate: func() (string, error) {
			return "generated-safe-id", nil
		},
	})(framework.Handler(func(
		w http.ResponseWriter,
		r *http.Request,
	) error {
		captured = request.CorrelationID(r)
		w.WriteHeader(http.StatusOK)

		return nil
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Correlation-ID", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	res := handler.Record(req)

	require.Equal(t, "generated-safe-id", captured)
	require.Equal(t, "generated-safe-id", res.Header.Get("X-Correlation-ID"))
}

func TestMiddlewareWithCustomGenerator(t *testing.T) {
	t.Parallel()

	var captured string

	handler := correlation.MiddlewareWith(correlation.MiddlewareConfig{
		Generate: func() (string, error) {
			return "custom-generated", nil
		},
	})(framework.Handler(func(
		w http.ResponseWriter,
		r *http.Request,
	) error {
		captured = request.CorrelationID(r)
		w.WriteHeader(http.StatusOK)

		return nil
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	handler.Record(req)

	require.Equal(t, "custom-generated", captured)
}

func TestMiddlewareGeneratorErrorFallsBackAndContinues(t *testing.T) {
	t.Parallel()

	handler := correlation.MiddlewareWith(correlation.MiddlewareConfig{
		Generate: func() (string, error) {
			return "", errors.New("generator failed")
		},
	})(framework.Handler(func(
		w http.ResponseWriter,
		r *http.Request,
	) error {
		w.WriteHeader(http.StatusOK)

		return nil
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	res := handler.Record(req)
	id := res.Header.Get("X-Correlation-ID")

	require.Equal(t, http.StatusOK, res.StatusCode)
	require.Len(t, id, 32)
}

func TestMiddlewareInvalidGeneratedIDFallsBackAndContinues(t *testing.T) {
	t.Parallel()

	handler := correlation.MiddlewareWith(correlation.MiddlewareConfig{
		Generate: func() (string, error) {
			return "bad\nvalue", nil
		},
	})(framework.Handler(func(
		w http.ResponseWriter,
		r *http.Request,
	) error {
		w.WriteHeader(http.StatusOK)

		return nil
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	res := handler.Record(req)
	id := res.Header.Get("X-Correlation-ID")

	require.Equal(t, http.StatusOK, res.StatusCode)
	require.Len(t, id, 32)
}

func TestFromMatchesRequestHelper(t *testing.T) {
	t.Parallel()

	var fromHelper string
	var fromPackage string

	handler := correlation.Middleware()(framework.Handler(func(
		w http.ResponseWriter,
		r *http.Request,
	) error {
		fromHelper = request.CorrelationID(r)
		fromPackage = correlation.From(r)
		w.WriteHeader(http.StatusOK)

		return nil
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	handler.Record(req)

	require.Equal(t, fromHelper, fromPackage)
	require.NotEmpty(t, fromHelper)
}
