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
2. If error implements `framework.HTTPStatusError` -> use that status.
3. If error implements `http.Handler` -> call `ServeHTTP` on the error.
4. Else -> `problem.NewDetails(err, status)` and serve.

If no error and nothing was written -> `204 No Content`.

After-response hooks run at the end.

---

## Hooks lifecycle

`contract.NewHooks()` provides request lifecycle hooks:

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
server := framework.NewServer(framework.ServerConfig{}, app)
if err := server.ListenAndServe(); err != nil {
	return err
}
```

Custom options:

```go
config := framework.DefaultServerConfig()
config.Port = 8443
server := framework.NewServer(config, app)
_ = server
```

---

## Built-in middleware

```go
app.Use(middleware.Recover())
app.Use(middleware.Logger(contract.NewLogger(frameworklogger.NewSlogFrom(slog.Default()))))
app.Use(middleware.CSRF(middleware.CSRFConfig{TrustedOrigins: []string{"https://example.com"}}))
app.Use(middleware.CORS(middleware.CORSConfig{}))
app.Use(middleware.SecureHeaders(middleware.DefaultSecureHeadersConfig))
app.Use(middleware.RateLimit(contract.NewCache(memory.NewMemory(memory.MemoryConfig{})), middleware.RateLimitConfig{}))

// Rate limiting uses cache-backed fixed-window counters.
app.Use(middleware.Provide("db", db))
app.Use(middleware.HTTP(stdlibMiddleware))
```

Recommended order near the top: Recover, Logger.

---

## Correlation package

```go
app.Use(middleware.Correlation(middleware.DefaultCorrelationConfig))
id := request.CorrelationID(r)
_ = id
```

---

## Session package

`middleware.Session` requires a `contract.SessionDriver`.

### Cache-backed session driver setup

```go
cacheDriver := memory.NewMemory(memory.MemoryConfig{})
typedCache := contract.NewCache(cacheDriver)
sessionDriver := session.NewCacheDriver(typedCache, session.CacheDriverConfig{})

app.Use(middleware.Session(sessionDriver, middleware.DefaultSessionConfig))
```

Custom middleware options:

```go
app.Use(middleware.Session(sessionDriver, middleware.SessionConfig{
	Name:            "my_session",
	Path:            "/",
	Domain:          "example.com",
	Secure:          true,
	SameSite:        "lax",
	Partitioned:     false,
	TTL:             24 * time.Hour,
	MaxLifetime:     24 * time.Hour,
	ExpirationDelta: 30 * time.Minute,
}))
```

Default constants:

- `middleware.DefaultSessionCookie`
- `middleware.DefaultSessionTTL`
- `middleware.DefaultSessionMaxLifetime`
- `middleware.DefaultSessionExpirationDelta`

---

## Cache package

Cache backends implement `contract.CacheDriver` (and counters when supported).
Wrap with `contract.NewCache` for typed API.

```go
memDriver := memory.NewMemory(memory.MemoryConfig{})
redisDriver := redis.NewRedis(redis.RedisConfig{Addr: "localhost:6379"})

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
aes, err := aes.NewAES(aes.AESConfig{Key: key}) // key: 16/24/32 bytes
if err != nil {
	return err
}
defer func() { _ = aes.Close() }()

aes.AdditionalData = []byte("context")

ciphertext, err := aes.Encrypt(plaintext)
if err != nil {
	return err
}

cc, err := chacha20.NewChaCha20(chacha20.ChaCha20Config{Key: chachaKey}) // key: 32 bytes
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
argon := argon2.NewArgon2(argon2.DefaultArgon2Config())
bcryptHasher := bcrypt.NewBcrypt(bcrypt.DefaultBcryptConfig) // default cost = bcrypt.DefaultBcryptCost (12)

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
sqlDriver, err := postgres.New(postgres.Config{DSN: dsn})
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

Brokers in `framework/event` implement `contract.EventPublisherDriver` and
`contract.EventSubscriberDriver`:

- Memory: `memory.NewMemoryBroker(memory.MemoryBrokerConfig{})`
- Redis: `redis.NewRedisBroker(redis.RedisBrokerConfig{})`
- NATS: `nats.NewNATSBroker(nats.NATSBrokerConfig{})`
- AMQP: `amqp.NewAMQPBroker(amqp.AMQPBrokerConfig{})`
- MQTT: `mqtt.NewMQTTBroker(mqtt.MQTTBrokerConfig{})`

Use `contract.NewEventPublisher` and `contract.NewEventSubscriber` for typed
JSON publish/subscribe:

```go
driver := memory.NewMemoryBroker(memory.MemoryBrokerConfig{})
publisher := contract.NewEventPublisher(driver)
subscriber := contract.NewEventSubscriber(driver)

ack, err := subscriber.Subscribe[UserCreated](ctx, "user.created", func(decode contract.EventDecoder[UserCreated]) {
	msg, err := decode()
	if err != nil {
		return
	}
	_ = msg
})
if err != nil {
	return err
}
defer func() { _ = ack() }()

if err := publisher.Publish(ctx, "user.created", UserCreated{ID: 42}); err != nil {
	return err
}
```

---

## Gotchas

- `session.NewCacheDriver` expects `*contract.Cache` and a `session.CacheDriverConfig`.
- `contract.Database.WithTransaction` callback receives `*contract.Database`.
- Framework handlers must return errors for centralized handling.
- If a handler writes partial response and then errors, framework logs the error (cannot safely re-render response).
