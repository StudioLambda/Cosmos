package middleware_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/studiolambda/cosmos/framework"
	"github.com/studiolambda/cosmos/framework/middleware"
	"github.com/studiolambda/cosmos/problem"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
)

func TestOpenTelemetryExtractsRemoteParent(t *testing.T) {
	t.Parallel()

	exporter := tracetest.NewInMemoryExporter()
	provider := trace.NewTracerProvider(trace.WithSyncer(exporter))
	t.Cleanup(func() { require.NoError(t, provider.Shutdown(context.Background())) })
	handler := middleware.OpenTelemetry(middleware.OpenTelemetryConfig{TracerProvider: provider})(framework.Handler(func(
		w http.ResponseWriter,
		r *http.Request,
	) error {
		w.WriteHeader(http.StatusNoContent)

		return nil
	}))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Traceparent", "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01")

	handler.Record(req)

	spans := exporter.GetSpans()
	require.Len(t, spans, 1)
	require.Equal(t, "4bf92f3577b34da6a3ce929d0e0e4736", spans[0].SpanContext.TraceID().String())
	require.Equal(t, "00f067aa0ba902b7", spans[0].Parent.SpanID().String())
}

func TestOpenTelemetryAndCorrelationShareTraceID(t *testing.T) {
	t.Parallel()

	exporter := tracetest.NewInMemoryExporter()
	provider := trace.NewTracerProvider(trace.WithSyncer(exporter))
	t.Cleanup(func() { require.NoError(t, provider.Shutdown(context.Background())) })
	app := framework.New()
	app.Use(
		middleware.OpenTelemetry(middleware.OpenTelemetryConfig{TracerProvider: provider}),
		middleware.Correlation(middleware.DefaultCorrelationConfig()),
	)
	app.Get("/", func(w http.ResponseWriter, r *http.Request) error {
		w.WriteHeader(http.StatusNoContent)

		return nil
	})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Traceparent", "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01")

	response := app.Record(req)

	spans := exporter.GetSpans()
	require.Len(t, spans, 1)
	require.Equal(t, spans[0].SpanContext.TraceID().String(), response.Header.Get(middleware.DefaultHeader))
}

func TestOpenTelemetryUsesRoutePatternAndFinalStatus(t *testing.T) {
	t.Parallel()

	exporter := tracetest.NewInMemoryExporter()
	provider := trace.NewTracerProvider(trace.WithSyncer(exporter))
	t.Cleanup(func() { require.NoError(t, provider.Shutdown(context.Background())) })
	app := framework.New()
	app.Use(middleware.OpenTelemetry(middleware.OpenTelemetryConfig{TracerProvider: provider}))
	app.Get("/users/{id}", func(w http.ResponseWriter, r *http.Request) error {
		return problem.Details{Status: http.StatusBadRequest, Title: "invalid user"}
	})

	response := app.Record(httptest.NewRequest(http.MethodGet, "/users/123?token=secret", nil))

	spans := exporter.GetSpans()
	require.Equal(t, http.StatusBadRequest, response.StatusCode)
	require.Len(t, spans, 1)
	require.Equal(t, "GET /users/{id}", spans[0].Name)
	require.Equal(t, "/users/{id}", otelAttributeValue(spans[0].Attributes, semconv.HTTPRouteKey))
	require.Equal(t, int64(http.StatusBadRequest), otelAttributeValue(spans[0].Attributes, semconv.HTTPResponseStatusCodeKey))
	require.Empty(t, otelAttributeValue(spans[0].Attributes, semconv.URLPathKey))
}

func TestOpenTelemetryMarksRenderedServerErrors(t *testing.T) {
	t.Parallel()

	exporter := tracetest.NewInMemoryExporter()
	provider := trace.NewTracerProvider(trace.WithSyncer(exporter))
	t.Cleanup(func() { require.NoError(t, provider.Shutdown(context.Background())) })
	handler := middleware.OpenTelemetry(middleware.OpenTelemetryConfig{TracerProvider: provider})(framework.Handler(func(
		w http.ResponseWriter,
		r *http.Request,
	) error {
		return errors.New("sensitive internal failure")
	}))

	response := handler.Record(httptest.NewRequest(http.MethodGet, "/", nil))

	spans := exporter.GetSpans()
	require.Equal(t, http.StatusInternalServerError, response.StatusCode)
	require.Len(t, spans, 1)
	require.Equal(t, codes.Error, spans[0].Status.Code)
	require.Equal(t, "HTTP 500", spans[0].Status.Description)
	require.Empty(t, spans[0].Events)
}

func TestOpenTelemetryObservesRecoveredPanic(t *testing.T) {
	t.Parallel()

	exporter := tracetest.NewInMemoryExporter()
	provider := trace.NewTracerProvider(trace.WithSyncer(exporter))
	t.Cleanup(func() { require.NoError(t, provider.Shutdown(context.Background())) })
	handler := middleware.Recover()(middleware.OpenTelemetry(middleware.OpenTelemetryConfig{TracerProvider: provider})(framework.Handler(func(
		w http.ResponseWriter,
		r *http.Request,
	) error {
		panic("unexpected")
	})))

	response := handler.Record(httptest.NewRequest(http.MethodGet, "/", nil))

	spans := exporter.GetSpans()
	require.Equal(t, http.StatusInternalServerError, response.StatusCode)
	require.Len(t, spans, 1)
	require.Equal(t, codes.Error, spans[0].Status.Code)
}

func otelAttributeValue(attributes []attribute.KeyValue, key attribute.Key) any {
	for _, value := range attributes {
		if value.Key == key {
			return value.Value.AsInterface()
		}
	}

	return nil
}
