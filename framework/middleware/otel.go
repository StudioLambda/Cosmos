package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/studiolambda/cosmos/contract/request"
	"github.com/studiolambda/cosmos/framework"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
	"go.opentelemetry.io/otel/trace"
)

const otelInstrumentationName = "github.com/studiolambda/cosmos/framework/middleware"

// OpenTelemetryConfig configures [OpenTelemetry] HTTP tracing middleware.
//
// Example:
//
//	provider := sdktrace.NewTracerProvider(sdktrace.WithBatcher(exporter))
//	app.Use(middleware.OpenTelemetry(middleware.OpenTelemetryConfig{
//		TracerProvider: provider,
//	}))
type OpenTelemetryConfig struct {
	// TracerProvider creates server spans. When nil, the global provider is used.
	TracerProvider trace.TracerProvider

	// Propagator extracts remote parent context from incoming request headers.
	// When nil, W3C Trace Context propagation is used.
	Propagator propagation.TextMapPropagator
}

// OpenTelemetry returns middleware that creates an OpenTelemetry server span for
// each request. Spans finish after Cosmos has rendered returned handler errors,
// ensuring their HTTP status reflects the response sent to the client.
//
// The middleware records only low-cardinality protocol, method, route, and
// response-status attributes. It does not record request URLs, headers, bodies,
// or handler error text.
//
// Applications own OpenTelemetry SDK configuration, exporters, sampling, and
// shutdown. Register OpenTelemetry before application middleware so it observes
// the complete request lifecycle, including errors rendered by Cosmos.
//
// Example:
//
//	provider := sdktrace.NewTracerProvider(sdktrace.WithBatcher(exporter))
//	defer provider.Shutdown(context.Background())
//
//	app.Use(
//		middleware.Recover(),
//		middleware.OpenTelemetry(middleware.OpenTelemetryConfig{
//			TracerProvider: provider,
//		}),
//		middleware.Correlation(middleware.DefaultCorrelationConfig()),
//	)
func OpenTelemetry(config OpenTelemetryConfig) framework.Middleware {
	config = config.withDefaults()
	tracer := config.TracerProvider.Tracer(otelInstrumentationName)

	return func(next framework.Handler) framework.Handler {
		return func(w http.ResponseWriter, r *http.Request) error {
			ctx := config.Propagator.Extract(r.Context(), propagation.HeaderCarrier(r.Header))
			route := otelRoutePattern(r)
			ctx, span := tracer.Start(ctx, otelSpanName(r.Method, route), trace.WithSpanKind(trace.SpanKindServer))
			span.SetAttributes(otelRequestAttributes(r, route)...)

			hooks := request.Hooks(r)
			status := http.StatusOK
			hooks.BeforeWriteHeader(func(_ http.ResponseWriter, value int) {
				status = value
			})
			hooks.AfterResponse(func(_ error) {
				span.SetAttributes(semconv.HTTPResponseStatusCode(status))

				if status >= http.StatusInternalServerError {
					span.SetStatus(codes.Error, fmt.Sprintf("HTTP %d", status))
				}

				span.End()
			})

			return next(w, r.WithContext(ctx))
		}
	}
}

func (config OpenTelemetryConfig) withDefaults() OpenTelemetryConfig {
	if config.TracerProvider == nil {
		config.TracerProvider = otel.GetTracerProvider()
	}

	if config.Propagator == nil {
		config.Propagator = propagation.TraceContext{}
	}

	return config
}

func otelRoutePattern(r *http.Request) string {
	pattern := r.Pattern

	if _, route, ok := strings.Cut(pattern, " "); ok {
		return route
	}

	return pattern
}

func otelSpanName(method, route string) string {
	if route == "" {
		return method
	}

	return method + " " + route
}

func otelRequestAttributes(r *http.Request, route string) []attribute.KeyValue {
	attributes := []attribute.KeyValue{
		semconv.HTTPRequestMethodKey.String(r.Method),
		semconv.URLScheme(otelRequestScheme(r)),
		semconv.NetworkProtocolName("http"),
		semconv.NetworkProtocolVersion(otelProtocolVersion(r.Proto)),
	}

	if route != "" {
		attributes = append(attributes, semconv.HTTPRoute(route))
	}

	return attributes
}

func otelRequestScheme(r *http.Request) string {
	if r.TLS != nil {
		return "https"
	}

	return "http"
}

func otelProtocolVersion(protocol string) string {
	if version, ok := strings.CutPrefix(protocol, "HTTP/"); ok {
		return version
	}

	return protocol
}
