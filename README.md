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

A collection of common service interfaces (Cache, Database, Session, Crypto, Hash) with zero dependencies. Designed for dependency injection and testing with included mock implementations.

**Module:** `github.com/studiolambda/cosmos/contract`

**Key Features:**
- Zero dependencies for maximum portability
- Request/response helper functions
- Hooks system for middleware lifecycle events
- Mock implementations via mockery
- Shared across all Cosmos modules

### framework

A complete HTTP framework that orchestrates the router, problem, and contract modules. Provides error-returning handlers, middleware composition, and extensive utilities for building web applications.

**Module:** `github.com/studiolambda/cosmos/framework`

**Key Features:**
- Error-returning handler pattern
- First-class middleware system
- Session management with cache backends
- Cryptography (AES-GCM, ChaCha20-Poly1305)
- Password hashing (Argon2, Bcrypt)
- SQL database wrapper with transactions
- Cache implementations (Memory, Redis)
- CSRF protection middleware

## Installation

Each module can be installed independently:

```bash
# Install individual modules
go get github.com/studiolambda/cosmos/router
go get github.com/studiolambda/cosmos/problem
go get github.com/studiolambda/cosmos/contract
go get github.com/studiolambda/cosmos/framework

# Framework includes router, problem, and contract as dependencies
go get github.com/studiolambda/cosmos/framework
```

## Quick Start

### Using the Framework

```go
package main

import (
    "log/slog"
    "net/http"

    "github.com/studiolambda/cosmos/framework"
    "github.com/studiolambda/cosmos/framework/middleware"
    "github.com/studiolambda/cosmos/contract/response"
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

    app.Use(middleware.Logger(slog.Default()))
    app.Use(middleware.Recover())

    app.Get("/users/{id}", getUser)

    http.ListenAndServe(":8080", app)
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

## Development

This project uses Go workspaces for local development:

```bash
# Initialize workspace (already configured)
go work init
go work use ./contract ./framework ./problem ./router

# Run tests for all modules
go test ./...

# Run tests for specific module
go test ./framework/...
go test ./router/...

# Run tests with coverage
go test -cover ./...

# Format all code
go fmt ./...

# Vet all code
go vet ./...
```

### Module Dependencies

The dependency hierarchy is:

```
contract (zero dependencies)
    ↓
problem (standalone)
router (standalone)
    ↓
framework (depends on: router, problem, contract)
```

### Working with Local Changes

When developing locally, the `go.work` file already contains replace directives:

```go
replace (
    github.com/studiolambda/cosmos/contract v0.6.1 => ./contract
    github.com/studiolambda/cosmos/framework v0.6.0 => ./framework
    github.com/studiolambda/cosmos/problem v0.2.0 => ./problem
    github.com/studiolambda/cosmos/router v0.2.0 => ./router
)
```

This allows you to make changes across modules and test them together before publishing.

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

- `BeforeWriteHeader`: Called before HTTP status is written
- `BeforeWrite`: Called before response body is written
- `AfterResponse`: Called after response completes (with error if any)

## Project Structure

```
cosmos/
├── contract/          # Service interfaces and helpers
│   ├── request/       # Request helper functions
│   ├── response/      # Response helper functions
│   └── mock/          # Generated mocks
├── framework/         # Complete HTTP framework
│   ├── cache/         # Cache implementations
│   ├── crypto/        # Encryption implementations
│   ├── database/      # Database wrapper
│   ├── hash/          # Password hashing
│   ├── middleware/    # HTTP middleware
│   └── session/       # Session management
├── problem/           # RFC 9457 implementation
│   └── internal/      # Internal utilities
├── router/            # HTTP router
└── .github/
    └── workflows/     # CI/CD workflows
```

## Testing

Each module has comprehensive tests:

```bash
# Run all tests
go test ./...

# Test specific module with verbose output
go test -v ./framework/...

# Run specific test
go test -run TestHandlerName ./framework

# Test with race detection
go test -race ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Contributing

Contributions are welcome! When contributing:

1. Write tests for new functionality
2. Follow existing code style and patterns
3. Run `go fmt` and `go vet` before committing
4. Ensure all tests pass
5. Update documentation as needed

## License

MIT License - Copyright (c) 2025 Erik C. Fores

See LICENSE file in each module for details.

## Links

- GitHub: https://github.com/studiolambda/cosmos
- Documentation: See individual module README files
- RFC 9457: https://datatracker.ietf.org/doc/html/rfc9457
