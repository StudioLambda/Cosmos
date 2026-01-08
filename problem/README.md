# Cosmos: Problem

A pure Go implementation of RFC 9457 (Problem Details for HTTP APIs) that provides structured, machine-readable error responses. Works with any HTTP framework or standalone with the standard library.

## Overview

The problem module helps standardize API error responses following RFC 9457. It provides:

- **RFC 9457 Compliant**: Full implementation of Problem Details for HTTP APIs
- **Content Negotiation**: Automatic support for JSON, Problem+JSON, and plain text
- **Stack Traces**: Optional error stack traces for development
- **HTTP Handler**: Implements `http.Handler` for direct serving
- **Error Wrapping**: Compatible with Go's error interface
- **Extensible**: Add custom fields via additional metadata
- **Framework Agnostic**: Works standalone or with any framework

## Installation

```bash
go get github.com/studiolambda/cosmos/problem
```

This module has zero dependencies.

## Quick Start

```go
package main

import (
    "net/http"
    "github.com/studiolambda/cosmos/problem"
)

var (
    ErrNotFound = problem.Problem{
        Title:  "Resource Not Found",
        Detail: "The requested resource does not exist",
        Status: http.StatusNotFound,
    }
    
    ErrUnauthorized = problem.Problem{
        Title:  "Unauthorized",
        Detail: "Authentication is required to access this resource",
        Status: http.StatusUnauthorized,
    }
)

func handler(w http.ResponseWriter, r *http.Request) {
    // Serve problem directly
    ErrNotFound.ServeHTTP(w, r)
}

func main() {
    http.HandleFunc("/api/users/{id}", handler)
    http.ListenAndServe(":8080", nil)
}
```

## RFC 9457 Structure

The Problem type follows RFC 9457 specification:

```go
type Problem struct {
    Type     string // URI reference identifying the problem type
    Title    string // Short, human-readable summary
    Detail   string // Human-readable explanation specific to this occurrence
    Status   int    // HTTP status code
    Instance string // URI reference identifying this specific occurrence
}
```

### Default Values

When serving, missing fields are automatically filled:

- `Type`: Defaults to "about:blank"
- `Title`: Defaults to HTTP status text (e.g., "Not Found")
- `Status`: Defaults to 500 Internal Server Error
- `Instance`: Defaults to request URL
- `Detail`: If error is wrapped, defaults to error message

## Basic Usage

### Define Problem Variables

```go
var (
    ErrNotFound = problem.Problem{
        Title:  "Resource Not Found",
        Detail: "The requested resource does not exist",
        Status: http.StatusNotFound,
    }
    
    ErrBadRequest = problem.Problem{
        Title:  "Invalid Request",
        Detail: "The request contains invalid data",
        Status: http.StatusBadRequest,
    }
    
    ErrForbidden = problem.Problem{
        Title:  "Access Denied",
        Detail: "You do not have permission to access this resource",
        Status: http.StatusForbidden,
    }
)
```

### Serve Problems

```go
func handler(w http.ResponseWriter, r *http.Request) {
    // Direct serving
    ErrNotFound.ServeHTTP(w, r)
}
```

### Create from Error

```go
err := someOperation()
if err != nil {
    problem := problem.NewProblem(err, http.StatusInternalServerError)
    problem.ServeHTTP(w, r)
}
```

## Advanced Features

### Additional Metadata

Add custom fields to problems:

```go
ErrNotFound.
    With("resource_id", userID).
    With("resource_type", "user").
    ServeHTTP(w, r)
```

JSON output:
```json
{
    "type": "about:blank",
    "title": "Resource Not Found",
    "detail": "The requested resource does not exist",
    "status": 404,
    "instance": "/api/users/123",
    "resource_id": "123",
    "resource_type": "user"
}
```

### Error Wrapping

Wrap native Go errors:

```go
err := database.Query()
if err != nil {
    problem := ErrInternalError.WithError(err)
    problem.ServeHTTP(w, r)
}
```

### Stack Traces

Enable stack traces for development:

```go
func handler(w http.ResponseWriter, r *http.Request) {
    err := someOperation()
    if err != nil {
        // Add stack trace
        problem := ErrInternalError.
            WithError(err).
            WithStackTrace()
        
        problem.ServeHTTP(w, r)
    }
}
```

Development mode shortcut:
```go
problem.ServeHTTPDev(w, r) // Automatically adds stack trace
```

JSON output with stack trace:
```json
{
    "type": "about:blank",
    "title": "Internal Server Error",
    "detail": "database connection failed",
    "status": 500,
    "instance": "/api/users",
    "stack_trace": [
        "database connection failed",
        "connection timeout: context deadline exceeded"
    ]
}
```

### Remove Information

Remove additional metadata or error information:

```go
problem := ErrNotFound.
    With("debug_info", "...").
    Without("debug_info") // Remove before sending to production

problem = problem.WithoutError() // Remove wrapped error
problem = problem.WithoutStackTrace() // Remove stack trace
```

## Content Negotiation

The problem module automatically negotiates content type based on the `Accept` header:

### JSON (application/json)

```bash
curl -H "Accept: application/json" http://localhost:8080/api/users/123
```

Response:
```json
{
    "type": "about:blank",
    "title": "Resource Not Found",
    "detail": "The requested resource does not exist",
    "status": 404,
    "instance": "/api/users/123"
}
```

### Problem+JSON (application/problem+json)

```bash
curl -H "Accept: application/problem+json" http://localhost:8080/api/users/123
```

Same JSON output with `Content-Type: application/problem+json` header.

### Plain Text (text/plain or default)

```bash
curl http://localhost:8080/api/users/123
```

Response:
```
404 Resource Not Found

The requested resource does not exist
```

## Working with Errors

### Implement Error Interface

Problem implements Go's error interface:

```go
var err error = problem.Problem{
    Title:  "Something went wrong",
    Status: http.StatusInternalServerError,
}

fmt.Println(err.Error()) // "500 internal server error: something went wrong"
```

### Unwrap Errors

Access wrapped errors:

```go
problem := ErrInternalError.WithError(originalErr)

// Unwrap to get original error
if err := problem.Unwrap(); err != nil {
    // err is originalErr
}

// Get all errors in chain
errors := problem.Errors() // []error
```

### Custom Status Codes

Implement `HTTPStatus() int` interface for custom status codes:

```go
type CustomError struct {
    message string
}

func (e CustomError) Error() string {
    return e.message
}

func (e CustomError) HTTPStatus() int {
    return http.StatusTeapot // 418
}

// When wrapped in Problem, status code is preserved
err := CustomError{message: "I'm a teapot"}
problem := problem.NewProblem(err, http.StatusInternalServerError)
problem.HTTPStatus() // Returns 418
```

## JSON Serialization

Problems can be marshaled and unmarshaled:

```go
// Marshal to JSON
problem := ErrNotFound.With("resource_id", 123)
data, err := json.Marshal(problem)

// Unmarshal from JSON
var restored problem.Problem
err = json.Unmarshal(data, &restored)
```

## Integration Examples

### With Standard Library

```go
http.HandleFunc("/api/users/{id}", func(w http.ResponseWriter, r *http.Request) {
    id := r.PathValue("id")
    user, err := getUser(id)
    
    if err != nil {
        ErrNotFound.
            With("user_id", id).
            ServeHTTP(w, r)
        return
    }
    
    json.NewEncoder(w).Encode(user)
})
```

### With Cosmos Framework

```go
import "github.com/studiolambda/cosmos/framework"

func getUser(w http.ResponseWriter, r *http.Request) error {
    id := request.Param(r, "id")
    user, err := fetchUser(id)
    
    if err != nil {
        // Return problem as error
        return ErrNotFound.With("user_id", id)
    }
    
    return response.JSON(w, http.StatusOK, user)
}

app := framework.New()
app.Get("/users/{id}", getUser)
```

### Middleware Error Handler

```go
func errorHandler(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        defer func() {
            if err := recover(); err != nil {
                problem := problem.Problem{
                    Title:  "Internal Server Error",
                    Detail: fmt.Sprintf("panic: %v", err),
                    Status: http.StatusInternalServerError,
                }
                problem.WithStackTrace().ServeHTTP(w, r)
            }
        }()
        next.ServeHTTP(w, r)
    })
}
```

## Best Practices

1. **Define Problems as Variables**: Reuse problem definitions across handlers
2. **Use Appropriate Status Codes**: Match HTTP status with problem severity
3. **Add Context**: Use `With()` to add relevant metadata
4. **Stack Traces in Dev Only**: Only enable in development/staging
5. **Consistent Titles**: Use consistent titles for the same problem type
6. **Detailed Details**: Provide specific, actionable detail messages
7. **Custom Types**: Set Type URI for common problems to enable client handling

## Example: Complete API

```go
package main

import (
    "errors"
    "net/http"
    "github.com/studiolambda/cosmos/problem"
)

var (
    ErrNotFound = problem.Problem{
        Type:   "https://api.example.com/errors/not-found",
        Title:  "Resource Not Found",
        Status: http.StatusNotFound,
    }
    
    ErrValidation = problem.Problem{
        Type:   "https://api.example.com/errors/validation",
        Title:  "Validation Failed",
        Status: http.StatusBadRequest,
    }
)

func getUser(w http.ResponseWriter, r *http.Request) {
    id := r.PathValue("id")
    
    if id == "" {
        ErrValidation.
            With("field", "id").
            With("message", "User ID is required").
            ServeHTTP(w, r)
        return
    }
    
    user, err := fetchUser(id)
    if errors.Is(err, sql.ErrNoRows) {
        ErrNotFound.
            With("user_id", id).
            ServeHTTP(w, r)
        return
    }
    
    if err != nil {
        problem.NewProblem(err, http.StatusInternalServerError).
            WithStackTrace().
            ServeHTTP(w, r)
        return
    }
    
    // Success response
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(user)
}
```

## License

MIT License - Copyright (c) 2025 Erik C. Fores
