# AGENTS.md

## Module Overview

Framework module: complete HTTP framework with error-returning handlers, middleware, sessions, caching, crypto, hashing, database. Built on router, problem, and contract modules.

Module: github.com/studiolambda/cosmos/framework
Dependencies: router v0.4.0, problem v0.4.0, contract v0.10.0, sqlx, pgx, go-sql-driver/mysql, modernc.org/sqlite, go-redis, go-cache, golang.org/x/crypto, golang.org/x/time, argon2, nats, amqp091, paho.golang (MQTT)

## Setup Commands

```bash
go test ./...
go test -cover ./...
go test ./middleware/...
go fmt ./...
```

## Architecture

Core pattern: `type Handler func(w http.ResponseWriter, r *http.Request) error`

Middleware: `type Middleware = func(next Handler) Handler`
Execution order: A → B → C → handler → C → B → A

Hooks: BeforeWriteHeader, BeforeWrite, AfterResponse

Packages: middleware/, session/, cache/, crypto/, hash/, database/

## Code Style

Handlers return errors for centralized handling. Always check errors. Use sync.Mutex for shared state with defer unlock. Table-driven tests with testify.

## Common Patterns

Application:
```go
app := framework.New()
app.Use(middleware.Logger(contract.NewLogger(frameworklogger.NewSlogFrom(slog.Default()))))
app.Use(middleware.Recover())
app.Get("/users/{id}", handler)
```

Handler:
```go
func handler(w http.ResponseWriter, r *http.Request) error {
    id := request.Param(r, "id")
    return response.JSON(w, http.StatusOK, data)
}
```

Middleware:
```go
func MyMiddleware() framework.Middleware {
    return func(next framework.Handler) framework.Handler {
        return func(w http.ResponseWriter, r *http.Request) error {
            // before
            err := next(w, r)
            // after
            return err
        }
    }
}
```

Sessions:
```go
driver := session.NewCache(cache, 24*time.Hour)
app.Use(middleware.Session(driver, middleware.DefaultSessionConfig))

sess := request.Session(r)
sess.Put("user_id", 123)
sess.Regenerate()
```

Cache:
```go
cache := memory.NewMemory(memory.MemoryConfig{Expiration: 5*time.Minute, Cleanup: 10*time.Minute})
cache.Remember(ctx, key, ttl, compute)
```

Crypto:
```go
aes := aes.NewAES(aes.AESConfig{Key: key}) // 16, 24, or 32 bytes
ciphertext, err := aes.Encrypt(ctx, plaintext)
```

Hash:
```go
hasher := argon2.NewArgon2(argon2.DefaultArgon2Config())
hashed, err := hasher.Hash(ctx, password)
err := hasher.Verify(ctx, password, hashed)
```

Database:
```go
driver, err := postgres.New(postgres.Config{DSN: connString})
db := contract.NewDatabase(driver)
err := db.Find(ctx, query, &user, id)
db.WithTransaction(ctx, func(tx contract.Database) error {
    return tx.Exec(ctx, query, args...)
})
```

## Testing

Test handlers with httptest. Test middleware by wrapping dummy handlers. Use contract mocks for services.

## Security

- CSRF: middleware.CSRF(origins...)
- Encryption: AES-GCM, ChaCha20-Poly1305 only
- Hashing: Argon2 preferred, Bcrypt acceptable
- Sessions: Regenerate after auth
- Never log sensitive data

## Common Gotchas

- Framework handlers return errors, stdlib doesn't
- Empty handlers return 204 No Content
- Middleware order: logger/recover first
- Session middleware required before access
- No nested database transactions
- Cache Remember: compute function signature matters
