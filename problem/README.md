# Cosmos: Problem

**Problem** provides an easy way to work with [RFC 9457 - Problem Details for HTTP APIs](https://datatracker.ietf.org/doc/html/rfc9457) in Go.
It helps standardize your error responses while keeping them flexible and developer-friendly.

---

## Features

- Define problems following the Problem Details specification (RFC 9457)
- Attach additional metadata to problems
- Track and attach stack traces automatically
- Customize JSON serialization and deserialization
- Serve problems directly as HTTP handlers
- Developer-friendly mode that includes stack traces in responses

---

## Example Usage

```go
package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/studiolambda/cosmos/fracture"
)

var (
	ErrUnauthenticated = fracture.Problem{
		Title:  "Unauthenticated",
		Detail: "You need to login first to access this resource",
		Status: http.StatusUnauthorized,
	}

	ErrForbidden = fracture.Problem{
		Title:  "Forbidden",
		Detail: "You dont have permissions to perform this action",
		Status: http.StatusForbidden,
	}
)

func unauthenticated(w http.ResponseWriter, r *http.Request) {
	ErrUnauthenticated.ServeHTTP(w, r)
}

func forbidden(w http.ResponseWriter, r *http.Request) {
	err := errors.New("some permission lacking")

	ErrForbidden.WithError(err).ServeHTTP(w, r)
}

func withStackTrace(w http.ResponseWriter, r *http.Request) {
	err := errors.Join(
		errors.New("first error"),
		fmt.Errorf("%w: some additional info", errors.New("second error")),
		errors.New("third error"),
	)

	ErrForbidden.WithError(err).WithStackTrace().ServeHTTP(w, r)
}

func main() {
	router := http.NewServeMux()

	router.HandleFunc("GET /unauthenticated", unauthenticated)
	router.HandleFunc("GET /forbidden", forbidden)
	router.HandleFunc("GET /stacktrace", withStackTrace)

	http.ListenAndServe(":8080", router)
}
```

---

## Install

```bash
go get github.com/studiolambda/cosmos/fracture
```

## License

MIT © 2025 Èrik C. Forés
