package middleware_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/studiolambda/cosmos/contract/request"
	"github.com/studiolambda/cosmos/framework"
	correlation "github.com/studiolambda/cosmos/framework/middleware"
	"github.com/studiolambda/cosmos/problem"

	"github.com/stretchr/testify/require"
)

func TestMiddlewareGeneratesNewID(t *testing.T) {
	t.Parallel()

	var captured string

	handler := correlation.Correlation(correlation.DefaultCorrelationConfig)(framework.Handler(func(
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

	handler := correlation.Correlation(correlation.DefaultCorrelationConfig)(framework.Handler(func(
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

	handler := correlation.Correlation(correlation.DefaultCorrelationConfig)(framework.Handler(func(
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

	handler := correlation.Correlation(correlation.DefaultCorrelationConfig)(framework.Handler(func(
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

	handler := correlation.Correlation(correlation.DefaultCorrelationConfig)(framework.Handler(func(
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

	handler := correlation.Correlation(correlation.DefaultCorrelationConfig)(framework.Handler(func(
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

	handler := correlation.CorrelationWith(correlation.CorrelationConfig{
		Header: "X-Request-ID",
	}, nil)(framework.Handler(func(
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

	handler := correlation.CorrelationWith(correlation.CorrelationConfig{}, func() string {
		return "generated-safe-id"
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

	handler := correlation.CorrelationWith(correlation.CorrelationConfig{}, func() string {
		return "generated-safe-id"
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

	handler := correlation.CorrelationWith(correlation.CorrelationConfig{}, func() string {
		return "custom-generated"
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

func TestMiddlewareInvalidGeneratedIDFallsBackAndContinues(t *testing.T) {
	t.Parallel()

	handler := correlation.CorrelationWith(correlation.CorrelationConfig{}, func() string {
		return "bad\nvalue"
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

func TestCorrelationStoresIDInRequestContext(t *testing.T) {
	t.Parallel()

	var fromHelper string
	handler := correlation.Correlation(correlation.DefaultCorrelationConfig)(framework.Handler(func(
		w http.ResponseWriter,
		r *http.Request,
	) error {
		fromHelper = request.CorrelationID(r)
		w.WriteHeader(http.StatusOK)

		return nil
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	handler.Record(req)

	require.NotEmpty(t, fromHelper)
}

func TestMiddlewareAddsCorrelationIDToProblemDetailsByDefault(t *testing.T) {
	t.Parallel()

	handler := correlation.Correlation(correlation.DefaultCorrelationConfig)(framework.Handler(func(
		w http.ResponseWriter,
		r *http.Request,
	) error {
		return problem.Details{Status: http.StatusBadRequest, Title: "bad request"}
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept", "application/problem+json")
	res := handler.Record(req)
	var body map[string]any
	require.NoError(t, json.NewDecoder(res.Body).Decode(&body))

	require.Equal(t, http.StatusBadRequest, res.StatusCode)
	require.Equal(t, res.Header.Get("X-Correlation-ID"), body["correlation_id"])
}

func TestMiddlewareDoesNotAddCorrelationIDToProblemDetailsWhenProblemKeyIsEmpty(t *testing.T) {
	t.Parallel()

	handler := correlation.CorrelationWith(correlation.CorrelationConfig{}, func() string {
		return "test-correlation-id"
	})(framework.Handler(func(
		w http.ResponseWriter,
		r *http.Request,
	) error {
		return problem.Details{Status: http.StatusBadRequest, Title: "bad request"}
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept", "application/problem+json")
	res := handler.Record(req)
	var body map[string]any
	require.NoError(t, json.NewDecoder(res.Body).Decode(&body))

	require.Equal(t, http.StatusBadRequest, res.StatusCode)
	_, exists := body["correlation_id"]
	require.False(t, exists)
}
