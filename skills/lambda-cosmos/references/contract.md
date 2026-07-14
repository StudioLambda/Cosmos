# Contract

The `contract` module provides shared interfaces and typed wrappers used across
Cosmos.

```bash
go get github.com/studiolambda/cosmos/contract
```

## Mental model

There are two layers:

1. **Driver interfaces** for backend adapters:
   - `CacheDriver`
   - `DatabaseDriver`
   - `EventDriver`
   - `SessionDriver`
2. **Typed wrappers** over drivers:
   - `*contract.Cache`
   - `*contract.Database`
   - `*contract.Events`

Plus shared abstractions:

- `contract.Encrypter`
- `contract.Hasher` (+ `contract.Rehashable`)
- `*contract.Session`
- `contract.Hooks`

---

## Cache

### Driver interface

```go
type CacheDriver interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Put(ctx context.Context, key string, value []byte, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	Has(ctx context.Context, key string) (bool, error)
	Ping(ctx context.Context) error
}
```

Optional atomic counters:

```go
type CacheCounter interface {
	Increment(ctx context.Context, key string, delta int64) (int64, error)
	Decrement(ctx context.Context, key string, delta int64) (int64, error)
}
```

### Typed wrapper

```go
driver := cache.NewMemory(5*time.Minute, 10*time.Minute)
c := contract.NewCache(driver)

if err := c.Put(ctx, "users:1", User{ID: 1}, time.Minute); err != nil {
	return err
}

user, err := c.Get[User](ctx, "users:1")
if err != nil {
	return err
}

profile, err := c.Remember(ctx, "profiles:1", time.Minute, func() (Profile, error) {
	return loadProfile(ctx, 1)
})
if err != nil {
	return err
}

_ = user
_ = profile
```

Sentinel errors:

- `contract.ErrCacheKeyNotFound`
- `contract.ErrCacheUnsupportedOperation`

---

## Database

### Driver interface

```go
type DatabaseDriver interface {
	Close() error
	Ping(ctx context.Context) error
	Exec(ctx context.Context, query string, args ...any) (int64, error)
	ExecNamed(ctx context.Context, query string, arg any) (int64, error)
	Select(ctx context.Context, query string, dest any, args ...any) error
	SelectNamed(ctx context.Context, query string, dest any, arg any) error
	Find(ctx context.Context, query string, dest any, args ...any) error
	FindNamed(ctx context.Context, query string, dest any, arg any) error
	WithTransaction(ctx context.Context, fn func(tx DatabaseDriver) error) error
}
```

### Typed wrapper

```go
driver, err := database.NewSQL("postgres", dsn)
if err != nil {
	return err
}

db := contract.NewDatabase(driver)

user, err := db.Find[User](ctx, "SELECT * FROM users WHERE id = $1", id)
if err != nil {
	return err
}

users, err := db.Select[[]User](ctx, "SELECT * FROM users WHERE active = $1", true)
if err != nil {
	return err
}

if err := db.WithTransaction(ctx, func(tx *contract.Database) error {
	_, err := tx.Exec(ctx, "UPDATE users SET last_seen = NOW() WHERE id = $1", id)
	return err
}); err != nil {
	return err
}

_ = user
_ = users
```

Sentinel errors:

- `contract.ErrDatabaseNoRows`
- `contract.ErrDatabaseNestedTransaction`

---

## Events

### Driver interface

```go
type EventHandler = func(payload []byte)
type EventUnsubscribeFunc = func() error
type EventDecoder[T any] = func() (T, error)

type EventDriver interface {
	Publish(ctx context.Context, event string, payload []byte) error
	Subscribe(ctx context.Context, event string, handler EventHandler) (EventUnsubscribeFunc, error)
	Close() error
}
```

### Typed wrapper

```go
driver := event.NewMemoryBroker()
ev := contract.NewEvents(driver)

unsubscribe, err := ev.Subscribe[UserCreated](ctx, "users.created", func(decode contract.EventDecoder[UserCreated]) {
	msg, err := decode()
	if err != nil {
		return
	}
	_ = msg
})
if err != nil {
	return err
}
defer unsubscribe()

if err := ev.Publish(ctx, "users.created", UserCreated{ID: 1}); err != nil {
	return err
}
```

---

## Session

### SessionDriver interface

```go
type SessionDriver interface {
	Get(ctx context.Context, id string) (*Session, error)
	Save(ctx context.Context, session *Session, ttl time.Duration) error
	Delete(ctx context.Context, id string) error
}
```

### Session API

`*contract.Session` is concurrency-safe and generic:

```go
session, err := contract.NewSession(time.Now().Add(24*time.Hour), map[string]any{})
if err != nil {
	return err
}

session.Put("user_id", 123)
userID, err := session.Get[int]("user_id")
if err != nil {
	return err
}

if err := session.Regenerate(); err != nil {
	return err
}

_ = userID
```

Use `Regenerate()` after auth state changes.

---

## Crypto / Hash interfaces

```go
type Encrypter interface {
	Encrypt(value []byte) ([]byte, error)
	Decrypt(value []byte) ([]byte, error)
	Close() error
}

type Hasher interface {
	Hash(value []byte) ([]byte, error)
	Check(value []byte, hash []byte) (bool, error)
}

type Rehashable interface {
	NeedsRehash(hash []byte) bool
}
```

`contract.ErrEncrypterClosed` is returned when `Encrypt` or `Decrypt` is
called after `Close`.

---

## Hooks and request context keys

Context keys exported by `contract`:

- `contract.HooksKey`
- `contract.SessionKey`
- `contract.CorrelationIDKey`

`contract.Hooks` interface:

```go
type Hooks interface {
	AfterResponse(callbacks ...AfterResponseHook)
	AfterResponseFuncs() []AfterResponseHook
	BeforeWrite(callbacks ...BeforeWriteHook)
	BeforeWriteFuncs() []BeforeWriteHook
	BeforeWriteHeader(callbacks ...BeforeWriteHeaderHook)
	BeforeWriteHeaderFuncs() []BeforeWriteHeaderHook
}
```

---

## Request helpers (`contract/request`)

Highlights:

```go
id := request.Param(r, "id")
idInt, err := request.ParamInt(r, "id")

q := request.QueryOr(r, "q", "")
page := request.QueryIntOr(r, "page", 1)

payload, err := request.StrictLimitedJSON[Input](r, -1)

session, ok := request.Session(r)
mustSession := request.MustSession(r) // panics if missing

hooks := request.Hooks(r)             // panics if missing
safeHooks, ok := request.TryHooks(r)  // non-panicking

corrID := request.CorrelationID(r)

_ = id
_ = idInt
_ = q
_ = page
_ = payload
_ = session
_ = mustSession
_ = hooks
_ = safeHooks
_ = corrID
```

Body helpers consume request body once.

---

## Response helpers (`contract/response`)

All return `error`:

```go
return response.Status(w, http.StatusNoContent)
return response.JSON(w, http.StatusOK, data)
return response.XML(w, http.StatusOK, data)
return response.HTML(w, http.StatusOK, "<h1>ok</h1>")
return response.SafeRedirect(w, http.StatusFound, "/dashboard")
return response.Stream(w, r, chunks)
return response.SSE(w, r, events)
```

Prefer `response.SafeRedirect` for user-influenced redirect targets.

---

## Gotchas

- Use wrappers (`contract.NewCache`, `contract.NewDatabase`, `contract.NewEvents`) when you want typed generic operations.
- `request.MustSession` and `request.Hooks` panic if middleware/context is missing.
- Session `Get[T]` returns `(T, error)`, not `(any, bool)`.
- `response.JSON` uses `json.Encoder`, so output ends with a trailing newline.
