# AGENTS.md

## Module Overview

Router module: generic HTTP router built on http.ServeMux. Middleware support, route groups, automatic trailing slash handling. Zero dependencies.

Module: github.com/studiolambda/cosmos/router
Dependencies: Zero

## Setup Commands

```bash
go test ./...
go test -cover ./...
go fmt ./...
```

## Architecture

Generic: `Router[H http.Handler]` works with any http.Handler type.
Middleware: `type Middleware[H http.Handler] = func(H) H`
Execution order: A → B → C → handler → C → B → A

Dual routes: "/users" registers both "/users" and "/users/"
Root "/" uses exact match only.

Hierarchy: child routers inherit parent middleware, all share same http.ServeMux.

## Code Style

Explicit type parameter: `router.New[http.Handler]()`
Use method helpers (Get, Post) over generic Method().
Middleware order matters: most general first.

## Common Patterns

Basic:
```go
r := router.New[http.Handler]()
r.Use(logger, recovery)
r.Get("/users/{id}", handler)
```

Groups:
```go
r.Group("/api", func(api *router.Router[http.Handler]) {
    api.Use(auth)
    api.Get("/users", listUsers)
})
```

Conditional middleware:
```go
r.With(auth).Get("/profile", handler)
```

Multiple methods:
```go
r.Methods([]string{"GET", "POST"}, "/path", handler)
r.Any("/webhook", handler)
```

Testing:
```go
req := httptest.NewRequest("GET", "/test", nil)
res := r.Record(req)
assert.Equal(t, http.StatusOK, res.StatusCode)
```

## Testing

Test route registration with Has() and Matches(). Test middleware execution. Use Record() helper for response testing.

## Package Structure

Single file: router.go with comprehensive functionality.

## Common Gotchas

- Must specify type parameter: `router.New[http.Handler]()`
- Middleware order affects execution
- Root "/" is exact match only
- Path parameters via r.PathValue("id"), not query string
- Use() modifies router, With() creates new router
- All routers share same ServeMux
- Pattern joining uses path.Join() which cleans paths
