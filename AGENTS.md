# AGENTS.md

## Project Overview

Cosmos is a Go monorepo with four HTTP modules: contract (interfaces), router (HTTP routing), problem (RFC 9457 errors), framework (complete framework). Go workspace with replace directives.

Dependency hierarchy: problem (zero deps) → contract (depends on problem) → router (standalone) → framework (uses all)

Current versions: contract v0.10.0, framework v0.11.0, problem v0.4.0, router v0.4.0

## Setup Commands

```bash
# Verify workspace and run all tests
go work sync && go test ./...

# Test specific module
go test ./contract/...
go test ./router/...
go test ./problem/...
go test ./framework/...

# Coverage and formatting
go test -cover ./...
go fmt ./...
go vet ./...

# Generate mocks (contract only)
cd contract && go generate ./...
```

## Module Navigation

Always run tests from workspace root for proper module resolution. Each module has its own go.mod. Use full import paths: `github.com/studiolambda/cosmos/contract`

## Framework Patterns

Error-returning handlers (unlike stdlib, Cosmos handlers return errors):

```go
func handler(w http.ResponseWriter, r *http.Request) error {
    return response.JSON(w, http.StatusOK, data)
}
```

Middleware composition:

```go
func MyMiddleware() framework.Middleware {
    return func(next framework.Handler) framework.Handler {
        return func(w http.ResponseWriter, r *http.Request) error {
            err := next(w, r)
            return err
        }
    }
}
```

Secure server (always use instead of http.ListenAndServe):

```go
server := framework.NewServer(":8080", app) // has timeout defaults
server.ListenAndServe()
```

## Common Patterns

Router:

```go
r := router.New[http.Handler]()
r.Use(middleware)
r.Get("/users/{id}", handler)
```

Problem Details:

```go
ErrNotFound.With("resource_id", id).ServeHTTP(w, r)
```

Cache:

```go
cache.Remember(ctx, key, ttl, func() (any, error) {
    return computeValue(), nil
})
```

Database transactions:

```go
db.WithTransaction(ctx, func(tx contract.Database) error {
    return tx.Exec(ctx, query, args...)
})
```

Database connection pool:

```go
db.Configure(func(raw *sql.DB) {
    raw.SetMaxOpenConns(25)
    raw.SetMaxIdleConns(5)
    raw.SetConnMaxLifetime(5 * time.Minute)
})
```

Sessions:

```go
session := request.MustSession(r)
session.Put("user_id", 123)
session.Regenerate() // After auth
```

Request helpers:

```go
id, err := request.ParamInt(r, "id")        // typed int parsing
body, err := request.LimitedJSON[T](r, -1)  // size-limited body (default 10MB)
body, err := request.StrictJSON[T](r)       // disallow unknown fields
hooks, ok := request.TryHooks(r)            // non-panicking hooks access
```

Response helpers:

```go
response.SafeRedirect(w, http.StatusFound, "/dashboard") // validates relative path
```

Encryption with AAD and key zeroing:

```go
enc, _ := crypto.NewAES(key)
enc.AdditionalData = []byte("context-binding")
ciphertext, _ := enc.Encrypt(plaintext)
defer enc.Close() // zeros key material
```

## Built-in Middleware

Available in `framework/middleware`:

- `Recover()` / `RecoverWith(fn)` — panic recovery with error wrapping
- `Logger(slog.Logger)` — structured request logging (fires AfterResponse)
- `CSRF(origins...)` / `CSRFWith(csrf, problem)` — cross-origin protection
- `CORS(CORSOptions)` — configurable CORS headers
- `SecureHeaders()` / `SecureHeadersWith(opts)` — security headers (HSTS, CSP, X-Frame-Options)
- `RateLimit()` / `RateLimitWith(opts)` — per-key token bucket (default 15 req/s, burst 30, idle eviction after 5m)
- `Provide(key, value)` / `ProvideWith(fn)` — context injection
- `HTTP(func(http.Handler) http.Handler)` — stdlib middleware adapter

Correlation ID in `framework/correlation`:

- `Middleware()` / `MiddlewareWith(opts)` — ensures every request has a correlation ID (W3C traceparent, header, or generated)
- `Handler(next)` — slog handler decorator that injects correlation ID into log records
- `From(r)` — retrieves correlation ID from request context

Session middleware in `framework/session`:

- `Middleware(driver)` / `MiddlewareWith(driver, opts)` — session lifecycle

## Common Gotchas

- Framework handlers return errors, stdlib handlers don't
- Middleware order matters: Recover and Logger first
- Session middleware required before accessing sessions
- Test from workspace root, not module directories
- Database transactions cannot be nested
- Problem methods return new instances (immutable copy-on-write)
- Router generic type parameter required: `router.New[http.Handler]()`
- `Defaulted()` does NOT auto-populate Detail from err.Error() (security: prevents info leakage)
- `Instance` uses `r.URL.Path` only (strips query string to prevent token leakage)
- `HTMLTemplate` uses `html/template` not `text/template` (XSS prevention)
- `MustSession` panics if session middleware is missing; prefer `Session()` (returns bool)
- Bcrypt default cost is 12 (OWASP recommendation), not 10
- Session IDs use `crypto/rand` + base64url (256-bit), not UUID
- All database methods use prepared statements with `defer Close()`
- `Any()` excludes TRACE and CONNECT methods (security: prevents XST)
- Recover middleware caps `io.Reader` panic values at 1MB and wraps all panic values with `ErrRecoverUnexpected`

## Security

- Use `framework.NewServer()` instead of `http.ListenAndServe` (timeout defaults)
- Use `middleware.CSRF()` for state-changing endpoints
- Use `middleware.SecureHeaders()` for all applications
- Use `middleware.RateLimit()` to prevent abuse
- Use `middleware.CORS()` for cross-origin APIs
- Use `request.LimitedJSON` / `LimitedBytes` instead of unlimited variants
- Use `response.SafeRedirect` instead of `response.Redirect` for user-supplied URLs
- Prefer Argon2 for password hashing
- Use AES-GCM or ChaCha20-Poly1305 (authenticated encryption)
- Call `enc.Close()` to zero key material when done
- Regenerate sessions after authentication
- Never log sensitive data
- Session `MaxLifetime` (default 24h) enforces absolute session age

## Test Coverage

| Module               | Coverage                     | Notes                            |
| -------------------- | ---------------------------- | -------------------------------- |
| contract/request     | 100%                         |                                  |
| contract/response    | 100%                         |                                  |
| router               | 100%                         | Dead code removed                |
| problem              | 100%                         |                                  |
| problem/internal     | 100%                         |                                  |
| framework (root)     | 100%                         |                                  |
| framework/middleware | 100%                         |                                  |
| framework/session    | 95.4%                        | Remaining: rand failure path     |
| framework/cache      | memory 100%, redis 0%        | Redis requires external server   |
| framework/crypto     | 90.2%                        | Remaining: rand failure paths    |
| framework/hash       | 95.8%                        | Remaining: bcrypt internal error |
| framework/event      | memory 100%, pure logic 100% | External brokers require servers |
| framework/database   | 0%                           | Pure sqlx adapter, requires DB   |

## File Locations

- Workspace: go.work
- Module configs: \*/go.mod
- Mock config: contract/.mockery.yml
- CI workflows: .github/workflows/\*.yml
- Contracts: contract/\*.go
- Request helpers: contract/request/\*.go
- Response helpers: contract/response/\*.go
- Framework core: framework/handler.go, framework/framework.go, framework/server.go
- Framework hooks: framework/hooks.go, framework/hooks_writer.go
- Router: router/router.go
- Problem: problem/problem.go
- Correlation: framework/correlation/middleware.go, framework/correlation/handler.go
- Middleware: framework/middleware/\*.go
- Session: framework/session/\*.go
- Cache: framework/cache/memory.go, framework/cache/redis.go
- Crypto: framework/crypto/aes.go, framework/crypto/chacha20.go
- Hash: framework/hash/argon2.go, framework/hash/bcrypt.go
- Database: framework/database/sql.go
- Events: framework/event/memory.go, framework/event/redis.go, framework/event/nats.go, framework/event/amqp.go, framework/event/mqtt.go
- Agent skills: .agents/skills/cosmos-\*/, skills/lambda-cosmos/
- Docs site: ../docs/src/content/docs/cosmos/
