# Subpackages

Detailed API reference for all framework subpackages. Load this
reference when working with sessions, caching, crypto, hashing,
database, or events.

## Table of Contents

- [Sessions](#sessions)
- [Cache](#cache)
- [Crypto](#crypto)
- [Hash](#hash)
- [Database](#database)
- [Events](#events)

---

## Sessions

Package: `github.com/studiolambda/cosmos/framework/session`

### Setup

```go
// Basic setup with in-memory cache
memCache := cache.NewMemory(5*time.Minute, 10*time.Minute)
driver := session.NewCacheDriver(memCache)
app.Use(session.Middleware(driver))
```

### Custom Configuration

```go
driver := session.NewCacheDriverWith(memCache, session.CacheDriverOptions{
    Prefix: "myapp.sessions",
})

app.Use(session.MiddlewareWith(driver, session.MiddlewareOptions{
    Name:            "my_session",        // cookie name
    Path:            "/",
    Domain:          "example.com",
    Secure:          true,
    SameSite:        http.SameSiteLaxMode,
    Partitioned:     false,
    TTL:             24 * time.Hour,
    ExpirationDelta: 30 * time.Minute,
    Key:             nil,                 // custom context key
}))
```

### Using Sessions in Handlers

```go
func handler(w http.ResponseWriter, r *http.Request) error {
    sess, ok := request.Session(r)
    if !ok {
        return ErrUnauthorized
    }

    // Read
    userID, exists := sess.Get("user_id")

    // Write
    sess.Put("user_id", 123)

    // Delete
    sess.Delete("temp_key")

    // Clear all data
    sess.Clear()

    // Regenerate ID (MUST do after authentication)
    sess.Regenerate()

    // Extend expiration
    sess.Extend(time.Now().Add(24 * time.Hour))

    return response.JSON(w, http.StatusOK, data)
}
```

### Session Lifecycle

- Sessions are loaded from the driver at the start of each request.
- Changes are persisted via a BeforeWriteHeader hook (fires just
  before the HTTP status line).
- New sessions get a UUID v7 ID and are marked as changed.
- `ExpirationDelta` controls automatic session extension: if the
  session expires within `delta` of the current time, it is
  automatically extended.

### Constants

| Constant | Value |
|---|---|
| `DefaultCookie` | `"cosmos.session"` |
| `DefaultTTL` | `2 * time.Hour` |
| `DefaultExpirationDelta` | `15 * time.Minute` |

---

## Cache

Package: `github.com/studiolambda/cosmos/framework/cache`

### Memory Cache

```go
mem := cache.NewMemory(
    5 * time.Minute,   // default expiration
    10 * time.Minute,  // cleanup interval
)
```

### Redis Cache

```go
rdb := cache.NewRedis(&cache.RedisOptions{
    Addr:     "localhost:6379",
    Password: "",
    DB:       0,
})

// Or from existing client
rdb := cache.NewRedisFrom(existingClient)
```

### Common Operations

Both implement `contract.Cache`:

```go
// Basic CRUD
err := c.Put(ctx, "key", "value", 5*time.Minute)
val, err := c.Get(ctx, "key")        // returns (any, error)
err := c.Delete(ctx, "key")
ok, err := c.Has(ctx, "key")
val, err := c.Pull(ctx, "key")       // get + delete

// Persistent storage
err := c.Forever(ctx, "key", "value") // no expiration

// Atomic counters
count, err := c.Increment(ctx, "visits", 1)
count, err := c.Decrement(ctx, "stock", 1)

// Compute-if-absent
val, err := c.Remember(ctx, "key", ttl, func() (any, error) {
    return expensiveComputation(), nil
})
val, err := c.RememberForever(ctx, "key", func() (any, error) {
    return staticValue(), nil
})
```

Sentinel errors: `contract.ErrCacheKeyNotFound`,
`contract.ErrCacheUnsupportedOperation`.

---

## Crypto

Package: `github.com/studiolambda/cosmos/framework/crypto`

### AES-GCM

```go
// Key must be 16 (AES-128), 24 (AES-192), or 32 (AES-256) bytes
aes, err := crypto.NewAES(key)

ciphertext, err := aes.Encrypt(plaintext)
plaintext, err := aes.Decrypt(ciphertext)
```

### ChaCha20-Poly1305

```go
// Key must be exactly 32 bytes
cc, err := crypto.NewChaCha20(key)

ciphertext, err := cc.Encrypt(plaintext)
plaintext, err := cc.Decrypt(ciphertext)
```

### Format

Both use authenticated encryption (AEAD). The output format is:

```
[nonce][ciphertext][authentication tag]
```

The nonce is randomly generated for each `Encrypt` call and prepended
to the output. `Decrypt` reads the nonce from the front.

Sentinel errors:
- `crypto.ErrMismatchedAESNonceSize` — ciphertext too short for AES.
- `crypto.ErrMismatchedChaCha20NonceSize` — ciphertext too short for
  ChaCha20.

---

## Hash

Package: `github.com/studiolambda/cosmos/framework/hash`

### Argon2id (Recommended)

```go
hasher := hash.NewArgon2()
// Or with custom config:
hasher := hash.NewArgon2With(hash.Argon2Config{
    // ... argon2 parameters
})

hashed, err := hasher.Hash(password)
ok, err := hasher.Check(password, hashed)
```

### Bcrypt

```go
hasher := hash.NewBcrypt()
// Or with custom cost:
hasher := hash.NewBcryptWith(hash.BcryptOptions{Cost: 12})

hashed, err := hasher.Hash(password)
ok, err := hasher.Check(password, hashed)
```

Default bcrypt cost: `hash.DefaultBcryptCost` (10).

Both implement `contract.Hasher`.

---

## Database

Package: `github.com/studiolambda/cosmos/framework/database`

### Setup

```go
db, err := database.NewSQL("postgres", "postgres://user:pass@localhost/mydb")
defer db.Close()

// Or from existing sqlx connection
db := database.NewSQLFrom(existingSqlxDB)

// Health check
err := db.Ping(ctx)
```

### Queries

```go
// Find single row (returns error if not found)
var user User
err := db.Find(ctx, "SELECT * FROM users WHERE id = $1", &user, id)

// Select multiple rows
var users []User
err := db.Select(ctx, "SELECT * FROM users WHERE active = $1", &users, true)

// Execute (INSERT, UPDATE, DELETE) — returns affected rows
rows, err := db.Exec(ctx, "DELETE FROM users WHERE id = $1", id)

// Named queries — use struct or map for parameters
rows, err := db.ExecNamed(ctx, "INSERT INTO users (name, email) VALUES (:name, :email)", user)
err := db.FindNamed(ctx, "SELECT * FROM users WHERE email = :email", &user, params)
err := db.SelectNamed(ctx, "SELECT * FROM users WHERE role = :role", &users, params)
```

### Transactions

```go
err := db.WithTransaction(ctx, func(tx contract.Database) error {
    _, err := tx.Exec(ctx, "UPDATE accounts SET balance = balance - $1 WHERE id = $2", amount, from)
    if err != nil {
        return err // triggers rollback
    }

    _, err = tx.Exec(ctx, "UPDATE accounts SET balance = balance + $1 WHERE id = $2", amount, to)

    return err // nil = commit, error = rollback
})
```

### Error Handling

- `Find` returns `errors.Join(sql.ErrNoRows, contract.ErrDatabaseNoRows)`
  when no rows found. Check with either:
  ```go
  if errors.Is(err, contract.ErrDatabaseNoRows) { ... }
  if errors.Is(err, sql.ErrNoRows) { ... }
  ```
- Nested transactions return `contract.ErrDatabaseNestedTransaction`.
- Rollback errors are joined with the original error via `errors.Join`.

---

## Events

Package: `github.com/studiolambda/cosmos/framework/event`

### Broker Implementations

| Broker | Backend | Constructor |
|---|---|---|
| `MemoryBroker` | In-memory | `NewMemoryBroker()` |
| `RedisBroker` | Redis Pub/Sub | `NewRedisBroker(options)` |
| `NATSBroker` | NATS | `NewNATSBroker(url)` |
| `AMQPBroker` | RabbitMQ | `NewAMQPBroker(url)` |
| `MQTTBroker` | MQTT v5 | `NewMQTTBroker(url)` |

All implement `contract.Events`.

### Usage

```go
broker := event.NewMemoryBroker()
defer broker.Close()

// Subscribe
unsub, err := broker.Subscribe(ctx, "user.*.created", func(payload contract.EventPayload) {
    var user User
    if err := payload(&user); err != nil {
        log.Error("failed to decode event", "error", err)
        return
    }
    // handle user creation
})
defer unsub()

// Publish
err := broker.Publish(ctx, "user.42.created", user)
```

### Wildcard Syntax

| Pattern | Meaning | MemoryBroker | Redis | NATS | AMQP | MQTT |
|---|---|---|---|---|---|---|
| `*` | Single token | `*` | `*` | `*` | `*` | `+` (auto-converted) |
| `#` | Zero or more tokens | `#` | `*` (auto-converted) | `>` (auto-converted) | `#` | `#` |
| `.` | Token separator | `.` | `.` | `.` | `.` | `/` (auto-converted) |

Write patterns using `*` and `#` with `.` separators. The broker
translates to the native wildcard syntax automatically.

### Advanced Configuration

```go
// NATS with full options
broker, err := event.NewNATSBrokerWith(&event.NATSBrokerOptions{
    URLs:          []string{"nats://host1:4222", "nats://host2:4222"},
    Name:          "my-service",
    MaxReconnects: -1,          // unlimited
    Username:      "user",
    Password:      "pass",
    TLSConfig:     tlsConfig,
})

// AMQP with custom exchange
broker, err := event.NewAMQPBrokerWith(&event.AMQPBrokerOptions{
    URL:      "amqp://guest:guest@localhost:5672/",
    Exchange: "my.events",
})

// From existing connections
broker := event.NewNATSBrokerFrom(existingNatsConn)
broker := event.NewRedisBrokerFrom(existingRedisClient)
broker, err := event.NewAMQPBrokerFrom(existingAmqpConn, "exchange")
broker := event.NewMQTTBrokerFrom(existingManager, qos)
```

### Broker Behavior Notes

- **MemoryBroker:** Handlers run in goroutines. `Publish` returns
  before handlers execute. Handler panics are silently recovered.
- **RedisBroker:** Uses Redis Pub/Sub. No message persistence — if no
  subscriber is listening, the message is lost.
- **NATSBroker:** `Close()` drains pending messages before closing.
- **AMQPBroker:** Each subscriber gets an exclusive auto-delete queue.
  Messages are auto-acknowledged.
- **MQTTBroker:** Uses clean sessions. Supports QoS 0, 1, 2.
  Default QoS is 1 (at least once). Auto-reconnection via autopaho.
