# Cosmos: Contract

`contract` defines raw driver interfaces and typed facades for application services, plus request and response helpers. It depends on `collection` and `problem`.

```bash
go get github.com/studiolambda/cosmos/contract
```

Requires Go 1.27 or later.

## Drivers and Facades

Drivers implement transport or storage behavior; facades add typed JSON serialization and convenience methods.

- `CacheDriver` operates on raw bytes; `contract.NewCache(driver)` provides `Get[T]`, `Put`, and `Remember[T]`.
- `DatabaseDriver` performs SQL operations; `contract.NewDatabase(driver)` provides typed `Find[T]`, `Select[T]`, cursors, and transactions.
- `EncrypterDriver` encrypts raw bytes; `contract.NewEncrypter(driver)` JSON-encodes typed values. Use `EncryptRaw` and `DecryptRaw` for bytes.
- `HasherDriver` hashes raw bytes; `contract.NewHasher(driver)` JSON-encodes typed values. Use `HashRaw` and `CheckRaw` for passwords or other byte values.
- `SecretDriver` retrieves raw secrets; `contract.NewSecrets(driver)` provides `Raw`, `String`, and JSON-decoding `Get[T]`.
- `ConfigurationDriver` backs `contract.NewConfiguration(driver)` and typed `Get[T]` access.

```go
cache := contract.NewCache(driver)
user, err := cache.Remember(ctx, "users:1", time.Hour, func() (User, error) {
	return loadUser(ctx, 1)
})
```

`HasherDriver` implementations clear the raw value passed to `Hash` and `Check`; do not reuse that byte slice. `Encrypter.Close` releases encryption resources and clears retained key material where possible.

## Sessions

`contract.Session` is a concurrency-safe value store persisted by a `SessionDriver`. It has typed accessors:

```go
session.Put("user_id", 42)
userID, err := session.Get[int]("user_id")
session.Regenerate()
```

Retrieve a session from an HTTP request with `request.Session(r)`, which returns `(*contract.Session, bool)`. `request.MustSession(r)` panics when session middleware is absent. Regenerate after authentication, logout, or privilege changes.

## HTTP Helpers

`contract/request` supplies typed path and query parsing, size-limited JSON helpers, hooks, correlation IDs, cookies, and session retrieval. `contract/response` supplies JSON, HTML, XML, streams, files, SSE, and `SafeRedirect`.

## Testing

Run from the workspace root:

```bash
go test ./contract/...
go test -race ./contract/...
```

Generated mocks are in `contract/mock`; regenerate them with `cd contract && go generate ./...`.
