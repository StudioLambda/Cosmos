---
name: cosmos-contract
description: >
  Interfaces and helpers for the Cosmos HTTP framework contract module
  (github.com/studiolambda/cosmos/contract). Use when consuming or
  implementing Cosmos contracts: Cache, Database, Encrypter, Hasher,
  Session, SessionDriver, Events, Hooks. Also covers the request and
  response helper packages for reading HTTP inputs and writing outputs.
---

# Cosmos Contract

The contract module defines all service interfaces and HTTP helper functions
for the Cosmos framework. It is the foundation layer — every other Cosmos
module depends on it.

```
go get github.com/studiolambda/cosmos/contract
```

## Architecture

```
contract/
├── cache.go       # Cache interface (10 methods)
├── crypto.go      # Encrypter interface
├── database.go    # Database interface (transactions, named queries)
├── event.go       # Events interface + type aliases
├── hash.go        # Hasher interface
├── hooks.go       # Hooks interface (lifecycle hooks)
├── session.go     # Session + SessionDriver interfaces
├── request/       # HTTP request helpers (params, query, body, headers, cookies, sessions)
├── response/      # HTTP response helpers (JSON, HTML, XML, SSE, redirects)
└── mock/          # Mockery-generated mocks for all interfaces
```

## Interfaces

### Cache

```go
type Cache interface {
    Get(ctx context.Context, key string) (any, error)
    Put(ctx context.Context, key string, value any, ttl time.Duration) error
    Delete(ctx context.Context, key string) error
    Has(ctx context.Context, key string) (bool, error)
    Pull(ctx context.Context, key string) (any, error)
    Forever(ctx context.Context, key string, value any) error
    Increment(ctx context.Context, key string, by int64) (int64, error)
    Decrement(ctx context.Context, key string, by int64) (int64, error)
    Remember(ctx context.Context, key string, ttl time.Duration, compute func() (any, error)) (any, error)
    RememberForever(ctx context.Context, key string, compute func() (any, error)) (any, error)
}
```

Sentinel errors: `ErrCacheKeyNotFound`, `ErrCacheUnsupportedOperation`.

### Database

```go
type Database interface {
    Close() error
    Ping(ctx context.Context) error
    Exec(ctx context.Context, query string, args ...any) (int64, error)
    ExecNamed(ctx context.Context, query string, arg any) (int64, error)
    Select(ctx context.Context, query string, dest any, args ...any) error
    SelectNamed(ctx context.Context, query string, dest any, arg any) error
    Find(ctx context.Context, query string, dest any, args ...any) error
    FindNamed(ctx context.Context, query string, dest any, arg any) error
    WithTransaction(ctx context.Context, fn func(Database) error) error
}
```

Sentinel errors: `ErrDatabaseNoRows`, `ErrDatabaseNestedTransaction`.
Transactions cannot be nested. `Find` joins `sql.ErrNoRows` with
`ErrDatabaseNoRows` — check with either via `errors.Is`.

### Encrypter

```go
type Encrypter interface {
    Encrypt(value []byte) ([]byte, error)
    Decrypt(value []byte) ([]byte, error)
}
```

### Hasher

```go
type Hasher interface {
    Hash(value []byte) ([]byte, error)
    Check(value, hash []byte) (bool, error)
}
```

### Session & SessionDriver

```go
type Session interface {
    SessionID() string
    OriginalSessionID() string
    Get(key string) (any, bool)
    Put(key string, value any)
    Delete(key string)
    Extend(expiresAt time.Time)
    Regenerate() error
    Clear()
    ExpiresAt() time.Time
    HasExpired() bool
    ExpiresSoon(delta time.Duration) bool
    HasChanged() bool
    HasRegenerated() bool
    MarkAsUnchanged()
}

type SessionDriver interface {
    Get(ctx context.Context, id string) (Session, error)
    Save(ctx context.Context, session Session, ttl time.Duration) error
    Delete(ctx context.Context, id string) error
}
```

### Events

```go
type EventPayload  = func(dest any) error
type EventHandler  = func(payload EventPayload)
type EventUnsubscribeFunc = func() error

type Events interface {
    Publish(ctx context.Context, event string, payload any) error
    Subscribe(ctx context.Context, event string, handler EventHandler) (EventUnsubscribeFunc, error)
    Close() error
}
```

### Hooks

```go
type AfterResponseHook    = func(err error)
type BeforeWriteHeaderHook = func(w http.ResponseWriter, status int)
type BeforeWriteHook       = func(w http.ResponseWriter, content []byte)

type Hooks interface {
    AfterResponse(hooks ...AfterResponseHook)
    AfterResponseFuncs() []AfterResponseHook
    BeforeWrite(hooks ...BeforeWriteHook)
    BeforeWriteFuncs() []BeforeWriteHook
    BeforeWriteHeader(hooks ...BeforeWriteHeaderHook)
    BeforeWriteHeaderFuncs() []BeforeWriteHeaderHook
}
```

Hooks are injected into request context by the framework. Access via
`request.Hooks(r)`. Hook `*Funcs()` methods return reversed clones (LIFO
execution order).

## Request Helpers

All functions in `contract/request` take `*http.Request` as first argument.

### Parameters & Query

```go
request.Param(r, "id")              // URL path parameter
request.ParamOr(r, "id", "default") // with fallback
request.Query(r, "page")            // query string value
request.QueryOr(r, "page", "1")     // with fallback
request.HasQuery(r, "verbose")      // existence check
```

### Headers & Cookies

```go
request.Header(r, "Authorization")
request.HeaderOr(r, "Accept", "application/json")
request.HasHeader(r, "X-Custom")
request.HeaderValues(r, "Accept")        // []string
request.Cookie(r, "session")             // *http.Cookie
request.CookieValue(r, "session")        // string
request.CookieValueOr(r, "session", "") // with fallback
```

### Body Parsing (generics)

```go
bytes, err := request.Bytes(r)
text, err := request.String(r)
user, err := request.JSON[User](r)   // generic — type param required
order, err := request.XML[Order](r)  // generic — type param required
```

Body is consumed on first read — call only once per request.

### Sessions

```go
sess, ok := request.Session(r)           // returns (Session, bool)
sess := request.MustSession(r)           // panics if no session middleware
sess, ok := request.SessionKeyed(r, key) // custom context key
sess := request.MustSessionKeyed(r, key) // panics variant
```

`MustSession` and `MustSessionKeyed` panic when session middleware is not
applied — these are programmer errors caught during development.

## Response Helpers

All functions in `contract/response` return `error` — usable as direct
handler return values.

```go
return response.Status(w, http.StatusNoContent)
return response.String(w, http.StatusOK, "hello")
return response.HTML(w, http.StatusOK, "<h1>hi</h1>")
return response.JSON(w, http.StatusOK, user)           // generic
return response.XML(w, http.StatusOK, order)
return response.Bytes(w, http.StatusOK, data)
return response.Raw(w, http.StatusOK, rawBytes)
return response.Redirect(w, http.StatusFound, "/login")

// Templates
return response.StringTemplate(w, http.StatusOK, tmpl, data)
return response.HTMLTemplate(w, http.StatusOK, tmpl, data)

// Streaming
return response.Stream(w, r, dataChan)  // raw streaming
return response.SSE(w, r, eventChan)    // Server-Sent Events
```

`response.JSON` uses `json.NewEncoder` which appends a trailing newline.

## Mocks

Generated mocks in `contract/mock/` for all interfaces. Each has a
`New<Name>Mock(t)` constructor that auto-registers cleanup.

```go
cache := mock.NewCacheMock(t)
cache.On("Get", mock.Anything, "key").Return("value", nil)
```

## Gotchas

- `request.Hooks(r)` **panics** without hooks context (only present inside
  framework's handler pipeline).
- `request.MustSession(r)` **panics** without session middleware.
- Body parsing functions consume the request body — one call per request.
- `response.JSON` appends a trailing newline (standard `json.NewEncoder`
  behavior).
- `EventPayload` is a function — call it with a pointer to unmarshal:
  `payload(&user)`.
- Context keys (`HooksKey`, `SessionKey`) are exported variables of
  unexported types — use the provided helpers, don't construct context
  values directly.
