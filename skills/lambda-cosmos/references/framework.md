# Framework

`framework` composes router + contract + problem into an error-returning HTTP stack.

```bash
go get github.com/studiolambda/cosmos/framework
```

## Core model

```go
type Handler func(http.ResponseWriter, *http.Request) error
type Middleware = router.Middleware[Handler]
```

`framework.New()` returns `*framework.Router`.

---

## Error pipeline (`Handler.ServeHTTP`)

For a returned error:

1. `context.Canceled` / `context.DeadlineExceeded` -> status `499`.
2. If error implements `framework.HTTPStatus` -> use that status.
3. If error implements `http.Handler` -> call `ServeHTTP` on the error.
4. Else -> `problem.NewProblem(err, status)` and serve.

If no error and nothing was written -> `204 No Content`.

After-response hooks run at the end.

---

## Hooks lifecycle

`framework.NewHooks()` provides request lifecycle hooks:

```go
hooks.BeforeWriteHeader(func(w http.ResponseWriter, status int) {
	_ = status
	w.Header().Set("X-App", "cosmos")
})

hooks.BeforeWrite(func(w http.ResponseWriter, content []byte) {
	_ = w
	_ = len(content)
})

hooks.AfterResponse(func(err error) {
	if err != nil {
		logger.Error("request failed", "error", err)
	}
})
```

Execution order is LIFO via reversed clones from `*Funcs()` methods.

---

## Response writer wrapping

```go
wrapped := framework.NewResponseWriter(w, hooks)
if !wrapped.WriteHeaderCalled() {
	wrapped.WriteHeader(http.StatusNoContent)
}
```

If original writer supports `http.Flusher`, wrapped value preserves flush support.

---

## Secure server helpers

```go
app := framework.New()
server := framework.NewServer(":8080", app)
if err := server.ListenAndServe(); err != nil {
	return err
}
```

Custom options:

```go
opts := framework.DefaultServerOptions()
opts.Addr = ":8443"
server := framework.NewServerWith(opts, app)
_ = server
```

---

## Built-in middleware

```go
app.Use(middleware.Recover())
app.Use(middleware.Logger(slog.Default()))
app.Use(middleware.CSRF("https://example.com"))
app.Use(middleware.CORS(middleware.CORSOptions{}))
app.Use(middleware.SecureHeaders())
app.Use(middleware.RateLimit())
app.Use(middleware.Provide("db", db))
app.Use(middleware.HTTP(stdlibMiddleware))
```

Recommended order near the top: Recover, Logger.

---

## Correlation package

```go
app.Use(correlation.Middleware())
logger := slog.New(correlation.Handler(baseHandler))
id := correlation.From(r)
_ = logger
_ = id
```

---

## Session package

`session.Middleware` requires a `contract.SessionDriver`.

### Cache-backed session driver setup

```go
cacheDriver := cache.NewMemory(5*time.Minute, 10*time.Minute)
typedCache := contract.NewCache(cacheDriver)
sessionDriver := session.NewCacheDriver(typedCache)

app.Use(session.Middleware(sessionDriver))
```

Custom middleware options:

```go
app.Use(session.MiddlewareWith(sessionDriver, session.MiddlewareOptions{
	Name:            "my_session",
	Path:            "/",
	Domain:          "example.com",
	Secure:          true,
	SameSite:        http.SameSiteLaxMode,
	Partitioned:     false,
	TTL:             24 * time.Hour,
	MaxLifetime:     24 * time.Hour,
	ExpirationDelta: 30 * time.Minute,
}))
```

Default constants:

- `session.DefaultCookie`
- `session.DefaultTTL`
- `session.DefaultMaxLifetime`
- `session.DefaultExpirationDelta`

---

## Cache package

Cache backends implement `contract.CacheDriver` (and counters when supported).
Wrap with `contract.NewCache` for typed API.

```go
memDriver := cache.NewMemory(5*time.Minute, 10*time.Minute)
redisDriver := cache.NewRedis(&cache.RedisOptions{Addr: "localhost:6379"})

c := contract.NewCache(memDriver)
value, err := c.Remember(ctx, "key", time.Minute, func() (string, error) {
	return "computed", nil
})
if err != nil {
	return err
}

_ = redisDriver
_ = value
```

---

## Crypto package

```go
aes, err := crypto.NewAES(key)      // key: 16/24/32 bytes
if err != nil {
	return err
}
defer func() { _ = aes.Close() }()

aes.AdditionalData = []byte("context")

ciphertext, err := aes.Encrypt(plaintext)
if err != nil {
	return err
}

cc, err := crypto.NewChaCha20(chachaKey) // key: 32 bytes
if err != nil {
	return err
}
defer func() { _ = cc.Close() }()

_ = ciphertext
_ = cc
```

---

## Hash package

```go
argon := hash.NewArgon2()
bcryptHasher := hash.NewBcrypt() // default cost = hash.DefaultBcryptCost (12)

hashed, err := argon.Hash(password)
if err != nil {
	return err
}

ok, err := argon.Check(password, hashed)
if err != nil {
	return err
}

if argon.NeedsRehash(hashed) {
	newHash, err := argon.Hash(password)
	if err != nil {
		return err
	}
	_ = newHash
}

_ = bcryptHasher
_ = ok
```

---

## Database package

`framework/database` provides SQL driver implementations (`contract.DatabaseDriver`).
Use `contract.NewDatabase` for typed convenience.

```go
sqlDriver, err := database.NewSQL("postgres", dsn)
if err != nil {
	return err
}

db := contract.NewDatabase(sqlDriver)

user, err := db.Find[User](ctx, "SELECT * FROM users WHERE id = $1", id)
if err != nil {
	return err
}

if err := db.WithTransaction(ctx, func(tx *contract.Database) error {
	_, err := tx.Exec(ctx, "UPDATE users SET seen_at = NOW() WHERE id = $1", id)
	return err
}); err != nil {
	return err
}

_ = user
```

---

## Event package

Brokers in `framework/event` implement `contract.EventDriver`:

- Memory: `event.NewMemoryBroker()`
- Redis: `event.NewRedisBroker(...)`
- NATS: `event.NewNATSBroker(...)`
- AMQP: `event.NewAMQPBroker(...)`
- MQTT: `event.NewMQTTBroker(...)`

Use `contract.NewEvents` for typed JSON publish/subscribe:

```go
driver := event.NewMemoryBroker()
ev := contract.NewEvents(driver)

ack, err := ev.Subscribe[UserCreated](ctx, "user.created", func(decode contract.EventDecoder[UserCreated]) {
	msg, err := decode()
	if err != nil {
		return
	}
	_ = msg
})
if err != nil {
	return err
}
defer ack()

if err := ev.Publish(ctx, "user.created", UserCreated{ID: 42}); err != nil {
	return err
}
```

---

## Gotchas

- `session.NewCacheDriver` expects `*contract.Cache`, not a raw cache driver.
- `contract.Database.WithTransaction` callback receives `*contract.Database`.
- Framework handlers must return errors for centralized handling.
- If a handler writes partial response and then errors, framework logs the error (cannot safely re-render response).
