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

// DefaultServerOptions holds secure timeout defaults suitable
// for production deployments. These protect against Slowloris
// and other connection-exhaustion attacks.
var DefaultServerOptions = ServerOptions{
	Addr:              ":8080",
	ReadHeaderTimeout: 10 * time.Second,
	ReadTimeout:       30 * time.Second,
	WriteTimeout:      60 * time.Second,
	IdleTimeout:       120 * time.Second,
	MaxHeaderBytes:    1 << 20, // 1 MB
}

// withDefaults returns a copy of the options with zero values
// replaced by the corresponding [DefaultServerOptions] fields.
func (o ServerOptions) withDefaults() ServerOptions {
	if o.ReadHeaderTimeout == 0 {
		o.ReadHeaderTimeout = DefaultServerOptions.ReadHeaderTimeout
	}

	if o.ReadTimeout == 0 {
		o.ReadTimeout = DefaultServerOptions.ReadTimeout
	}

	if o.WriteTimeout == 0 {
		o.WriteTimeout = DefaultServerOptions.WriteTimeout
	}

	if o.IdleTimeout == 0 {
		o.IdleTimeout = DefaultServerOptions.IdleTimeout
	}

	if o.MaxHeaderBytes == 0 {
		o.MaxHeaderBytes = DefaultServerOptions.MaxHeaderBytes
	}

	return o
}

// NewServer creates an [http.Server] with secure timeout defaults
// using the given handler. It applies [DefaultServerOptions]
// values, protecting against Slowloris and connection-exhaustion
// attacks that are possible when using [http.ListenAndServe]
// directly (which sets all timeouts to zero/infinite).
func NewServer(addr string, handler http.Handler) *http.Server {
	opts := DefaultServerOptions
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
