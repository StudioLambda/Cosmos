# Cosmos: Orbit

**Orbit** is a lightweight HTTP router designed to keep things simple using `http.ServeMux`. It brings structured routing and elegant middleware chaining to your HTTP apps.

---

## Features

- Fully generic over `http.Handler` implementations.
- Automatic dual route registration (`/`, `/{$}`, `/path`, `/path/{$}`, etc.).
- Hierarchical routers with middleware inheritance.
- Middleware support via simple generic functions.
- Built-in route existence and handler matching helpers.
- Friendly to both standard servers and testing workflows.

---

## Example Usage

```go
package main

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/studiolambda/cosmos/orbit"
)

func logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		slog.Info("incoming request request", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

func main() {
	router := orbit.New[http.Handler]()

	// Use a global middleware
	router.Use(logger)

	// Basic GET route
	router.Get("/hello", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello from orbit!")
	}))

	// Grouped routes
	router.Group("/api", func(api *orbit.Router[http.Handler]) {
		api.Use( /* another middleware to only apply on api */ )

		api.Get("/users", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, "List of users")
		}))

		api.With(logger).Post("/users", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, "Create a user")
		}))
	})

	// Serve on port 8080
	http.ListenAndServe(":8080", r)
}
```

---

## Install

```bash
go get github.com/studiolambda/cosmos/orbit
```

---

## Testing Helpers

Use `.Record(req)` to easily test route output:

```go
req := httptest.NewRequest("GET", "/hello", nil)
res := router.Record(req)
fmt.Println(res.Body.String()) // => "Hello from orbit!"
```

---

## License

MIT © 2025 Èrik C. Forés
