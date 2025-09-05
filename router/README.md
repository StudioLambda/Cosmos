# Cosmos: Router

A lightweight HTTP router designed to keep things simple using `http.ServeMux`. It brings structured routing and elegant middleware chaining to your HTTP apps.

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

	"github.com/studiolambda/cosmos/router"
)

func logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		slog.Info("incoming request request", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

func main() {
	r := router.New[http.Handler]()

	// Use a global middleware
	r.Use(logger)

	// Basic GET route
	r.Get("/hello", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello from cosmos router!")
	}))

	// Grouped routes
	r.Group("/api", func(api *router.Router[http.Handler]) {
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
go get github.com/studiolambda/cosmos/router
```

---

## Testing Helpers

Use `.Record(req)` to easily test route output:

```go
req := httptest.NewRequest("GET", "/hello", nil)
res := r.Record(req)
fmt.Println(res.Body.String()) // => "Hello from cosmos router!"
```

---

## License

MIT © 2025 Èrik C. Forés
