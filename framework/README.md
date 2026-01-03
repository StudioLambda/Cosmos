# Cosmos: Framework

A lightweight HTTP framework designed to close to the standard library. In fact,
it is a thin wrapper on top of the cosmos router and the HTTP package of go.
It provides features to work with http requests and responses in a way that
feels more natural.

---

## Features

- Close to the standard library
- Based on the cosmos router (http.ServeMux)
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

	"github.com/studiolambda/cosmos/framework"
	"github.com/studiolambda/cosmos/framework/middleware"
	"github.com/studiolambda/cosmos/contract/request"
	"github.com/studiolambda/cosmos/contract/response"
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
	app := framework.New()

	app.Use(middleware.Logger(slog.Default()))
	app.Use(middleware.Recover())

	app.Get("/", handler)

	http.ListenAndServe(":8080", app)
}

```

---

## Install

```bash
go get github.com/studiolambda/cosmos/framework
```

## License

MIT © 2025 Èrik C. Forés
