package atlas

import (
	"log/slog"
	"net/http"
	"time"
)

type Options struct {
	Addr              string
	ReadTimeout       time.Duration
	ReadHeaderTimeout time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
	MaxHeaderBytes    int
	ShutdownTimeout   time.Duration
	Logger            *slog.Logger
	RouterMiddlewares bool
	IsDev             bool
}

var DefaultOptions = Options{
	Addr:              ":http",
	ReadTimeout:       5 * time.Second,
	ReadHeaderTimeout: 2 * time.Second,
	WriteTimeout:      10 * time.Second,
	IdleTimeout:       120 * time.Second,
	MaxHeaderBytes:    http.DefaultMaxHeaderBytes,
	ShutdownTimeout:   5 * time.Second,
	Logger:            slog.Default(),
	RouterMiddlewares: true,
	IsDev:             false,
}
