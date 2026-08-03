# AGENTS.md

## Module Overview

Problem module: RFC 9457 (Problem Details for HTTP APIs) implementation. Zero dependencies. Structured error responses with content negotiation and error wrapping.

Module: github.com/studiolambda/cosmos/problem
Dependencies: Zero

## Setup Commands

```bash
go test ./...
go test -cover ./...
go fmt ./...
```

## Architecture

Details is the RFC 9457 structure: Type, Title, Detail, Status, Instance. Additional metadata via map[string]any. Wrapped errors are not serialized. It implements http.Handler and error.

Content negotiation: application/problem+json, application/json, text/plain.

## Code Style

Define problems as package variables with consistent Type URIs. Methods return new instances (immutable). Use With() for metadata and WithError() for wrapping.

## Common Patterns

Define:

```go
var ErrNotFound = problem.Details{
    Type:   "https://api.example.com/errors/not-found",
    Title:  "Resource Not Found",
    Status: http.StatusNotFound,
}
```

Serve:

```go
ErrNotFound.With("user_id", id).ServeHTTP(w, r)
```

From error (preferred pattern):

```go
ErrInternal.WithError(err).ServeHTTP(w, r)
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
└── internal/
    └── accept.go  # Content negotiation
```

## Common Gotchas

- Immutable: methods return new instances, use returned value
- Zero dependencies: no external imports
- Defaulting happens at serve time, not creation
- Wrapped errors are never serialized
- Errors with HTTPStatus() int preserve custom status
