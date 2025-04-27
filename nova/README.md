# Cosmos: Nova

**Nova** is a lightweight HTTP framework designed to close to the standard library. In fact,
it is a thin wrapper on top of the Orbit router and the HTTP package of go. Nova provides
features to work with http requests and responses in a way to feels more natural.

---

## Features

- Close to the standard library
- Based on the Orbit router (http.ServeMux)
- Provides first-class middlewares such as panic recovery or logging.
- It is designed to be extensible with other types as return values.
- Works well with the Fracture Problems's API RFC.
- Provides convenient helpers to work with requests and responses.

---

## Example Usage

```go
package main

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/studiolambda/cosmos/nova"
	"github.com/studiolambda/cosmos/nova/middleware"
	"github.com/studiolambda/cosmos/nova/request"
	"github.com/studiolambda/cosmos/nova/response"
)

type Data struct {
	Message string `json:"message"`
}

var (
	ErrNotEnabled = errors.New("feature not enabled")
)

func handler(w http.ResponseWriter, r *http.Request) error {
	if request.Query(r, "enabled") != "true" {
		return ErrNotEnabled
	}

	return response.JSON(w, http.StatusOK, Data{
		Message: "seems its working",
	})
}

func main() {
	router := nova.New()

	router.Use(middleware.ErrorHandler(middleware.ErrorHandlerOptions{
		Logger: slog.Default(),
		IsDev:  true,
	}))

	router.Use(middleware.Recover())

	router.Get("/", handler)

	http.ListenAndServe(":8080", router)
}

```

---

## Install

```bash
go get github.com/studiolambda/cosmos/nova
```

## License

MIT © 2025 Èrik C. Forés
