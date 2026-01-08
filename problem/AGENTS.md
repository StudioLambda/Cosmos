# AGENTS.md

## Module Overview

Problem module: RFC 9457 (Problem Details for HTTP APIs) implementation. Zero dependencies. Structured error responses with content negotiation, stack traces, error wrapping.

Module: github.com/studiolambda/cosmos/problem
Dependencies: Zero

## Setup Commands

```bash
go test ./...
go test -cover ./...
go fmt ./...
```

## Architecture

RFC 9457 structure: Type, Title, Detail, Status, Instance. Additional metadata via map[string]any. Wrapped error (not serialized). Implements http.Handler and error interface.

Content negotiation: application/problem+json, application/json, text/plain.

## Code Style

Define problems as package variables with consistent Type URIs. Methods return new instances (immutable). Use With() for metadata, WithError() for wrapping, WithStackTrace() for dev.

## Common Patterns

Define:
```go
var ErrNotFound = problem.Problem{
    Type:   "https://api.example.com/errors/not-found",
    Title:  "Resource Not Found",
    Status: http.StatusNotFound,
}
```

Serve:
```go
ErrNotFound.With("user_id", id).ServeHTTP(w, r)
```

From error:
```go
problem.NewProblem(err, http.StatusInternalServerError).ServeHTTP(w, r)
```

Stack traces:
```go
problem.WithStackTrace().ServeHTTP(w, r)
// or
problem.ServeHTTPDev(w, r)
```

Remove data:
```go
problem.Without("debug_info").WithoutError().ServeHTTP(w, r)
```

## Testing

Test with httptest and different Accept headers. Verify status codes, content types, and JSON structure.

```go
req := httptest.NewRequest("GET", "/", nil)
req.Header.Set("Accept", "application/json")
rec := httptest.NewRecorder()
problem.ServeHTTP(rec, req)
assert.Equal(t, http.StatusNotFound, rec.Code)
```

## Package Structure

```
problem/
├── problem.go     # Main implementation
├── utils.go       # Helpers
└── internal/
    └── accept.go  # Content negotiation
```

## Common Gotchas

- Immutable: methods return new instances, use returned value
- Zero dependencies: no external imports
- Stack traces via "stack_trace" key
- Defaulting happens at serve time, not creation
- Wrapped errors not in JSON unless WithStackTrace()
- Errors with HTTPStatus() int preserve custom status
