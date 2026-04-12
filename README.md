# Cosmos

Cosmos is a lightweight, powerful, and flexible collection of HTTP modules for Go designed to work seamlessly together while maintaining excellent compatibility with the standard library. The modules are intentionally designed as thin wrappers that enhance rather than replace Go's built-in HTTP capabilities.

## Philosophy

Cosmos follows these core principles:

- Stay close to the standard library
- Provide composable, independent modules
- Enable extensibility without complexity
- Maintain zero unnecessary dependencies
- Support both standalone and integrated usage

## Modules

This monorepo contains four independently publishable Go modules:

### router

A generic HTTP router built on top of `http.ServeMux` with middleware support and a familiar, chainable API. Supports path parameters, route groups, and automatic trailing slash handling.

**Module:** `github.com/studiolambda/cosmos/router`

**Key Features:**
- Built on Go's standard `http.ServeMux`
- Generic over any `http.Handler` implementation
- Hierarchical routers with middleware inheritance
- Automatic dual route registration for trailing slashes
- Testing helpers with response recording

### problem

A pure Go implementation of RFC 9457 (Problem Details for HTTP APIs) that works with any HTTP framework. Provides structured, machine-readable error responses with content negotiation support.

**Module:** `github.com/studiolambda/cosmos/problem`

**Key Features:**
- RFC 9457 compliant problem details
- Automatic content negotiation (JSON, Problem+JSON, text)
- Stack trace support for development
- Implements `http.Handler` for direct serving
- Works standalone or integrated with frameworks

### contract

A collection of common service interfaces (Cache, Database, Session, Crypto, Hash, Events, Hooks) with zero dependencies. Designed for dependency injection and testing with included mock implementations.

**Module:** `github.com/studiolambda/cosmos/contract`

**Key Features:**
- Zero dependencies for maximum portability
- Request helpers with typed integer parsing (`ParamInt`, `QueryInt`)
- Size-limited body parsing (`LimitedJSON`, `StrictJSON`)
- Safe redirect validation
- Response helpers (JSON, HTML, XML, SSE, streaming)
- Hooks system for middleware lifecycle events
- Mock implementations via mockery

### framework

A complete HTTP framework that orchestrates the router, problem, and contract modules. Provides error-returning handlers, middleware composition, and extensive utilities for building web applications.

**Module:** `github.com/studiolambda/cosmos/framework`

**Key Features:**
- Error-returning handler pattern
- First-class middleware system (CORS, CSRF, rate limiting, secure headers, recovery, logging)
- Secure HTTP server with timeout defaults
- Session management with cache backends and absolute lifetime enforcement
- Cryptography (AES-GCM, ChaCha20-Poly1305) with key zeroing and additional authenticated data
- Password hashing (Argon2, Bcrypt) with secure defaults
- SQL database wrapper with transactions and connection pool tuning
- Cache implementations (Memory, Redis)
- Event system (Memory, Redis, NATS, AMQP, MQTT) with wildcard subscriptions

## Installation

Each module can be installed independently:

```bash
# Framework includes router, problem, and contract as dependencies
go get github.com/studiolambda/cosmos/framework

# Or install individual modules
go get github.com/studiolambda/cosmos/router
go get github.com/studiolambda/cosmos/problem
go get github.com/studiolambda/cosmos/contract
```

Requires **Go 1.25** or later.

## Quick Start

### Using the Framework

```go
package main

import (
    "log/slog"
    "net/http"

    "github.com/studiolambda/cosmos/contract/response"
    "github.com/studiolambda/cosmos/framework"
    "github.com/studiolambda/cosmos/framework/middleware"
)

type User struct {
    ID   int    `json:"id"`
    Name string `json:"name"`
}

func getUser(w http.ResponseWriter, r *http.Request) error {
    user := User{ID: 1, Name: "Alice"}
    return response.JSON(w, http.StatusOK, user)
}

func main() {
    app := framework.New()

    app.Use(middleware.Recover())
    app.Use(middleware.Logger(slog.Default()))
    app.Use(middleware.SecureHeaders())

    app.Get("/users/{id}", getUser)

    server := framework.NewServer(":8080", app)
    server.ListenAndServe()
}
```

### Using Individual Modules

```go
// Just the router
import "github.com/studiolambda/cosmos/router"

r := router.New[http.Handler]()
r.Get("/hello", http.HandlerFunc(handler))
http.ListenAndServe(":8080", r)

// Just problem details
import "github.com/studiolambda/cosmos/problem"

var ErrNotFound = problem.Problem{
    Title:  "Resource Not Found",
    Detail: "The requested resource does not exist",
    Status: http.StatusNotFound,
}

http.HandleFunc("/api/users/{id}", func(w http.ResponseWriter, r *http.Request) {
    ErrNotFound.ServeHTTP(w, r)
})
```

## Security

Cosmos includes security hardening across all modules:

- **CORS** — Configurable cross-origin resource sharing middleware
- **CSRF** — Cross-site request forgery protection via Go's `http.CrossOriginProtection`
- **Rate Limiting** — Per-key token bucket rate limiting middleware
- **Secure Headers** — X-Content-Type-Options, X-Frame-Options, HSTS, Referrer-Policy, CSP
- **Server Timeouts** — Secure defaults preventing Slowloris and connection exhaustion
- **Session Security** — Cryptographically random IDs (256-bit), absolute lifetime enforcement, session ID validation, secure cookie defaults
- **Body Limits** — Size-limited body parsing to prevent OOM attacks
- **Safe Redirects** — URL validation preventing open redirect vulnerabilities
- **XSS Prevention** — `html/template` for HTML rendering with charset enforcement
- **Prepared Statement Cleanup** — Automatic `defer Close()` preventing connection pool exhaustion
- **Key Zeroing** — `Close()` methods on encryption types to clear key material from memory
- **Password Hashing** — Argon2 with password zeroing, Bcrypt with OWASP cost defaults

## Development

This project uses Go workspaces for local development:

```bash
# Run tests for all modules
go test ./...

# Run tests for specific module
go test ./framework/...
go test ./router/...

# Run tests with coverage
go test -cover ./...

# Format and vet
go fmt ./...
go vet ./...

# Generate mocks (contract only)
cd contract && go generate ./...
```

### Module Dependencies

```
contract (zero dependencies)
    ↓
problem (standalone)
router (standalone)
    ↓
framework (depends on: router, problem, contract)
```

## Architecture

### Error-Returning Handlers

The framework uses handlers that return errors instead of requiring manual error response writing:

```go
type Handler func(w http.ResponseWriter, r *http.Request) error
```

This pattern enables:
- Centralized error handling
- Cleaner handler code
- Consistent error formatting via Problem Details
- Custom status codes via `HTTPStatus() int` interface

### Middleware Composition

Middleware wraps handlers in layers:

```go
type Middleware = func(next Handler) Handler
```

Execution order: `A → B → C → handler → C → B → A`

### Hooks System

The hooks system allows middleware to inject behavior at key points in the request/response lifecycle:

- `BeforeWriteHeader` — Called before HTTP status is written
- `BeforeWrite` — Called before response body is written
- `AfterResponse` — Called after response completes (with error if any)

## Project Structure

```
cosmos/
├── contract/          # Service interfaces and helpers
│   ├── request/       # Request helper functions
│   ├── response/      # Response helper functions
│   └── mock/          # Generated mocks
├── framework/         # Complete HTTP framework
│   ├── cache/         # Cache implementations (Memory, Redis)
│   ├── crypto/        # Encryption (AES-GCM, ChaCha20-Poly1305)
│   ├── database/      # SQL database wrapper
│   ├── event/         # Event brokers (Memory, Redis, NATS, AMQP, MQTT)
│   ├── hash/          # Password hashing (Argon2, Bcrypt)
│   ├── middleware/     # HTTP middleware
│   └── session/       # Session management
├── problem/           # RFC 9457 implementation
│   └── internal/      # Internal utilities
├── router/            # HTTP router
└── .github/
    └── workflows/     # CI/CD workflows
```

## Testing

Each module has comprehensive tests. Run all tests from the workspace root:

```bash
go test ./...              # All modules
go test -v ./framework/... # Verbose output for one module
go test -race ./...        # Race detection
go test -cover ./...       # Coverage report
```

## Contributing

Contributions are welcome! When contributing:

1. Write tests for new functionality
2. Follow existing code style and patterns (see `.agents/skills/cosmos-go/`)
3. Run `go fmt` and `go vet` before committing
4. Ensure all tests pass
5. Update documentation as needed

## License

MIT License - Copyright (c) 2025 Erik C. Fores

See LICENSE file in each module for details.

## Links

- GitHub: https://github.com/studiolambda/cosmos
- Documentation: https://studiolambda.com/cosmos/getting-started
- RFC 9457: https://datatracker.ietf.org/doc/html/rfc9457
