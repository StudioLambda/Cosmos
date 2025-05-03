# Cosmos: Atlas

**Atlas** is a lightweight HTTP lifecycle manager for Nova-based web applications.
It provides a clean interface to bootstrap, run, and gracefully shut down HTTP servers using customizable hooks and options.

---

## Features

- Plug-and-play Nova integration.
- Graceful shutdown via `SIGINT`/`SIGTERM`.
- Lifecycle hooks (`BeforeStart`, `AfterStart`, `BeforeShutdown`, `AfterShutdown`).
- Automatic recovery and error handling middleware.
- Thread-safe server management.

---

## Example Usage

```go
package main

import (
	"log/slog"
	"net/http"

	"github.com/studiolambda/cosmos/atlas"
	"github.com/studiolambda/cosmos/nova"
	"github.com/studiolambda/cosmos/nova/response"
)

type App struct {
	//
}

func (a *App) Register(router *nova.Router) error {
	router.Get("/", a.Index)

	return nil
}

func (a *App) Index(w http.ResponseWriter, r *http.Request) error {
	return response.JSON(w, http.StatusOK, map[string]any{
		"status": "working",
	})
}

func main() {
	app := atlas.New(&App{})

	if err := app.Start(atlas.DefaultOptions); err != nil {
		slog.Error("failed to start app", "err", err)
	}
}
```

---

## Install

```bash
go get github.com/studiolambda/cosmos/atlas
```

## License

MIT © 2025 Èrik C. Forés
