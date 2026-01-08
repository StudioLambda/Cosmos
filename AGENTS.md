# AGENTS.md

## Project Overview

Cosmos is a Go monorepo with four HTTP modules: contract (interfaces), router (HTTP routing), problem (RFC 9457 errors), framework (complete framework). Go 1.25.0 workspace with replace directives.

Dependency hierarchy: contract (zero deps) → router/problem (standalone) → framework (uses all)

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

## Code Style

- Standard Go formatting with gofmt
- Descriptive names, early returns, single-purpose functions
- Doc comments starting with name being documented
- Always check errors with errors.Is() and errors.As()
- Table-driven tests with testify assertions

### Framework Patterns

Error-returning handlers:
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

Sessions:
```go
session := request.Session(r)
session.Put("user_id", 123)
session.Regenerate() // After auth
```

## Common Gotchas

- Framework handlers return errors, stdlib handlers don't
- Middleware order matters: logger/recover first
- Session middleware required before accessing sessions
- Test from workspace root, not module directories
- Database transactions cannot be nested
- Problem methods return new instances (immutable)
- Router generic type parameter required: `router.New[http.Handler]()`

## Security

- Use middleware.CSRF() for state-changing endpoints
- Prefer Argon2 for password hashing
- Use AES-GCM or ChaCha20-Poly1305 (authenticated encryption)
- Regenerate sessions after authentication
- Never log sensitive data

## File Locations

- Workspace: go.work
- Module configs: */go.mod
- Mock config: contract/.mockery.yml
- CI workflows: .github/workflows/*.yml
- Contracts: contract/*.go
- Framework core: framework/handler.go, framework/framework.go
- Router: router/router.go
- Problem: problem/problem.go
