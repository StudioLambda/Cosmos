# Core Patterns

Detailed patterns for the framework core: handler error pipeline,
lifecycle hooks, and response writer wrapping.

## Handler Error Pipeline

`Handler.ServeHTTP` is the bridge between the framework's error-returning
handlers and the standard `http.Handler` interface:

```go
func (handler Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    // 1. Create hooks and wrap the response writer
    hooks := NewHooks()
    writer := NewResponseWriter(w, hooks)
    r = r.WithContext(context.WithValue(r.Context(), contract.HooksKey, hooks))

    // 2. Call the handler
    err := handler(writer, r)

    // 3. Error handling pipeline
    if err != nil {
        status := http.StatusInternalServerError

        // Check for context cancellation → 499
        if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
            status = 499
        }

        // Check for HTTPStatus interface → custom status
        var httpStatus HTTPStatus
        if errors.As(err, &httpStatus) {
            status = httpStatus.HTTPStatus()
        }

        // Check for http.Handler interface → self-rendering error
        var handler http.Handler
        if errors.As(err, &handler) {
            handler.ServeHTTP(writer, r)
        } else {
            problem.NewProblem(err, status).ServeHTTP(writer, r)
        }
    }

    // 4. Default 204 if nothing was written
    if !writer.WriteHeaderCalled() {
        writer.WriteHeader(http.StatusNoContent)
    }

    // 5. Fire AfterResponse hooks
    for _, hook := range hooks.AfterResponseFuncs() {
        hook(err)
    }
}
```

Key points:
- Uses `errors.As` (not type assertions) for wrapped error support.
- `problem.Problem` implements both `HTTPStatus` and `http.Handler`,
  so it takes the `http.Handler` branch and renders itself.
- Status 499 is a non-standard code for client disconnections.
- AfterResponse hooks always fire, even on error.

## Lifecycle Hooks

Three hook points in the request lifecycle:

### BeforeWriteHeader

Fires just before the HTTP status line is sent. Use for setting
response headers that depend on the status code.

```go
hooks.BeforeWriteHeader(func(w http.ResponseWriter, status int) {
    if status >= 400 {
        w.Header().Set("X-Error", "true")
    }
})
```

### BeforeWrite

Fires before each body write. Use for response body transformation
or logging.

```go
hooks.BeforeWrite(func(w http.ResponseWriter, content []byte) {
    log.Debug("writing response body", "size", len(content))
})
```

### AfterResponse

Fires after the handler completes. Receives the error (or nil).
Use for cleanup, logging, or metrics.

```go
hooks.AfterResponse(func(err error) {
    duration := time.Since(start)
    metrics.RecordLatency(duration)
})
```

### Execution Order

Hooks execute in LIFO order (last registered fires first). The
`*Funcs()` methods return reversed clones of the registered hooks.

### Session Middleware Uses BeforeWriteHeader

Session persistence happens in a BeforeWriteHeader hook. The session
cookie and session data are saved just before the status line is sent.
If you manually call `w.WriteHeader()` before the hook fires, the
session will not be persisted.

## ResponseWriter Wrapping

`NewResponseWriter` wraps `http.ResponseWriter` to intercept writes
and fire hooks:

```go
writer := framework.NewResponseWriter(w, hooks)
```

Returns `WrappedResponseWriter` interface:

```go
type WrappedResponseWriter interface {
    http.ResponseWriter
    WriteHeaderCalled() bool
}
```

If the underlying writer implements `http.Flusher` (needed for SSE
and streaming), the wrapped writer also implements `http.Flusher`.

### Write Behavior

- `WriteHeader(status)`: fires BeforeWriteHeader hooks, then delegates.
  No-op on subsequent calls (first status wins).
- `Write(content)`: if WriteHeader hasn't been called, defaults to 200.
  Fires BeforeWrite hooks before each write.

## The Record Test Helper

Both `Handler` and `Router` provide a `Record` method for testing:

```go
handler := framework.Handler(func(w http.ResponseWriter, r *http.Request) error {
    return response.JSON(w, http.StatusOK, data)
})

req := httptest.NewRequest(http.MethodGet, "/", nil)
res := handler.Record(req) // *http.Response

require.Equal(t, http.StatusOK, res.StatusCode)
```

`Record` creates an `httptest.NewRecorder`, calls `ServeHTTP`, and
returns `recorder.Result()`. This runs through the full pipeline
including hooks, error handling, and the 204 default.
