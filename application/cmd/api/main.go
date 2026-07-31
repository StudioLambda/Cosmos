package main

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/studiolambda/cosmos/application/internal/bootstrap"
	"github.com/studiolambda/cosmos/framework/configuration"
)

//go:embed config/*.yml
var configurationFS embed.FS

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if err := bootstrap.LoadEnv(".env"); err != nil {
		return fmt.Errorf("load .env: %w", err)
	}

	configuration, err := bootstrap.NewConfig(
		configuration.Filesystem(configurationFS),
		configuration.Environment("COSMOS"),
	)
	if err != nil {
		return err
	}

	logger, err := bootstrap.NewLogger(configuration)
	if err != nil {
		return err
	}

	database, err := bootstrap.NewDatabase(configuration)
	if err != nil {
		return err
	}
	defer database.Close()

	events := bootstrap.NewEvents(configuration)
	defer events.Close()

	crypto, err := bootstrap.NewCrypto(configuration)
	if err != nil {
		return err
	}
	defer crypto.Close()

	_ = bootstrap.NewHasher(configuration)

	router := bootstrap.NewHTTPRouter(configuration, logger)
	server := bootstrap.NewHTTPServer(configuration, router)

	serverErrors := make(chan error, 1)

	go func() {
		logger.InfoContext(ctx, "started http server", "addr", "http://"+server.Addr)

		if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			serverErrors <- err
		}
	}()

	select {
	case err := <-serverErrors:
		return fmt.Errorf("listen http server: %w", err)
	case <-ctx.Done():
	}

	fmt.Fprint(os.Stdout, "\r")
	logger.InfoContext(ctx, "shutting down...")

	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.ErrorContext(ctx, "failed to stop http server", "err", err)
	}

	logger.InfoContext(ctx, "shutdown complete")

	return nil
}
