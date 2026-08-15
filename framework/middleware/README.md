# framework/middleware

Built-in middleware for `github.com/studiolambda/cosmos/framework`.

## Included middleware

- Panic recovery (`Recover`, `RecoverWith`)
- Structured logging (`Logger`)
- CORS and CSRF protection
- Secure response headers
- Rate limiting
- OpenTelemetry HTTP tracing (`OpenTelemetry`)
- Context value injection (`Provide`, `ProvideWith`)
- Adapter for standard `func(http.Handler) http.Handler` middleware (`HTTP`)

## Ordering guidance

Place recovery and logging near the beginning of the chain so downstream
failures are consistently captured.

`middleware.HTTP` adapts standard middleware; `framework.HTTP` adapts a
standard `http.Handler` for a Cosmos route.

## OpenTelemetry tracing

`OpenTelemetry` creates one server span for every request. It extracts a W3C
`traceparent` parent from incoming headers and ends the span after Cosmos has
rendered returned errors and written the final response.

Applications configure OpenTelemetry exporters, resources, sampling, and
provider shutdown. Cosmos does not configure or own the provider. Pass the
configured provider to `OpenTelemetry`:

```go
provider := sdktrace.NewTracerProvider(
    sdktrace.WithBatcher(exporter),
)
defer provider.Shutdown(context.Background())

app.Use(
    middleware.Recover(),
    middleware.OpenTelemetry(middleware.OpenTelemetryConfig{
        TracerProvider: provider,
    }),
     middleware.Correlation(middleware.DefaultCorrelationConfig()),
)
```

Register `Recover` first so recovered panics become normal Cosmos errors, then
register `OpenTelemetry` before application middleware. `Correlation` may be
registered after it; both use the incoming W3C trace ID.

When `TracerProvider` is nil, the middleware uses OpenTelemetry's global
provider. When `Propagator` is nil, it uses W3C Trace Context propagation:

```go
app.Use(middleware.OpenTelemetry(middleware.OpenTelemetryConfig{}))
```

Provide a propagator only when the application needs one other than W3C Trace
Context:

```go
app.Use(middleware.OpenTelemetry(middleware.OpenTelemetryConfig{
    TracerProvider: provider,
    Propagator:     propagator,
}))
```

Span names use the HTTP method and matched route pattern, such as
`GET /users/{id}`. Spans contain `http.request.method`, `url.scheme`,
`network.protocol.name`, `network.protocol.version`, `http.route` when a route
is matched, and `http.response.status_code`. Final 5xx responses set the span
status to error. URLs, query strings, headers, bodies, and handler error text
are excluded. `Correlation` uses the same W3C trace ID as its correlation ID
when one is present.
