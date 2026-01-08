# Cosmos: Framework

A lightweight HTTP framework for Go that extends the standard library with error-returning handlers, middleware composition, and comprehensive utilities for building web applications. Designed to feel natural to Go developers while providing modern conveniences.

## Overview

The framework module provides:

- **Error-Returning Handlers**: Handlers that return errors for centralized error handling
- **Middleware System**: Composable middleware with hooks for lifecycle events
- **Session Management**: Thread-safe sessions with cache-backed storage
- **Cryptography**: AES-GCM and ChaCha20-Poly1305 encryption
- **Password Hashing**: Argon2 and Bcrypt implementations
- **Database Wrapper**: SQL database interface with transaction support
- **Cache Implementations**: Memory (go-cache) and Redis backends
- **Event Broker**: Publish/subscribe messaging with Memory, Redis, RabbitMQ, MQTT, and NATS backends
- **Response Utilities**: JSON, streaming, and static file helpers
- **Security Middleware**: CSRF protection, panic recovery, logging

## Installation

```bash
go get github.com/studiolambda/cosmos/framework
```

This will also install the required dependencies:
- `github.com/studiolambda/cosmos/router`
- `github.com/studiolambda/cosmos/problem`
- `github.com/studiolambda/cosmos/contract`

## Quick Start

```go
package main

import (
    "log/slog"
    "net/http"

    "github.com/studiolambda/cosmos/framework"
    "github.com/studiolambda/cosmos/framework/middleware"
    "github.com/studiolambda/cosmos/contract/request"
    "github.com/studiolambda/cosmos/contract/response"
)

type User struct {
    ID   int    `json:"id"`
    Name string `json:"name"`
}

func getUser(w http.ResponseWriter, r *http.Request) error {
    id := request.Param(r, "id")
    user := User{ID: 1, Name: "Alice"}
    return response.JSON(w, http.StatusOK, user)
}

func main() {
    app := framework.New()

    // Add middleware
    app.Use(middleware.Logger(slog.Default()))
    app.Use(middleware.Recover())

    // Define routes
    app.Get("/users/{id}", getUser)

    // Start server
    http.ListenAndServe(":8080", app)
}
```

## Core Concepts

### Error-Returning Handlers

Unlike standard `http.HandlerFunc`, framework handlers return errors:

```go
type Handler func(w http.ResponseWriter, r *http.Request) error
```

This enables centralized error handling:

```go
func handler(w http.ResponseWriter, r *http.Request) error {
    user, err := fetchUser()
    if err != nil {
        return err // Error handled by framework
    }
    return response.JSON(w, http.StatusOK, user)
}
```

Errors implementing the `HTTPStatus() int` interface can specify custom status codes.

### Middleware Composition

Middleware wraps handlers in layers:

```go
func MyMiddleware() framework.Middleware {
    return func(next framework.Handler) framework.Handler {
        return func(w http.ResponseWriter, r *http.Request) error {
            // Before handler execution
            
            err := next(w, r) // Call next handler
            
            // After handler execution
            
            return err // Propagate error
        }
    }
}

app.Use(MyMiddleware())
```

Execution order: `A → B → C → handler → C → B → A`

### Hooks System

Hooks allow middleware to inject behavior at key lifecycle points:

```go
hooks := request.Hooks(r)

// Capture status code
hooks.BeforeWriteHeader(func(w http.ResponseWriter, status int) {
    log.Printf("Status: %d", status)
})

// Log after completion
hooks.AfterResponse(func(err error) {
    log.Printf("Completed with error: %v", err)
})
```

## Middleware

### Logger

Logs HTTP requests with structured logging:

```go
import "log/slog"

logger := slog.Default()
app.Use(middleware.Logger(logger))
```

### Recover

Recovers from panics and returns 500 Internal Server Error:

```go
app.Use(middleware.Recover())
```

### CSRF Protection

Validates Origin and Sec-Fetch-Site headers:

```go
app.Use(middleware.CSRF("https://example.com", "https://app.example.com"))
```

Returns 403 Forbidden if request is from untrusted origin.

### HTTP Adapter

Adapts standard `http.Handler` to framework handlers:

```go
standardHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("Hello"))
})

app.Get("/hello", middleware.HTTP(standardHandler))
```

### Provide

Injects dependencies into request context:

```go
db := database.NewSQL(...)
app.Use(middleware.Provide(dbKey, db))

// Access in handler
db := r.Context().Value(dbKey).(contract.Database)
```

## Session Management

Sessions are thread-safe and track changes automatically:

```go
import (
    "github.com/studiolambda/cosmos/framework/session"
    "github.com/studiolambda/cosmos/framework/cache"
)

// Create cache-backed session driver
cacheImpl := cache.NewMemory(5*time.Minute, 10*time.Minute)
driver := session.NewCache(cacheImpl, 24*time.Hour)

// Add session middleware
app.Use(session.Middleware(driver, "session_id"))

// Use in handlers
func handler(w http.ResponseWriter, r *http.Request) error {
    sess := request.Session(r)
    
    // Store data
    sess.Put("user_id", 123)
    
    // Retrieve data
    if val, ok := sess.Get("user_id"); ok {
        userID := val.(int)
    }
    
    // Regenerate after authentication
    sess.Regenerate()
    
    return response.JSON(w, http.StatusOK, data)
}
```

## Caching

### Memory Cache

In-memory cache using go-cache:

```go
import "github.com/studiolambda/cosmos/framework/cache"

cache := cache.NewMemory(
    5*time.Minute,  // default expiration
    10*time.Minute, // cleanup interval
)

// Basic operations
cache.Put(ctx, "key", "value", 1*time.Hour)
value, err := cache.Get(ctx, "key")
cache.Delete(ctx, "key")

// Remember pattern (lazy-loading)
value, err := cache.Remember(ctx, "expensive", 1*time.Hour, func() (any, error) {
    return expensiveComputation(), nil
})
```

### Redis Cache

Redis-backed cache:

```go
import (
    "github.com/studiolambda/cosmos/framework/cache"
    "github.com/redis/go-redis/v9"
)

client := redis.NewClient(&redis.Options{
    Addr: "localhost:6379",
})

cache := cache.NewRedis(client)

// Same interface as memory cache
cache.Put(ctx, "key", "value", 1*time.Hour)
```

## Event Broker

Event brokers provide publish/subscribe messaging for decoupled communication between application components.

### Memory Broker

Pure in-memory pub/sub with zero dependencies, ideal for testing and local development:

```go
import "github.com/studiolambda/cosmos/framework/event"

// Create broker (no configuration needed!)
broker := event.NewMemoryBroker()
defer broker.Close()

// Publish events
err := broker.Publish(ctx, "user.created", map[string]any{
    "id":    123,
    "email": "user@example.com",
})

// Subscribe to events
type UserCreated struct {
    ID    int    `json:"id"`
    Email string `json:"email"`
}

unsubscribe, err := broker.Subscribe(ctx, "user.created", func(payload contract.EventPayload) {
    var event UserCreated
    if err := payload(&event); err != nil {
        log.Printf("Failed to unmarshal: %v", err)
        return
    }
    log.Printf("User created: %d - %s", event.ID, event.Email)
})
defer unsubscribe()

// Subscribe with wildcards
broker.Subscribe(ctx, "user.*", handler)       // Matches: user.created, user.deleted
broker.Subscribe(ctx, "logs.#", handler)       // Matches: logs, logs.error, logs.info.debug
```

**Features**:
- Zero dependencies (stdlib only)
- Zero configuration required
- Thread-safe concurrent access
- Async message delivery with panic recovery
- Native fan-out (all subscribers receive messages)
- Wildcard subscriptions (`*` single-level, `#` multi-level)
- Perfect for unit tests and local development

**Note**: No persistence - messages exist only in memory and are lost on restart.

### Redis Broker

Redis-backed pub/sub messaging:

```go
import (
    "github.com/studiolambda/cosmos/framework/event"
    "github.com/redis/go-redis/v9"
)

// Create Redis client
client := redis.NewClient(&redis.Options{
    Addr: "localhost:6379",
})

// Create broker
broker := event.NewRedisBrokerFrom(client)
defer broker.Close()

// Publish events
err := broker.Publish(ctx, "user.created", map[string]any{
    "id":    123,
    "email": "user@example.com",
})

// Subscribe to events
type UserCreated struct {
    ID    int    `json:"id"`
    Email string `json:"email"`
}

unsubscribe := broker.Subscribe(ctx, "user.created", func(payload contract.EventPayload) {
    var event UserCreated
    if err := payload(&event); err != nil {
        log.Printf("Failed to unmarshal event: %v", err)
        return
    }
    log.Printf("User created: %d - %s", event.ID, event.Email)
})

// Unsubscribe when done
defer unsubscribe()
```

### AMQP/RabbitMQ Broker

RabbitMQ-backed pub/sub messaging with topic exchange:

```go
import "github.com/studiolambda/cosmos/framework/event"

// Create broker with default exchange name "cosmos.events"
broker, err := event.NewAMQPBroker("amqp://guest:guest@localhost:5672/")
if err != nil {
    log.Fatal(err)
}
defer broker.Close()

// Or with custom options
broker, err := event.NewAMQPBrokerWithOptions(&event.AMQPBrokerOptions{
    URL:      "amqp://guest:guest@localhost:5672/",
    Exchange: "my-events",
})

// Publish events (broadcasted to all subscribers)
err := broker.Publish(ctx, "order.placed", map[string]any{
    "order_id": "123",
    "amount":   99.99,
})

// Subscribe to specific events
type OrderPlaced struct {
    OrderID string  `json:"order_id"`
    Amount  float64 `json:"amount"`
}

unsubscribe := broker.Subscribe(ctx, "order.placed", func(payload contract.EventPayload) {
    var event OrderPlaced
    if err := payload(&event); err != nil {
        log.Printf("Failed to unmarshal event: %v", err)
        return
    }
    log.Printf("Order placed: %s - $%.2f", event.OrderID, event.Amount)
})

defer unsubscribe()
```

### MQTT Broker

MQTT v5 pub/sub messaging with automatic reconnection and clean sessions:

```go
import "github.com/studiolambda/cosmos/framework/event"

// Simple connection with defaults (QoS 1, clean sessions)
broker, err := event.NewMQTTBroker("mqtt://localhost:1883")
if err != nil {
    log.Fatal(err)
}
defer broker.Close()

// Or with custom options
broker, err := event.NewMQTTBrokerWith(&event.MQTTBrokerOptions{
    URLs: []string{
        "mqtt://broker1:1883",
        "mqtt://broker2:1883", // Failover support
    },
    QoS:      1,
    Username: "user",
    Password: "pass",
})

// Publish events (topics auto-converted: . → /, * → +)
err := broker.Publish(ctx, "sensor.temperature", map[string]any{
    "device_id": "sensor1",
    "value":     23.5,
})

// Subscribe to specific topics
type TempReading struct {
    DeviceID string  `json:"device_id"`
    Value    float64 `json:"value"`
}

unsubscribe, err := broker.Subscribe(ctx, "sensor.temperature", func(payload contract.EventPayload) {
    var reading TempReading
    if err := payload(&reading); err != nil {
        log.Printf("Error: %v", err)
        return
    }
    log.Printf("Temperature: %.1f", reading.Value)
})
defer unsubscribe()

// Subscribe with wildcards (auto-converted: * → +)
broker.Subscribe(ctx, "sensor.*.status", handler) // Becomes: sensor/+/status
broker.Subscribe(ctx, "device.#", handler)        // Becomes: device/#
```

**Topic Conversion**: Event names are automatically converted to MQTT format:
- `.` (dot) → `/` (MQTT topic separator)
- `*` (asterisk) → `+` (MQTT single-level wildcard)
- `#` (hash) → `#` (MQTT multi-level wildcard, unchanged)

Examples:
- `user.created` → `user/created`
- `user.*.events` → `user/+/events`
- `logs.#` → `logs/#`

### NATS Broker

High-performance, lightweight pub/sub messaging with NATS:

```go
import "github.com/studiolambda/cosmos/framework/event"

// Simple connection with defaults
broker, err := event.NewNATSBroker("nats://localhost:4222")
if err != nil {
    log.Fatal(err)
}
defer broker.Close()

// Or with custom options (clustering, auth, TLS)
broker, err := event.NewNATSBrokerWith(&event.NATSBrokerOptions{
    URLs: []string{
        "nats://server1:4222",
        "nats://server2:4222", // Automatic failover
    },
    Name:          "myapp-service",
    MaxReconnects: -1,                    // Unlimited reconnection
    ReconnectWait: 2 * time.Second,
    Username:      "user",
    Password:      "pass",
})

// Or from existing connection
conn, _ := nats.Connect("nats://localhost:4222")
broker := event.NewNATSBrokerFrom(conn)

// Publish events (subject uses dot notation)
err := broker.Publish(ctx, "order.placed", map[string]any{
    "order_id": "12345",
    "total":    149.99,
})

// Subscribe to specific subjects
type OrderPlaced struct {
    OrderID string  `json:"order_id"`
    Total   float64 `json:"total"`
}

unsubscribe, err := broker.Subscribe(ctx, "order.placed", func(payload contract.EventPayload) {
    var order OrderPlaced
    if err := payload(&order); err != nil {
        log.Printf("Error: %v", err)
        return
    }
    log.Printf("Order %s placed: $%.2f", order.OrderID, order.Total)
})
defer unsubscribe()

// Subscribe with wildcards (auto-converted: # → >)
broker.Subscribe(ctx, "order.*", handler)      // Matches: order.placed, order.shipped
broker.Subscribe(ctx, "logs.#", handler)       // Matches: logs.error, logs.info.debug
```

**Subject Conversion**: Event patterns are automatically converted to NATS format:
- `.` (dot) - NATS subject separator (no conversion needed)
- `*` (asterisk) - Matches single token (no conversion needed)
- `#` (hash) → `>` (NATS multi-level wildcard)

Examples:
- `user.created` → `user.created`
- `user.*.events` → `user.*.events`
- `logs.#` → `logs.>`

**Authentication Options**:
- Username/password
- Token-based authentication
- NKey cryptographic authentication
- JWT credentials file (recommended for production)
- TLS with client certificates

**Features**:
- Automatic reconnection with configurable backoff
- Cluster support with automatic failover
- Native fan-out (all subscribers receive messages)
- Graceful shutdown with message draining
- Lightweight and high-performance

### Event Broker Patterns

**Multiple Subscribers**: All subscribers to an event receive their own copy:

```go
// Start multiple workers processing the same events
for i := 0; i < 3; i++ {
    broker.Subscribe(ctx, "task.created", func(payload contract.EventPayload) {
        var task Task
        payload(&task)
        processTask(task)
    })
}
```

**Event Fan-out**: Publish once, multiple handlers react:

```go
// Email notification handler
broker.Subscribe(ctx, "user.registered", sendWelcomeEmail)

// Analytics handler
broker.Subscribe(ctx, "user.registered", trackUserRegistration)

// Provisioning handler
broker.Subscribe(ctx, "user.registered", createUserResources)

// Single publish reaches all handlers
broker.Publish(ctx, "user.registered", userData)
```

**Context Cancellation**: Subscriptions respect context cancellation:

```go
ctx, cancel := context.WithTimeout(context.Background(), 1*time.Hour)
defer cancel()

unsubscribe := broker.Subscribe(ctx, "events", handler)

// Subscription automatically stops when context is cancelled or times out
```

## Cryptography

### AES-GCM Encryption

```go
import "github.com/studiolambda/cosmos/framework/crypto"

// Key must be 16, 24, or 32 bytes (AES-128/192/256)
key := []byte("your-32-byte-key-here-padding!!")
aes := crypto.NewAES(key)

// Encrypt
plaintext := []byte("secret message")
ciphertext, err := aes.Encrypt(ctx, plaintext)

// Decrypt
plaintext, err := aes.Decrypt(ctx, ciphertext)
```

### ChaCha20-Poly1305 Encryption

```go
import "github.com/studiolambda/cosmos/framework/crypto"

// Key must be 32 bytes
key := []byte("your-32-byte-key-here-padding!!")
chacha := crypto.NewChaCha20(key)

ciphertext, err := chacha.Encrypt(ctx, plaintext)
plaintext, err := chacha.Decrypt(ctx, ciphertext)
```

## Password Hashing

### Argon2

Recommended for new applications (memory-hard):

```go
import "github.com/studiolambda/cosmos/framework/hash"

hasher := hash.NewArgon2()

// Hash password
password := []byte("user-password")
hashed, err := hasher.Hash(ctx, password)

// Verify password
err := hasher.Verify(ctx, password, hashed)
if err != nil {
    // Password incorrect
}
```

### Bcrypt

Compatible with existing systems:

```go
import "github.com/studiolambda/cosmos/framework/hash"

hasher := hash.NewBcrypt(10) // cost factor

hashed, err := hasher.Hash(ctx, password)
err := hasher.Verify(ctx, password, hashed)
```

## Database

SQL database wrapper built on sqlx:

```go
import "github.com/studiolambda/cosmos/framework/database"

db := database.NewSQL("postgres", "connection-string")

// Single row query
var user User
err := db.Find(ctx, "SELECT * FROM users WHERE id = $1", &user, userID)

// Multiple rows
var users []User
err := db.Select(ctx, "SELECT * FROM users", &users)

// Execute statement
affected, err := db.Exec(ctx, "DELETE FROM users WHERE id = $1", userID)

// Named parameters
query := "INSERT INTO users (name, email) VALUES (:name, :email)"
arg := map[string]any{"name": "Alice", "email": "alice@example.com"}
affected, err := db.ExecNamed(ctx, query, arg)

// Transactions
err := db.WithTransaction(ctx, func(tx contract.Database) error {
    _, err := tx.Exec(ctx, "INSERT INTO ...", args...)
    if err != nil {
        return err // Automatic rollback
    }
    return nil // Automatic commit
})
```

## Routing

The framework uses the Cosmos router with full support for:

```go
// HTTP methods
app.Get("/path", handler)
app.Post("/path", handler)
app.Put("/path", handler)
app.Delete("/path", handler)
app.Patch("/path", handler)

// Path parameters
app.Get("/users/{id}", handler)
app.Get("/posts/{id}/comments/{commentID}", handler)

// Wildcards
app.Get("/static/{path...}", handler)

// Route groups
app.Group("/api", func(api *framework.Router) {
    api.Use(authMiddleware)
    api.Get("/users", listUsers)
    api.Post("/users", createUser)
})

// Conditional middleware
app.With(middleware.CSRF()).Post("/users", createUser)
```

## Response Helpers

### JSON Responses

```go
import "github.com/studiolambda/cosmos/contract/response"

type User struct {
    ID   int    `json:"id"`
    Name string `json:"name"`
}

user := User{ID: 1, Name: "Alice"}
return response.JSON(w, http.StatusOK, user)
```

### Streaming Responses

```go
reader := strings.NewReader("streaming data")
return response.Stream(w, http.StatusOK, "text/plain", reader)
```

### Static Files

```go
return response.Static(w, r, "/path/to/file.pdf")
```

## Request Helpers

```go
import "github.com/studiolambda/cosmos/contract/request"

// Path parameters
id := request.Param(r, "id")

// Query strings
page := request.Query(r, "page")
tags := request.Queries(r, "tag") // multiple values

// Headers
auth := request.Header(r, "Authorization")

// Cookies
sessionID, err := request.Cookie(r, "session_id")

// JSON body
var data RequestData
err := request.Body(r, &data)
```

## Testing

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific package tests
go test ./middleware/...

# Run specific test
go test -run TestMiddlewareName ./middleware/
```

## Security Considerations

1. **CSRF Protection**: Use `middleware.CSRF()` for state-changing endpoints
2. **Password Hashing**: Use Argon2 for new applications
3. **Encryption**: Use AES-GCM or ChaCha20-Poly1305 (authenticated encryption)
4. **Session Security**: Regenerate session after authentication
5. **Panic Recovery**: Use `middleware.Recover()` to prevent crashes

## Performance Tips

1. **Cache Strategy**: Use Remember pattern for expensive operations
2. **Database**: Close connections with `defer db.Close()`
3. **Middleware Order**: Place cheap middleware first
4. **Sessions**: Set appropriate expiration times

## License

MIT License - Copyright (c) 2025 Erik C. Fores
