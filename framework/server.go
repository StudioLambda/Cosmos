package framework

import (
	"net/http"
	"time"
)

// ServerOptions configures the HTTP server created by [NewServer].
// All timeout fields default to secure values when zero.
type ServerOptions struct {
	// Addr is the TCP address to listen on (e.g. ":8080").
	Addr string

	// ReadHeaderTimeout limits the time allowed to read request
	// headers. Protects against Slowloris attacks.
	// Defaults to 10s.
	ReadHeaderTimeout time.Duration

	// ReadTimeout limits the total time for reading the entire
	// request including the body. Defaults to 30s.
	ReadTimeout time.Duration

	// WriteTimeout limits the time for writing the response.
	// Defaults to 60s.
	WriteTimeout time.Duration

	// IdleTimeout limits the keep-alive idle time between
	// requests. Defaults to 120s.
	IdleTimeout time.Duration

	// MaxHeaderBytes limits the maximum size of request headers.
	// Defaults to 1 MB (Go's default).
	MaxHeaderBytes int
}

// DefaultServerOptions returns secure timeout defaults suitable
// for production deployments. These protect against Slowloris
// and other connection-exhaustion attacks.
func DefaultServerOptions() ServerOptions {
	return ServerOptions{
		Addr:              ":8080",
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       120 * time.Second,
		MaxHeaderBytes:    1 << 20, // 1 MB
	}
}

// NewServer creates an [http.Server] with secure timeout defaults
// using the given handler. It applies [DefaultServerOptions]
// values, protecting against Slowloris and connection-exhaustion
// attacks that are possible when using [http.ListenAndServe]
// directly (which sets all timeouts to zero/infinite).
func NewServer(addr string, handler http.Handler) *http.Server {
	opts := DefaultServerOptions()
	opts.Addr = addr

	return NewServerWith(opts, handler)
}

// NewServerWith creates an [http.Server] with the provided
// options and handler. Zero-valued timeout fields are replaced
// with their secure defaults from [DefaultServerOptions].
func NewServerWith(opts ServerOptions, handler http.Handler) *http.Server {
	defaults := DefaultServerOptions()

	if opts.ReadHeaderTimeout == 0 {
		opts.ReadHeaderTimeout = defaults.ReadHeaderTimeout
	}

	if opts.ReadTimeout == 0 {
		opts.ReadTimeout = defaults.ReadTimeout
	}

	if opts.WriteTimeout == 0 {
		opts.WriteTimeout = defaults.WriteTimeout
	}

	if opts.IdleTimeout == 0 {
		opts.IdleTimeout = defaults.IdleTimeout
	}

	if opts.MaxHeaderBytes == 0 {
		opts.MaxHeaderBytes = defaults.MaxHeaderBytes
	}

	return &http.Server{
		Addr:              opts.Addr,
		Handler:           handler,
		ReadHeaderTimeout: opts.ReadHeaderTimeout,
		ReadTimeout:       opts.ReadTimeout,
		WriteTimeout:      opts.WriteTimeout,
		IdleTimeout:       opts.IdleTimeout,
		MaxHeaderBytes:    opts.MaxHeaderBytes,
	}
}
