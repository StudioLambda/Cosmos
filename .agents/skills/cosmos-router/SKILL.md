---
name: cosmos-router
description: >
  Generic HTTP router for Go built on http.ServeMux
  (github.com/studiolambda/cosmos/router). Use when building HTTP
  services with the Cosmos router: registering routes, applying
  middleware, grouping routes, testing with Record, or using the router
  with custom handler types via generics.
---

# Cosmos Router

A generic HTTP router built on `http.ServeMux` with zero external
dependencies. The type parameter makes it usable with any handler type
that satisfies `http.Handler`.

```
go get github.com/studiolambda/cosmos/router
```

## Quick Start

```go
// With standard http.HandlerFunc
r := router.New[http.HandlerFunc]()

r.Get("/hello", func(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("hello"))
})

http.ListenAndServe(":8080", r)
```

```go
// With the Cosmos framework's error-returning Handler
r := router.New[framework.Handler]()

r.Get("/users/{id}", func(w http.ResponseWriter, r *http.Request) error {
    id := request.Param(r, "id")
    return response.JSON(w, http.StatusOK, user)
})
```

## API

### Constructor

```go
func New[H http.Handler]() *Router[H]
```

Always specify the type parameter. The router implements `http.Handler`
and can be used directly with `http.ListenAndServe`.

### Route Registration

```go
r.Get(pattern, handler)
r.Post(pattern, handler)
r.Put(pattern, handler)
r.Patch(pattern, handler)
r.Delete(pattern, handler)
r.Head(pattern, handler)
r.Options(pattern, handler)
r.Connect(pattern, handler)
r.Trace(pattern, handler)
r.Any(pattern, handler)              // all 9 methods
r.Method(method, pattern, handler)   // explicit method string
r.Methods(methods, pattern, handler) // multiple methods
```

Patterns follow `http.ServeMux` syntax: `"/users/{id}"`, `"/files/{path...}"`.

Every route automatically registers both with and without trailing slash.
`/users` also matches `/users/`. The root path `/` only registers exact
match.

### Middleware

```go
type Middleware[H http.Handler] = func(H) H
```

The type alias preserves direct assignability — no type conversion needed
when passing function literals.

```go
// Mutates the router (appends middleware in-place)
r.Use(loggingMiddleware, recoveryMiddleware)

// Returns a NEW sub-router with additional middleware (immutable)
authed := r.With(authMiddleware)
authed.Get("/profile", profileHandler)
```

**Execution order:** First registered = outermost layer.
`A -> B -> C -> handler -> C -> B -> A` (onion model).

**Critical distinction:** `Use()` mutates, `With()` creates new.

### Grouping

```go
// Prefix group via callback
r.Group("/api/v1", func(api *Router[H]) {
    api.Get("/users", listUsers)    // matches /api/v1/users
    api.Post("/users", createUser)  // matches /api/v1/users
})

// Same-prefix group (middleware scoping without path change)
r.Grouped(func(sub *Router[H]) {
    sub.Use(adminOnly)
    sub.Delete("/users/{id}", deleteUser)
})

// Manual clone
sub := r.Clone()
```

Sub-routers inherit middleware (via `slices.Clone`) and share the same
underlying `http.ServeMux`. Mutations to sub-router middleware don't
affect the parent.

### Route Inspection

```go
r.Has("GET", "/users/{id}")  // bool — does this route exist?
r.Matches(req)               // bool — does any route match this request?
r.Handler("GET", "/users")   // (H, bool) — get the handler
r.HandlerMatch(req)          // (H, bool) — get handler for request
```

### Testing

```go
req := httptest.NewRequest(http.MethodGet, "/users/42", nil)
res := r.Record(req) // *http.Response

require.Equal(t, http.StatusOK, res.StatusCode)
```

`Record` executes the request through the full router pipeline
(middleware + handler) and returns the response. Uses `httptest`
internally.

## Patterns & Conventions

### Path Parameters

Use `http.ServeMux` syntax. Extract with `request.Param` from the
contract module:

```go
r.Get("/users/{id}", func(w http.ResponseWriter, r *http.Request) error {
    id := request.Param(r, "id")
    // ...
})

// Wildcard (catch-all)
r.Get("/files/{path...}", serveFile)
```

### Middleware Composition

```go
func AuthMiddleware(db Database) router.Middleware[framework.Handler] {
    return func(next framework.Handler) framework.Handler {
        return func(w http.ResponseWriter, r *http.Request) error {
            token := request.Header(r, "Authorization")
            if token == "" {
                return response.Status(w, http.StatusUnauthorized)
            }

            return next(w, r)
        }
    }
}

r.Use(AuthMiddleware(db))
```

### Adapting stdlib Middleware

When using the framework, use `middleware.HTTP` to adapt standard
`func(http.Handler) http.Handler` middleware:

```go
r.Use(middleware.HTTP(corsMiddleware))
```

## Gotchas

- **Always specify the type parameter:** `router.New[http.Handler]()`
  — the compiler cannot infer it.
- **`Use()` mutates, `With()` creates new:** Calling `Use()` on a
  sub-router from `Group()` is safe (scoped), but calling it on the
  root router after routes are registered adds middleware to all
  subsequently registered routes only.
- **Trailing slashes are automatic:** Don't register both `/users` and
  `/users/` — the router handles this.
- **`path.Join` normalizes patterns:** Double slashes and dots are
  cleaned. This is usually desired but affects exact pattern matching.
- **Sub-routers share one `http.ServeMux`:** All routes across the
  entire tree register on the root's mux.
- **Catch-all wildcards:** Routes ending in `{path...}` also register
  the prefix without the wildcard segment.
