package main

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/samber/do/v2"
	"github.com/studiolambda/cosmos/application/internal/bootstrap"
	"github.com/studiolambda/cosmos/application/internal/config"
)

//go:embed config.yml
var configuration []byte

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	i := do.New(config.Package, bootstrap.Package)
	do.ProvideNamedValue(i, "configuration", configuration)

	logger := do.MustInvoke[*slog.Logger](i)
	server := do.MustInvoke[*http.Server](i)

	wg := sync.WaitGroup{}

	wg.Go(func() {
		logger.InfoContext(ctx, "started http server", "addr", "http://"+server.Addr)

		if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			logger.ErrorContext(ctx, "failed to listen http server", "err", err)
		}
	})

	<-ctx.Done()

	fmt.Fprint(os.Stdout, "\r")
	logger.InfoContext(ctx, "shutting down...")

	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.ErrorContext(ctx, "failed to stop http server", "err", err)
	}

	logger.InfoContext(ctx, "shutdown complete")
}
