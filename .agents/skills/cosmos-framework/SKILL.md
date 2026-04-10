---
name: cosmos-framework
description: >
  Full HTTP framework for Go built on the Cosmos modules
  (github.com/studiolambda/cosmos/framework). Use when building HTTP
  applications with error-returning handlers, middleware, sessions,
  caching, encryption, hashing, database access, or event-driven
  messaging. Covers the framework core and all subpackages: middleware,
  session, cache, crypto, hash, database, event.
---

# Cosmos Framework

Complete HTTP framework with error-returning handlers. Built on the
contract, router, and problem modules.

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

http.ListenAndServe(":8080", app)
```

## Core Types

```go
type Handler    = func(w http.ResponseWriter, r *http.Request) error
type Router     = router.Router[Handler]
type Middleware  = router.Middleware[Handler]  // = func(Handler) Handler
```

`Handler` is the fundamental type. It wraps the standard `http.Handler`
pattern with an error return for centralized error handling.

`framework.New()` returns `*Router` — use the full router API (Get, Post,
Group, Use, With, etc.) documented in the cosmos-router skill.

## Error Handling Pipeline

When `Handler.ServeHTTP` is called, errors flow through this pipeline:

1. `context.Canceled` / `context.DeadlineExceeded` → status 499.
2. Error implements `HTTPStatus` interface → custom status code.
3. Error implements `http.Handler` → error renders itself (e.g. `problem.Problem`).
4. Otherwise → wrapped in `problem.NewProblem(err, status)` and served.

If the handler returns `nil` and never writes a response → **204 No Content**.

```go
// Custom status via HTTPStatus interface
type HTTPStatus interface {
    HTTPStatus() int
}

// problem.Problem implements both http.Handler and HTTPStatus,
// so returning a Problem from a handler renders it directly.
```

## Lifecycle Hooks

Access via `request.Hooks(r)` inside any handler in the pipeline.

```go
func handler(w http.ResponseWriter, r *http.Request) error {
    hooks := request.Hooks(r)

    hooks.BeforeWriteHeader(func(w http.ResponseWriter, status int) {
        w.Header().Set("X-Request-ID", "abc")
    })

    hooks.AfterResponse(func(err error) {
        if err != nil {
            log.Error("request failed", "error", err)
        }
    })

    return response.JSON(w, http.StatusOK, data)
}
```

Hook execution order is LIFO (last registered fires first).

For detailed hook, handler, and response writer patterns, see
[references/core-patterns.md](references/core-patterns.md).

## Middleware

### Built-in Middleware

```go
middleware.Recover()                         // panic recovery → error
middleware.RecoverWith(customHandler)         // panic recovery with custom converter
middleware.Logger(slog.Default())            // logs failed requests (errors + 5xx)
middleware.CSRF("https://example.com")       // cross-origin protection
middleware.CSRFWith(csrf, customProblem)      // custom CSRF config + error
middleware.Provide("db", db)                 // inject static value into context
middleware.ProvideWith(dynamicProvider)       // inject dynamic value
middleware.HTTP(stdlibMiddleware)            // adapt stdlib middleware
```

### Writing Middleware

```go
func RateLimit(limit int) framework.Middleware {
    return func(next framework.Handler) framework.Handler {
        return func(w http.ResponseWriter, r *http.Request) error {
            if !allowed(r, limit) {
                return ErrTooManyRequests
            }

            return next(w, r)
        }
    }
}
```

**Order matters:** Register `Recover` and `Logger` first (outermost).

### Middleware Ordering

```go
app.Use(middleware.Recover())   // 1st — outermost, catches all panics
app.Use(middleware.Logger(log)) // 2nd — logs all requests
app.Use(session.Middleware(d))  // 3rd — session available in handlers
```

## Subpackages

For complete API reference of each subpackage, see
[references/subpackages.md](references/subpackages.md).

### Sessions

```go
// Setup
memCache := cache.NewMemory(5*time.Minute, 10*time.Minute)
driver := session.NewCacheDriver(memCache)
app.Use(session.Middleware(driver))

// In handlers
sess, ok := request.Session(r)
sess.Put("user_id", 123)
sess.Regenerate() // Always regenerate after authentication
```

Defaults: cookie `"cosmos.session"`, TTL 2h, secure, HttpOnly, SameSite=Lax.
Customize with `session.MiddlewareWith(driver, options)`.

### Cache

```go
// In-memory (development/testing)
mem := cache.NewMemory(5*time.Minute, 10*time.Minute)

// Redis (production)
rdb := cache.NewRedis(&cache.RedisOptions{Addr: "localhost:6379"})

// Common operations (both implement contract.Cache)
cache.Put(ctx, "key", "value", 5*time.Minute)
val, err := cache.Get(ctx, "key")
val, err := cache.Remember(ctx, "key", ttl, func() (any, error) {
    return computeExpensiveValue(), nil
})
```

### Crypto

```go
// AES-GCM (key: 16, 24, or 32 bytes)
aes, err := crypto.NewAES(key)
ciphertext, err := aes.Encrypt(plaintext)
plaintext, err := aes.Decrypt(ciphertext)

// ChaCha20-Poly1305 (key: exactly 32 bytes)
cc, err := crypto.NewChaCha20(key)
ciphertext, err := cc.Encrypt(plaintext)
plaintext, err := cc.Decrypt(ciphertext)
```

Both prepend the nonce to ciphertext. Decrypt expects this format.

### Hash

```go
// Argon2id (recommended)
hasher := hash.NewArgon2()
hashed, err := hasher.Hash(password)
ok, err := hasher.Check(password, hashed)

// Bcrypt (when compatibility needed)
hasher := hash.NewBcrypt()
```

### Database

```go
db, err := database.NewSQL("postgres", connString)
defer db.Close()

// Queries
err := db.Find(ctx, "SELECT * FROM users WHERE id = $1", &user, id)
err := db.Select(ctx, "SELECT * FROM users", &users)
rows, err := db.Exec(ctx, "DELETE FROM users WHERE id = $1", id)

// Transactions
err := db.WithTransaction(ctx, func(tx contract.Database) error {
    _, err := tx.Exec(ctx, "UPDATE accounts SET balance = balance - $1 WHERE id = $2", amount, from)
    if err != nil {
        return err // triggers rollback
    }

    _, err = tx.Exec(ctx, "UPDATE accounts SET balance = balance + $1 WHERE id = $2", amount, to)

    return err
})
```

Transactions cannot be nested — returns `ErrDatabaseNestedTransaction`.

### Events

```go
// In-memory (testing/single-instance)
broker := event.NewMemoryBroker()

// Redis Pub/Sub
broker := event.NewRedisBroker(&event.RedisBrokerOptions{Addr: "localhost:6379"})

// NATS
broker, err := event.NewNATSBroker("nats://localhost:4222")

// RabbitMQ (AMQP)
broker, err := event.NewAMQPBroker("amqp://guest:guest@localhost:5672/")

// MQTT
broker, err := event.NewMQTTBroker("mqtt://localhost:1883")
```

All implement `contract.Events`:

```go
// Subscribe with wildcards
unsub, err := broker.Subscribe(ctx, "user.*.created", func(payload contract.EventPayload) {
    var user User
    payload(&user) // unmarshal into struct
})
defer unsub()

// Publish
err := broker.Publish(ctx, "user.42.created", user)

// Cleanup
broker.Close()
```

Wildcard syntax varies by broker — see
[references/subpackages.md](references/subpackages.md) for details.

## When to Load References

- **Core handler/hook/writer patterns:** See
  [references/core-patterns.md](references/core-patterns.md)
- **Subpackage API details (sessions, cache, crypto, hash, database,
  events):** See [references/subpackages.md](references/subpackages.md)
