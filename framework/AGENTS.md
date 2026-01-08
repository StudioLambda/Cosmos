# AGENTS.md

## Module Overview

Framework module: complete HTTP framework with error-returning handlers, middleware, sessions, caching, crypto, hashing, database. Built on router, problem, and contract modules.

Module: github.com/studiolambda/cosmos/framework
Dependencies: router, problem, contract, sqlx, go-redis, go-cache, golang.org/x/crypto

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
app.Use(middleware.Logger(slog.Default()))
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
app.Use(session.Middleware(driver, "session_id"))

sess := request.Session(r)
sess.Put("user_id", 123)
sess.Regenerate()
```

Cache:
```go
cache := cache.NewMemory(5*time.Minute, 10*time.Minute)
cache.Remember(ctx, key, ttl, compute)
```

Crypto:
```go
aes := crypto.NewAES(key) // 16, 24, or 32 bytes
ciphertext, err := aes.Encrypt(ctx, plaintext)
```

Hash:
```go
hasher := hash.NewArgon2()
hashed, err := hasher.Hash(ctx, password)
err := hasher.Verify(ctx, password, hashed)
```

Database:
```go
db := database.NewSQL("postgres", connString)
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
