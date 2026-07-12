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

Cosmos is a modular HTTP framework for Go.
Dependency direction:

`problem (zero deps) -> contract (depends on problem) -> router (standalone) -> framework (uses all)`

```bash
go get github.com/studiolambda/cosmos/framework
```

## Quick Start

```go
app := framework.New()
app.Use(middleware.Recover())
app.Use(middleware.Logger(slog.Default()))

app.Get("/users/{id}", func(w http.ResponseWriter, r *http.Request) error {
	id, err := request.ParamInt(r, "id")
	if err != nil {
		return err
	}

	user, err := findUser(id)
	if err != nil {
		return problem.Problem{
			Title:  "User not found",
			Status: http.StatusNotFound,
		}.WithError(err).With("user_id", id)
	}

	return response.JSON(w, http.StatusOK, user)
})

server := framework.NewServer(":8080", app)
_ = server.ListenAndServe()
```

## Core Types

```go
type Handler    = func(w http.ResponseWriter, r *http.Request) error
type Router     = router.Router[Handler]
type Middleware = router.Middleware[Handler] // func(Handler) Handler
```

`framework.New()` returns `*framework.Router`.

## Error Handling Pipeline

When a `framework.Handler` returns an error:

1. `context.Canceled` / `context.DeadlineExceeded` -> status `499`.
2. If error implements `framework.HTTPStatus` -> uses `HTTPStatus()`.
3. If error implements `http.Handler` -> error renders itself.
4. Otherwise -> `problem.NewProblem(err, status).ServeHTTP(...)`.

If handler returns `nil` and wrote nothing -> **204 No Content**.

## Router

```go
r := router.New[framework.Handler]()

r.Get("/users/{id}", handler)
r.Post("/users", handler)
r.Any("/health", handler)           // GET, HEAD, POST, PUT, PATCH, DELETE, OPTIONS
r.Method("CUSTOM", "/rpc", handler)

r.Use(loggingMiddleware)             // mutates current router
admin := r.With(authMiddleware)      // returns NEW sub-router

r.Group("/api/v1", func(api *router.Router[framework.Handler]) {
	api.Get("/users", listUsers)
})
```

Patterns use `http.ServeMux` style (`{id}`, `{path...}`), with automatic
trailing-slash pair registration.

For the full router API, load [references/router.md](references/router.md).

## Problem Details (RFC 9457)

```go
var ErrNotFound = problem.Problem{
	Type:   "https://api.example.com/errors/not-found",
	Title:  "Resource Not Found",
	Status: http.StatusNotFound,
}

return ErrNotFound.WithError(err).With("user_id", id)
```

`problem.Problem` implements `error`, `http.Handler`, and `json.Marshaler`.
Negotiation supports `application/problem+json`, `application/json`, then
plain text fallback.

Consumer guidance: define `problem.Problem` as package-level reusable error
variables (`var ErrUserNotFound = problem.Problem{...}`) and derive request-
specific values using `WithError` / `With`.

For full details, load [references/problem.md](references/problem.md).

## Contract Layer (Important)

`contract` exposes:

- low-level backend interfaces: `CacheDriver`, `DatabaseDriver`, `EventDriver`, `SessionDriver`
- typed wrappers: `*contract.Cache`, `*contract.Database`, `*contract.Events`
- shared abstractions: `contract.Encrypter`, `contract.Hasher`, `contract.Hooks`, `*contract.Session`

For signatures and usage, load [references/contract.md](references/contract.md).

## Request Helpers

```go
id, err := request.ParamInt(r, "id")
limit := request.QueryIntOr(r, "limit", 20)

body, err := request.StrictLimitedJSON[CreateUserInput](r, -1) // -1 => default max

sess, ok := request.Session(r)
if ok {
	sess.Put("user_id", 123)
	_ = sess.Regenerate() // after auth changes
}

if hooks, ok := request.TryHooks(r); ok {
	hooks.AfterResponse(func(err error) { /* cleanup */ })
}
```

## Response Helpers

```go
return response.Status(w, http.StatusNoContent)
return response.JSON(w, http.StatusOK, user)
return response.HTML(w, http.StatusOK, "<h1>hi</h1>")
return response.SafeRedirect(w, http.StatusFound, "/dashboard")
return response.SSE(w, r, eventChan)
```

## Built-in Middleware

```go
middleware.Recover()
middleware.Logger(slog.Default())
middleware.CSRF("https://example.com")
middleware.CORS(middleware.CORSOptions{})
middleware.SecureHeaders()
middleware.RateLimit()
middleware.Provide("db", db)
middleware.HTTP(stdlibMiddleware)
```

Recommended order: `Recover` then `Logger` near the start, and session
middleware before any code that reads sessions.

## Subpackages

### Correlation

```go
app.Use(correlation.Middleware())
logger := slog.New(correlation.Handler(baseHandler))
id := correlation.From(r)
```

### Sessions

```go
driver := session.NewCacheDriver(contract.NewCache(cache.NewMemory(5*time.Minute, 10*time.Minute)))
app.Use(session.Middleware(driver))
```

### Cache

```go
memDriver := cache.NewMemory(5*time.Minute, 10*time.Minute)
redisDriver := cache.NewRedis(&cache.RedisOptions{Addr: "localhost:6379"})

c := contract.NewCache(memDriver)
value, err := c.Remember(ctx, "key", time.Minute, computeFn)
_ = redisDriver
_ = value
```

### Crypto

```go
aes, err := crypto.NewAES(key)      // 16, 24, or 32 bytes
cc, err := crypto.NewChaCha20(key)  // exactly 32 bytes
ciphertext, err := aes.Encrypt(plaintext)
_ = cc
_ = ciphertext
```

### Hash

```go
argon := hash.NewArgon2()  // recommended
bcrypt := hash.NewBcrypt() // compatibility
hashed, err := argon.Hash(password)
ok, err := argon.Check(password, hashed)
_ = bcrypt
_ = ok
```

### Database

```go
driver, err := database.NewSQL("postgres", connString)
db := contract.NewDatabase(driver)

user, err := db.Find[User](ctx, "SELECT * FROM users WHERE id = $1", id)

err = db.WithTransaction(ctx, func(tx *contract.Database) error {
	_, err := tx.Exec(ctx, "UPDATE users SET name = $1 WHERE id = $2", "Alice", id)
	return err
})

_ = user
```

### Events

```go
driver := event.NewMemoryBroker()
ev := contract.NewEvents(driver)

unsubscribe, err := ev.Subscribe[UserCreated](ctx, "user.created", func(decode contract.EventDecoder[UserCreated]) {
	msg, err := decode()
	if err != nil {
		return
	}
	_ = msg
})
if err != nil {
	return
}
defer unsubscribe()

_ = ev.Publish(ctx, "user.created", UserCreated{ID: 42})
```

For complete framework subpackage coverage, load
[references/framework.md](references/framework.md).

## Gotchas

- Framework handlers return `error`; stdlib handlers do not.
- `nil` return with no writes -> `204 No Content`.
- `MustSession` and `request.Hooks` panic when middleware/context is missing.
- Body parsing consumes request body once.
- `problem` methods are copy-on-write (immutable style).
- `router.Use()` mutates; `router.With()` returns new sub-router.
- `router.Any()` intentionally excludes TRACE and CONNECT.
- Database transactions cannot be nested.
- `problem.Defaulted()` does **not** auto-fill `Detail` from `err.Error()`.
- Prefer `framework.NewServer()` over raw `http.ListenAndServe`.

## When to Load References

- Contract wrappers/interfaces + request/response helpers:
  [references/contract.md](references/contract.md)
- Router API and matching behavior:
  [references/router.md](references/router.md)
- Problem details API and serving rules:
  [references/problem.md](references/problem.md)
- Framework middleware, hooks, and subpackages:
  [references/framework.md](references/framework.md)
