# Cosmos: Contract

Service interfaces and helper utilities for building HTTP applications in Go. This module provides a collection of common interfaces (Cache, Database, Session, Crypto, Hash) with zero dependencies, designed for dependency injection and testing.

## Overview

The contract module serves as the foundation for Cosmos, providing:

- **Service Interfaces**: Common abstractions for caching, databases, sessions, encryption, and hashing
- **Request Helpers**: Functions for working with HTTP requests (headers, cookies, params, query strings, body parsing)
- **Response Helpers**: Functions for writing HTTP responses (JSON, streams, static files)
- **Hooks System**: Lifecycle hooks for middleware to inject behavior
- **Mock Implementations**: Generated mocks for all interfaces via mockery

This module has zero dependencies (except testing libraries) and can be used standalone or with any framework.

## Installation

```bash
go get github.com/studiolambda/cosmos/contract
```

## Service Interfaces

### Cache

Generic cache interface inspired by Laravel's Cache Repository. Supports CRUD operations, atomic operations, and lazy-loading via Remember patterns.

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

**Standard Errors:**
- `ErrCacheKeyNotFound`: Key does not exist in cache
- `ErrCacheUnsupportedOperation`: Operation not supported by backend

### Database

Generic SQL database interface with support for queries, transactions, and named parameters.

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
    WithTransaction(ctx context.Context, fn func(tx Database) error) error
}
```

**Standard Errors:**
- `ErrDatabaseNoRows`: No rows found
- `ErrDatabaseNestedTransaction`: Nested transaction attempted

### Session

User session interface with data storage and lifecycle management.

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

### Crypto

Encryption interface for authenticated encryption.

```go
type Crypto interface {
    Encrypt(ctx context.Context, plaintext []byte) ([]byte, error)
    Decrypt(ctx context.Context, ciphertext []byte) ([]byte, error)
}
```

### Hash

Password hashing interface.

```go
type Hash interface {
    Hash(ctx context.Context, password []byte) ([]byte, error)
    Verify(ctx context.Context, password, hash []byte) error
}
```

## Request Helpers

### Headers

```go
// Get request header value
value := request.Header(r, "Authorization")

// Get all header values
values := request.Headers(r, "Accept-Language")
```

### Cookies

```go
// Get cookie value
value, err := request.Cookie(r, "session_id")

// Check if cookie exists
exists := request.HasCookie(r, "preferences")
```

### Path Parameters

```go
// Get path parameter from route pattern
// Example: /users/{id}
id := request.Param(r, "id")
```

### Query Strings

```go
// Get single query parameter
page := request.Query(r, "page")

// Get all values for a query parameter
tags := request.Queries(r, "tags")

// Check if query parameter exists
hasFilter := request.HasQuery(r, "filter")
```

### Body Parsing

```go
// Parse JSON body
var user User
if err := request.Body(r, &user); err != nil {
    return err
}
```

### Hooks

```go
// Access request hooks
hooks := request.Hooks(r)

// Add hook before status is written
hooks.BeforeWriteHeader(func(w http.ResponseWriter, status int) {
    log.Printf("Status: %d", status)
})

// Add hook after response completes
hooks.AfterResponse(func(err error) {
    log.Printf("Request completed with error: %v", err)
})
```

### Sessions

```go
// Get session from request (requires session middleware)
session := request.Session(r)

// Check if request has session
if request.HasSession(r) {
    session := request.Session(r)
    userID, ok := session.Get("user_id")
}
```

## Response Helpers

### JSON Responses

```go
// Send JSON response
type User struct {
    ID   int    `json:"id"`
    Name string `json:"name"`
}

user := User{ID: 1, Name: "Alice"}
return response.JSON(w, http.StatusOK, user)
```

### Stream Responses

```go
// Stream data with custom content type
reader := strings.NewReader("streaming data")
return response.Stream(w, http.StatusOK, "text/plain", reader)
```

### Static Files

```go
// Serve static file
return response.Static(w, r, "/path/to/file.pdf")
```

## Hooks System

The hooks system provides lifecycle events for middleware:

```go
type Hooks interface {
    BeforeWriteHeader(fn func(w http.ResponseWriter, status int))
    BeforeWrite(fn func(w http.ResponseWriter, data []byte))
    AfterResponse(fn func(err error))
}
```

**Hook Execution:**

1. `BeforeWriteHeader`: Called before HTTP status is written (can be used to inspect/modify status)
2. `BeforeWrite`: Called before response body is written (can be used to inspect/modify data)
3. `AfterResponse`: Called after response completes, even if an error occurred (useful for logging)

## Testing with Mocks

The contract module includes generated mocks for all interfaces:

```go
import "github.com/studiolambda/cosmos/contract/mock"

// Create mock cache
mockCache := &mock.Cache{}
mockCache.On("Get", mock.Anything, "key").Return("value", nil)

// Use in tests
service := NewService(mockCache)
result, err := service.FetchData()

// Verify expectations
mockCache.AssertExpectations(t)
```

### Generating Mocks

Mocks are generated using mockery:

```bash
cd contract/
go generate ./...
```

Configuration is in `.mockery.yml`.

## Usage Examples

### Using Cache Interface

```go
func fetchUserData(ctx context.Context, cache contract.Cache, userID string) (*User, error) {
    // Try to get from cache
    key := fmt.Sprintf("user:%s", userID)
    
    // Use Remember pattern for lazy-loading
    data, err := cache.Remember(ctx, key, 1*time.Hour, func() (any, error) {
        // Fetch from database if not cached
        return fetchUserFromDB(userID)
    })
    
    if err != nil {
        return nil, err
    }
    
    return data.(*User), nil
}
```

### Using Database Interface

```go
func createUser(ctx context.Context, db contract.Database, user *User) error {
    query := `INSERT INTO users (name, email) VALUES ($1, $2)`
    
    _, err := db.Exec(ctx, query, user.Name, user.Email)
    return err
}

func getUserByEmail(ctx context.Context, db contract.Database, email string) (*User, error) {
    query := `SELECT id, name, email FROM users WHERE email = $1`
    
    var user User
    if err := db.Find(ctx, query, &user, email); err != nil {
        return nil, err
    }
    
    return &user, nil
}
```

### Using Session Interface

```go
func login(w http.ResponseWriter, r *http.Request) error {
    session := request.Session(r)
    
    // Authenticate user
    user, err := authenticateUser(r)
    if err != nil {
        return err
    }
    
    // Store user ID in session
    session.Put("user_id", user.ID)
    
    // Regenerate session to prevent fixation
    if err := session.Regenerate(); err != nil {
        return err
    }
    
    return response.JSON(w, http.StatusOK, user)
}
```

## Design Philosophy

The contract module follows these principles:

1. **Zero Dependencies**: Can be used with any framework or standalone
2. **Interface-Based**: All services are interfaces for maximum flexibility
3. **Standard Patterns**: Inspired by proven patterns from Laravel and other frameworks
4. **Testing First**: Includes mocks for all interfaces out of the box
5. **Context-Aware**: All operations accept context.Context for cancellation and timeouts

## Error Handling

All errors returned from contract interfaces should be checked. Standard errors are defined as variables that can be compared using `errors.Is()`:

```go
value, err := cache.Get(ctx, "key")
if errors.Is(err, contract.ErrCacheKeyNotFound) {
    // Handle missing key
}
```

## License

MIT License - Copyright (c) 2025 Erik C. Fores
