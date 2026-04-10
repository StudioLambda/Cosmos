---
name: cosmos-problem
description: >
  RFC 9457 Problem Details implementation for Go
  (github.com/studiolambda/cosmos/problem). Use when creating structured
  HTTP error responses, working with problem details, content negotiation,
  error wrapping, or returning errors from Cosmos framework handlers. The
  Problem type implements error, http.Handler, and json.Marshaler
  simultaneously.
---

# Cosmos Problem

RFC 9457 (Problem Details for HTTP APIs) implementation. Zero external
dependencies. A `Problem` is simultaneously an `error`, an `http.Handler`,
and a `json.Marshaler`.

```
go get github.com/studiolambda/cosmos/problem
```

## Quick Start

```go
// Define reusable problem templates as package-level vars
var ErrNotFound = problem.Problem{
    Type:   "https://api.example.com/errors/not-found",
    Title:  "Resource Not Found",
    Status: http.StatusNotFound,
}

var ErrForbidden = problem.Problem{
    Type:   "https://api.example.com/errors/forbidden",
    Title:  "Forbidden",
    Status: http.StatusForbidden,
}

// In a handler — derive with additional context
func getUser(w http.ResponseWriter, r *http.Request) error {
    user, err := db.FindUser(id)
    if err != nil {
        return ErrNotFound.WithError(err).With("user_id", id)
    }

    return response.JSON(w, http.StatusOK, user)
}
```

When returned from a Cosmos framework handler, the framework detects
that `Problem` implements `http.Handler` and calls `ServeHTTP` to render
the response with content negotiation.

## The Problem Type

```go
type Problem struct {
    Type     string // URI identifying the problem type
    Title    string // Short human-readable summary
    Detail   string // Specific explanation for this occurrence
    Status   int    // HTTP status code
    Instance string // URI identifying this occurrence
}
```

All five standard RFC 9457 fields are exported. Additional fields are
stored internally and accessed via `With`/`Additional`.

## Immutable API (Copy-on-Write)

Every method returns a **new** `Problem` — the original is never modified.
This makes package-level vars safe for concurrent derivation.

```go
// Additional fields
enriched := prob.With("user_id", 42).With("trace_id", "abc")
value, ok := enriched.Additional("user_id")
stripped := enriched.Without("user_id")

// Error wrapping
wrapped := prob.WithError(err)
clean := wrapped.WithoutError()

// Stack traces (for development)
debug := prob.WithStackTrace()
prod := debug.WithoutStackTrace()
```

**Symmetric pairs:** `With`/`Without`, `WithError`/`WithoutError`,
`WithStackTrace`/`WithoutStackTrace`.

## Creating Problems

```go
// From struct literal (recommended for reusable templates)
var ErrConflict = problem.Problem{
    Type:   "https://api.example.com/errors/conflict",
    Title:  "Conflict",
    Status: http.StatusConflict,
}

// From an error + status code
prob := problem.NewProblem(err, http.StatusInternalServerError)
```

## Serving Problems

### As HTTP Response (Direct)

```go
// Content-negotiated response (JSON, Problem+JSON, or plain text)
ErrNotFound.With("id", id).ServeHTTP(w, r)

// Development mode — includes error stack trace in response
ErrNotFound.WithError(err).ServeHTTPDev(w, r)
```

### As Framework Handler Return

```go
func handler(w http.ResponseWriter, r *http.Request) error {
    // The framework calls ServeHTTP on the returned Problem
    return ErrNotFound.With("resource", "user")
}
```

The framework's error pipeline detects:
1. `http.Handler` interface → calls `ServeHTTP` (Problem renders itself).
2. `HTTPStatus` interface → uses the status code.
3. Otherwise → wraps in `NewProblem(err, status)`.

## Content Negotiation

`ServeHTTP` inspects the `Accept` header and responds with the best match:

| Accept | Content-Type | Format |
|---|---|---|
| `application/problem+json` | `application/problem+json` | RFC 9457 JSON |
| `application/json` | `application/json` | Standard JSON |
| `text/plain` or anything else | `text/plain` | `Error()` string |

## JSON Format

Standard fields + additional fields are merged into a flat JSON object:

```json
{
    "type": "https://api.example.com/errors/not-found",
    "title": "Resource Not Found",
    "detail": "user with id 42 not found",
    "status": 404,
    "instance": "/users/42",
    "user_id": 42,
    "trace_id": "abc-123"
}
```

Zero-value standard fields are omitted. The internal `err` field is
**never** serialized — use `WithStackTrace()` to include error details
under the `stack_trace` key.

## Defaulting

`Defaulted(r)` fills zero-value fields with sensible defaults. It is
called automatically inside `ServeHTTP`:

| Field | Default |
|---|---|
| `Type` | `"about:blank"` |
| `Status` | `500` |
| `Title` | `http.StatusText(status)` |
| `Instance` | Request URL |
| `Detail` | First error in the chain |

This means minimal definitions work:

```go
var ErrNotFound = problem.Problem{Status: 404}
// Served as: type=about:blank, title=Not Found, status=404, instance=<url>
```

## Error Chain Integration

`Problem` participates in Go's error chain:

```go
prob := ErrNotFound.WithError(fmt.Errorf("db: %w", sql.ErrNoRows))

errors.Is(prob, sql.ErrNoRows)   // true
errors.Unwrap(prob)              // the wrapped error
prob.Errors()                    // []error — flattened stack trace
```

`Errors()` recursively unwraps both `Unwrap() error` and
`Unwrap() []error` (from `errors.Join`) into a flat slice.

## Gotchas

- **Immutable:** All `With`/`Without` methods return new instances.
  The original is unchanged. Always capture the return value.
- **Defaulting is lazy:** `Defaulted()` runs inside `ServeHTTP`, not
  at construction. A `Problem{Status: 0}` becomes 500 only when served.
- **Wrapped errors are not in JSON:** The `err` field is never
  serialized. Use `WithStackTrace()` or `ServeHTTPDev()` to expose them.
- **`UnmarshalJSON` requires pointer receiver:** This is the only
  pointer-receiver method. All others use value receivers.
- **JSON encoding errors are silently discarded** in `ServeHTTP`:
  headers are already written at that point.
- **`StackTraceKey`** is the constant `"stack_trace"` — the key used
  by `WithStackTrace()` in the additional data map.
