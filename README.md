# Cosmos

Cosmos is a collection of composable HTTP modules for Go that stay close to the standard library.

## Modules

This workspace contains five independently publishable modules:

- `collection`: generic slice and map helpers.
- `router`: generic HTTP router built on `http.ServeMux`.
- `problem`: RFC 9457 problem details.
- `contract`: typed facades, service-driver interfaces, and request/response helpers.
- `framework`: error-returning HTTP application framework and integrations.

`collection`, `router`, and `problem` are foundational. `contract` depends on `collection` and `problem`; `framework` depends on `contract`, `router`, and `problem`.

Requires Go 1.27 or later.

## Installation

```bash
go get github.com/studiolambda/cosmos/framework
```

Modules can also be imported independently, for example:

```go
import (
	"github.com/studiolambda/cosmos/contract"
	"github.com/studiolambda/cosmos/framework"
	"github.com/studiolambda/cosmos/problem"
	"github.com/studiolambda/cosmos/router"
)
```

## Quick Start

```go
package main

import (
	"net/http"

	"github.com/studiolambda/cosmos/contract/response"
	"github.com/studiolambda/cosmos/framework"
	"github.com/studiolambda/cosmos/framework/middleware"
)

func main() {
	app := framework.New()
	app.Use(middleware.Recover())
	app.Use(middleware.SecureHeaders(middleware.DefaultSecureHeadersConfig))
	app.Get("/health", func(w http.ResponseWriter, r *http.Request) error {
		return response.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	server := framework.NewServer(framework.ServerConfig{}, app)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		panic(err)
	}
}
```

Use `framework.HTTP(handler)` to register a standard `http.Handler` as a Cosmos route. Use `middleware.HTTP(middleware)` to adapt standard `func(http.Handler) http.Handler` middleware.

## Development

Run commands from the workspace root:

```bash
go work sync
go test ./...
go test -race ./...
go test -race -cover ./...
go fmt ./...
go vet ./...
```

Generate contract mocks with `cd contract && go generate ./...`.

## Security

Use `framework.NewServer` for secure timeout defaults. For browser-facing applications, add `middleware.Recover`, `middleware.SecureHeaders`, CSRF protection where appropriate, and rate limiting for sensitive endpoints. Regenerate sessions after authentication or privilege changes, and use Argon2 for new password stores.

## Links

- GitHub: https://github.com/studiolambda/cosmos
- Documentation: https://studiolambda.com/cosmos/getting-started
- RFC 9457: https://datatracker.ietf.org/doc/html/rfc9457
