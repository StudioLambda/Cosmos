# Cosmos: Router

A lightweight, generic HTTP router built on top of Go's standard `http.ServeMux`. Provides structured routing with middleware support, route groups, and automatic trailing slash handling while maintaining full compatibility with the standard library.

## Overview

The router module provides:

- **Generic Design**: Works with any `http.Handler` implementation
- **Standard Library Foundation**: Built on `http.ServeMux` for reliability
- **Middleware Support**: Composable middleware with inheritance
- **Route Groups**: Hierarchical routing with scoped middleware
- **Automatic Dual Routes**: Handles trailing slashes automatically
- **Path Parameters**: Extract parameters from URL patterns
- **Testing Helpers**: Built-in response recording for tests
- **Zero Magic**: Transparent, predictable behavior

## Installation

```bash
go get github.com/studiolambda/cosmos/router
```

This module has zero dependencies.

## Quick Start

```go
package main

import (
    "fmt"
    "log/slog"
    "net/http"
    "github.com/studiolambda/cosmos/router"
)

func logger(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        slog.Info("request", "method", r.Method, "path", r.URL.Path)
        next.ServeHTTP(w, r)
    })
}

func main() {
    r := router.New[http.Handler]()

    // Global middleware
    r.Use(logger)

    // Basic routes
    r.Get("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprintln(w, "Hello, World!")
    }))

    r.Get("/users/{id}", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        id := r.PathValue("id")
        fmt.Fprintf(w, "User ID: %s", id)
    }))

    // Route groups
    r.Group("/api", func(api *router.Router[http.Handler]) {
        api.Get("/users", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            fmt.Fprintln(w, "List users")
        }))

        api.Post("/users", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            fmt.Fprintln(w, "Create user")
        }))
    })

    http.ListenAndServe(":8080", r)
}
```

## Core Concepts

### Generic Design

The router is generic over any `http.Handler` implementation:

```go
// Standard http.Handler
r := router.New[http.Handler]()

// Custom handler type
type CustomHandler func(http.ResponseWriter, *http.Request) error
r := router.New[CustomHandler]()
```

This allows frameworks to use their own handler types while benefiting from the router's features.

### Middleware

Middleware is a function that wraps handlers:

```go
type Middleware[H http.Handler] = func(H) H

func logger(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        log.Printf("%s %s", r.Method, r.URL.Path)
        next.ServeHTTP(w, r)
    })
}

r.Use(logger)
```

Middleware executes in order: `A → B → C → handler → C → B → A`

### Trailing Slash Handling

The router automatically registers dual routes for trailing slashes:

```go
r.Get("/users", handler)

// Both routes work:
// GET /users
// GET /users/
```

Root route (`/`) uses exact match:
```go
r.Get("/", handler) // Only matches GET /
```

## HTTP Methods

All standard HTTP methods are supported:

```go
r.Get("/path", handler)       // GET
r.Post("/path", handler)      // POST
r.Put("/path", handler)       // PUT
r.Patch("/path", handler)     // PATCH
r.Delete("/path", handler)    // DELETE
r.Head("/path", handler)      // HEAD
r.Options("/path", handler)   // OPTIONS
r.Connect("/path", handler)   // CONNECT
r.Trace("/path", handler)     // TRACE
```

Multiple methods for the same route:

```go
r.Methods([]string{"GET", "POST"}, "/path", handler)
```

All methods:

```go
r.Any("/path", handler) // Registers all HTTP methods
```

Dynamic method:

```go
r.Method("GET", "/path", handler)
```

## Path Parameters

Path parameters use `http.ServeMux` syntax:

```go
// Single parameter
r.Get("/users/{id}", handler)

// Multiple parameters
r.Get("/posts/{postID}/comments/{commentID}", handler)

// Wildcard (rest of path)
r.Get("/static/{path...}", handler)
```

Extract parameters:

```go
func handler(w http.ResponseWriter, r *http.Request) {
    id := r.PathValue("id")
    fmt.Fprintf(w, "ID: %s", id)
}
```

## Route Groups

Groups provide hierarchical routing with middleware inheritance:

```go
r.Group("/api", func(api *router.Router[http.Handler]) {
    // Middleware applies to all routes in this group
    api.Use(authMiddleware)

    api.Get("/users", listUsers)
    api.Post("/users", createUser)

    // Nested groups
    api.Group("/admin", func(admin *router.Router[http.Handler]) {
        admin.Use(adminMiddleware)
        admin.Get("/settings", getSettings)
    })
})
```

Clone router for conditional middleware:

```go
r.Grouped(func(clone *router.Router[http.Handler]) {
    clone.Use(specificMiddleware)
    clone.Get("/special", handler)
})
```

## Conditional Middleware

Apply middleware to specific routes only:

```go
// With() creates a new router with additional middleware
authenticated := r.With(authMiddleware)
authenticated.Get("/profile", profileHandler)

// Original router unchanged
r.Get("/public", publicHandler) // No auth middleware
```

Chain multiple middleware:

```go
r.With(middleware1, middleware2).Get("/path", handler)
```

## Middleware Composition

### Global Middleware

Applied to all routes:

```go
r.Use(logger, recovery, cors)
r.Get("/path", handler) // All three middleware applied
```

### Group Middleware

Applied to group routes:

```go
r.Group("/api", func(api *router.Router[http.Handler]) {
    api.Use(rateLimiter) // Only for /api/* routes
    api.Get("/users", handler)
})
```

### Route Middleware

Applied to single route:

```go
r.With(authMiddleware).Get("/admin", handler)
```

### Inheritance

Child routers inherit parent middleware:

```go
r.Use(logger) // Global

r.Group("/api", func(api *router.Router[http.Handler]) {
    api.Use(auth) // api routes have: logger + auth

    api.Group("/admin", func(admin *router.Router[http.Handler]) {
        admin.Use(adminCheck) // admin routes have: logger + auth + adminCheck
    })
})
```

## Route Checking

### Check if Route Exists

```go
if r.Has("GET", "/users/123") {
    // Route exists
}
```

### Check if Request Matches

```go
req := httptest.NewRequest("GET", "/users/123", nil)
if r.Matches(req) {
    // Request matches a registered route
}
```

### Get Handler for Route

```go
handler, ok := r.Handler("GET", "/users/123")
if ok {
    // Handler found
}
```

### Get Handler for Request

```go
req := httptest.NewRequest("GET", "/users/123", nil)
handler, ok := r.HandlerMatch(req)
if ok {
    // Handler found
}
```

## Testing Helpers

### Record Response

Test route responses easily:

```go
r := router.New[http.Handler]()
r.Get("/hello", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("Hello"))
}))

req := httptest.NewRequest("GET", "/hello", nil)
res := r.Record(req)

fmt.Println(res.StatusCode) // 200
fmt.Println(res.Body.String()) // "Hello"
```

### Test Middleware

```go
func TestMiddleware(t *testing.T) {
    called := false
    middleware := func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            called = true
            next.ServeHTTP(w, r)
        })
    }

    r := router.New[http.Handler]()
    r.Use(middleware)
    r.Get("/test", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
    }))

    req := httptest.NewRequest("GET", "/test", nil)
    res := r.Record(req)

    assert.True(t, called)
    assert.Equal(t, http.StatusOK, res.StatusCode)
}
```

## Advanced Usage

### Custom Handler Types

Use with frameworks that have custom handler signatures:

```go
type ErrorHandler func(http.ResponseWriter, *http.Request) error

r := router.New[ErrorHandler]()

handler := func(w http.ResponseWriter, r *http.Request) error {
    if err := validate(r); err != nil {
        return err
    }
    w.Write([]byte("OK"))
    return nil
}

r.Get("/users", handler)
```

### Serving Static Files

```go
r.Get("/static/{path...}", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    path := r.PathValue("path")
    http.ServeFile(w, r, filepath.Join("static", path))
}))
```

### Subdomain Routing

Use middleware to handle subdomains:

```go
func subdomain(subdomain string, handler http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if strings.HasPrefix(r.Host, subdomain+".") {
            handler.ServeHTTP(w, r)
            return
        }
        http.NotFound(w, r)
    })
}

r.Get("/api/users", subdomain("api", usersHandler))
```

### Method Not Allowed

Handle unsupported methods:

```go
r.Get("/users", getUsers)

r.Post("/users", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}))
```

## Complete Example

```go
package main

import (
    "encoding/json"
    "log"
    "log/slog"
    "net/http"
    "github.com/studiolambda/cosmos/router"
)

type User struct {
    ID   string `json:"id"`
    Name string `json:"name"`
}

func logger(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        slog.Info("request", "method", r.Method, "path", r.URL.Path)
        next.ServeHTTP(w, r)
    })
}

func requireAuth(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        token := r.Header.Get("Authorization")
        if token == "" {
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }
        next.ServeHTTP(w, r)
    })
}

func main() {
    r := router.New[http.Handler]()

    // Global middleware
    r.Use(logger)

    // Public routes
    r.Get("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("Welcome"))
    }))

    // API routes with authentication
    r.Group("/api", func(api *router.Router[http.Handler]) {
        api.Use(requireAuth)

        api.Get("/users", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            users := []User{{ID: "1", Name: "Alice"}}
            json.NewEncoder(w).Encode(users)
        }))

        api.Get("/users/{id}", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            id := r.PathValue("id")
            user := User{ID: id, Name: "Alice"}
            json.NewEncoder(w).Encode(user)
        }))

        api.Post("/users", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            var user User
            json.NewDecoder(r.Body).Decode(&user)
            w.WriteHeader(http.StatusCreated)
            json.NewEncoder(w).Encode(user)
        }))
    })

    log.Fatal(http.ListenAndServe(":8080", r))
}
```

## Testing

```bash
# Run tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific test
go test -run TestRouterName
```

## Performance

The router is built on `http.ServeMux`, which uses efficient pattern matching. Performance characteristics:

- Route registration: O(1) per route
- Route matching: O(log n) where n is number of routes
- Middleware: O(m) where m is number of middleware layers
- Memory: Minimal overhead per route

## License

MIT License - Copyright (c) 2025 Erik C. Fores
