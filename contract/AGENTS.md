# AGENTS.md

## Module Overview

Contract module: service interfaces (Cache, Database, Session, Crypto, Hash) and request/response helpers. Zero dependencies except testing. Foundation for all Cosmos modules.

Module: github.com/studiolambda/cosmos/contract
Dependencies: Zero (testify for tests, mockery as tool)

## Setup Commands

```bash
go test ./...
go test -cover ./...
go generate ./...  # Generate mocks
go fmt ./...
```

## Architecture

Five interfaces: Cache, Database, Session, Crypto, Hash. All accept context.Context. Standard errors: ErrCacheKeyNotFound, ErrDatabaseNoRows, ErrDatabaseNestedTransaction.

Helpers in request/ and response/ packages. Hooks system for middleware lifecycle.

## Code Style

Interfaces accept context first, return (result, error) or error. Sentinel errors as variables, compare with errors.Is(). Helpers don't panic except for programmer errors.

Zero dependencies requirement: no external imports in non-test code.

## Common Patterns

Cache Remember:
```go
cache.Remember(ctx, key, ttl, func() (any, error) {
    return expensiveOp(), nil
})
```

Database single row:
```go
var user User
err := db.Find(ctx, "SELECT * FROM users WHERE id = $1", &user, id)
```

Database transaction:
```go
db.WithTransaction(ctx, func(tx contract.Database) error {
    _, err := tx.Exec(ctx, query, args...)
    return err // auto rollback/commit
})
```

Request helpers:
```go
id := request.Param(r, "id")
page := request.Query(r, "page")
auth := request.Header(r, "Authorization")
err := request.Body(r, &data)
```

Response helpers:
```go
return response.JSON(w, http.StatusOK, data)
```

Hooks:
```go
hooks := request.Hooks(r)
hooks.AfterResponse(func(err error) { log(err) })
```

Session:
```go
session.Put("user_id", 123)
if val, ok := session.Get("user_id"); ok { }
session.Regenerate()
```

Mock usage:
```go
mockCache := &mock.Cache{}
mockCache.On("Get", mock.Anything, "key").Return("value", nil)
```

## Testing

Table-driven tests with testify. Regenerate mocks after interface changes: `go generate ./...`

## Package Structure

```
contract/
├── cache.go, database.go, session.go, crypto.go, hash.go, hooks.go
├── request/    # Request helpers
├── response/   # Response helpers
└── mock/       # Generated mocks
```

## Common Gotchas

- Zero dependencies: no external imports
- Type assertions for session values: val.(int)
- Context always first parameter
- Mock expectations must match exactly
