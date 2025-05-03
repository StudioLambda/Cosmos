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
