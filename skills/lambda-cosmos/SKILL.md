---
name: lambda-cosmos
description: >
  Complete reference for the Cosmos HTTP framework modules
  (github.com/studiolambda/cosmos). Covers all four modules: contract
  (interfaces, request/response helpers), router (generic HTTP routing),
  problem (RFC 9457 error responses), and framework (handlers, middleware,
  sessions, cache, crypto, hash, database, events). Use when building,
  consuming, or reviewing any Cosmos module API.
---

# Cosmos

Cosmos is a modular HTTP framework for Go. Four modules, one dependency
direction: contract (zero deps) -> router, problem (standalone) ->
framework (uses all).

```
go get github.com/studiolambda/cosmos/framework
```

## Quick Start

```go
app := framework.New()
app.Use(middleware.Recover())
app.Use(middleware.Logger(slog.Default()))

app.Get("/users/{id}", func(w http.ResponseWriter, r *http.Request) error {
    id := request.Param(r, "id")
    user, err := findUser(id)
    if err != nil {
        return ErrNotFound.WithError(err).With("user_id", id)
    }

    return response.JSON(w, http.StatusOK, user)
})

server := framework.NewServer(":8080", app)
server.ListenAndServe()
```

## Core Types

```go
type Handler    = func(w http.ResponseWriter, r *http.Request) error
type Router     = router.Router[Handler]
type Middleware  = router.Middleware[Handler]  // = func(Handler) Handler
```

`framework.New()` returns `*Router`. Use the full router API (Get, Post,
Group, Use, With, etc.).

## Error Handling Pipeline

When a handler returns an error:

1. `context.Canceled` / `context.DeadlineExceeded` -> status 499.
2. Error implements `HTTPStatus` interface -> custom status code.
3. Error implements `http.Handler` -> error renders itself (e.g. Problem).
4. Otherwise -> wrapped in `problem.NewProblem(err, status)` and served.

If the handler returns `nil` and never writes -> **204 No Content**.

## Router

```go
r := router.New[framework.Handler]()

r.Get("/users/{id}", handler)
r.Post("/users", handler)
r.Any("/health", handler)              // all methods
r.Method("CUSTOM", "/rpc", handler)    // explicit method

// Middleware
r.Use(loggingMiddleware)               // mutates (appends in-place)
authed := r.With(authMiddleware)       // returns NEW sub-router

// Grouping
r.Group("/api/v1", func(api *router.Router[framework.Handler]) {
    api.Get("/users", listUsers)       // matches /api/v1/users
})
```

Patterns follow `http.ServeMux` syntax: `{id}`, `{path...}`.
Trailing slashes are registered automatically.

For the full router API, load [references/router.md](references/router.md).

## Problem Details (RFC 9457)

```go
// Define reusable templates as package-level vars
var ErrNotFound = problem.Problem{
    Type:   "https://api.example.com/errors/not-found",
    Title:  "Resource Not Found",
    Status: http.StatusNotFound,
}

// Derive with context (copy-on-write, original unchanged)
return ErrNotFound.WithError(err).With("user_id", id)
```

Problem implements `error`, `http.Handler`, and `json.Marshaler`
simultaneously. Content negotiation: `application/problem+json`,
`application/json`, `text/plain`.

For the full problem API, load [references/problem.md](references/problem.md).

## Contract Interfaces

The contract module defines all service interfaces:

| Interface | Purpose |
|---|---|
| `Cache` | Key-value caching (10 methods) |
| `Database` | SQL queries and transactions |
| `Encrypter` | Symmetric encrypt/decrypt |
| `Hasher` | Password hash/check |
| `Session` | Session read/write/regenerate |
| `SessionDriver` | Session persistence |
| `Events` | Pub/sub messaging |
| `Hooks` | Request lifecycle hooks |

For full interface signatures and helpers, load
[references/contract.md](references/contract.md).

## Request Helpers

```go
// Parameters & query
id := request.Param(r, "id")
page := request.QueryOr(r, "page", "1")

// Body parsing (generics, consumes body once)
user, err := request.JSON[User](r)

// Sessions
sess, ok := request.Session(r)
sess.Put("user_id", 123)
sess.Regenerate() // after authentication

// Hooks
hooks := request.Hooks(r)
hooks.AfterResponse(func(err error) { /* cleanup */ })
```

## Response Helpers

```go
return response.Status(w, http.StatusNoContent)
return response.JSON(w, http.StatusOK, user)
return response.HTML(w, http.StatusOK, "<h1>hi</h1>")
return response.Redirect(w, http.StatusFound, "/login")
return response.SSE(w, r, eventChan)
```

## Built-in Middleware

```go
middleware.Recover()                     // panic recovery
middleware.Logger(slog.Default())        // structured logging
middleware.CSRF("https://example.com")   // cross-origin protection
middleware.CORS(middleware.CORSOptions{}) // CORS headers
middleware.SecureHeaders()               // security headers
middleware.RateLimit()                   // token bucket rate limiting
middleware.Provide("db", db)             // context injection
middleware.HTTP(stdlibMiddleware)        // adapt stdlib middleware
```

**Order matters:** Recover first, Logger second, Session third.

## Subpackages

### Correlation

```go
app.Use(correlation.Middleware())           // ensures correlation ID on every request
logger := slog.New(correlation.Handler(h))  // injects correlation_id into log records
id := correlation.From(r)                   // retrieve from request context
```

### Sessions

```go
driver := session.NewCacheDriver(memCache)
app.Use(session.Middleware(driver))
```

### Cache

```go
mem := cache.NewMemory(5*time.Minute, 10*time.Minute)
rdb := cache.NewRedis(&cache.RedisOptions{Addr: "localhost:6379"})
val, err := c.Remember(ctx, "key", ttl, computeFn)
```

### Crypto

```go
aes, err := crypto.NewAES(key)     // 16, 24, or 32 bytes
cc, err := crypto.NewChaCha20(key) // exactly 32 bytes
ciphertext, err := enc.Encrypt(plaintext)
```

### Hash

```go
hasher := hash.NewArgon2()  // recommended
hasher := hash.NewBcrypt()  // when compatibility needed
hashed, err := hasher.Hash(password)
ok, err := hasher.Check(password, hashed)
```

### Database

```go
db, err := database.NewSQL("postgres", connString)
err := db.Find(ctx, query, &user, id)
err := db.WithTransaction(ctx, func(tx contract.Database) error {
    return tx.Exec(ctx, query, args...)
})
```

### Events

```go
broker := event.NewMemoryBroker()       // testing
broker := event.NewRedisBroker(opts)    // Redis Pub/Sub
broker, err := event.NewNATSBroker(url) // NATS
unsub, err := broker.Subscribe(ctx, "user.*.created", handler)
err := broker.Publish(ctx, "user.42.created", user)
```

For complete subpackage APIs, load
[references/framework.md](references/framework.md).

## Gotchas

- Framework handlers return `error`, stdlib handlers don't.
- `nil` return with no writes -> 204 No Content.
- Middleware order matters: Recover and Logger first.
- Session middleware required before accessing sessions.
- `MustSession` panics without session middleware (prefer `Session`).
- Body parsing consumes the request body — one call per request.
- `response.JSON` appends a trailing newline.
- Problem methods return new instances (immutable copy-on-write).
- `Use()` mutates the router, `With()` creates a new sub-router.
- Database transactions cannot be nested.
- Always use `framework.NewServer()` instead of `http.ListenAndServe`.

## When to Load References

- **Contract interfaces, request/response helpers, mocks:**
  [references/contract.md](references/contract.md)
- **Router API, grouping, middleware composition, testing:**
  [references/router.md](references/router.md)
- **Problem Details creation, serving, content negotiation:**
  [references/problem.md](references/problem.md)
- **Framework core patterns, hooks, subpackage APIs:**
  [references/framework.md](references/framework.md)
