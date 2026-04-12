package framework

import (
	"net/http"
	"time"
)

// ServerOptions configures the HTTP server created by [NewServer].
// All zero-valued fields default to secure values from [DefaultServerOptions].
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

// DefaultServerOptions returns the default server options with secure
// timeout values. Each call returns a fresh copy, preventing
// accidental mutation of shared defaults.
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

// withDefaults returns a copy of the options with zero values
// replaced by the corresponding [DefaultServerOptions] fields.
func (options ServerOptions) withDefaults() ServerOptions {
	defaults := DefaultServerOptions()

	if options.ReadHeaderTimeout == 0 {
		options.ReadHeaderTimeout = defaults.ReadHeaderTimeout
	}

	if options.ReadTimeout == 0 {
		options.ReadTimeout = defaults.ReadTimeout
	}

	if options.WriteTimeout == 0 {
		options.WriteTimeout = defaults.WriteTimeout
	}

	if options.IdleTimeout == 0 {
		options.IdleTimeout = defaults.IdleTimeout
	}

	if options.MaxHeaderBytes == 0 {
		options.MaxHeaderBytes = defaults.MaxHeaderBytes
	}

	return options
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
	opts = opts.withDefaults()

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
