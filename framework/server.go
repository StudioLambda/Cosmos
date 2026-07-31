package framework

import (
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/studiolambda/cosmos/contract"
)

// ServerConfig configures the HTTP server created by [NewServer].
// All zero-valued fields default to secure values from [DefaultServerConfig].
type ServerConfig struct {
	// Host is the interface or host address to listen on.
	// Defaults to 0.0.0.0.
	Host string

	// Port is the TCP port to listen on.
	// Defaults to 8080.
	Port int

	// ReadHeaderTimeout limits the time allowed to read request
	// headers. Protects against Slowloris attacks.
	// Defaults to 10s.
	ReadHeaderTimeout time.Duration

	// ReadTimeout limits the total time for reading the entire
	// request including the body. Defaults to 30s.
	ReadTimeout time.Duration

	// WriteTimeout limits the time for writing the response.
	// Defaults to 60s. Set a negative value to disable the deadline for
	// long-lived streaming responses.
	WriteTimeout time.Duration

	// IdleTimeout limits the keep-alive idle time between
	// requests. Defaults to 120s.
	IdleTimeout time.Duration

	// MaxHeaderBytes limits the maximum size of request headers.
	// Defaults to 1 MB (Go's default).
	MaxHeaderBytes int
}

// FromConfiguration populates the server configuration from configuration.
func (config *ServerConfig) FromConfiguration(configuration *contract.Configuration) {
	config.Host = configuration.GetOr("host", "")
	config.Port = configuration.GetOr("port", 0)
	config.ReadHeaderTimeout = configuration.GetOr("read_header_timeout", time.Duration(0))
	config.ReadTimeout = configuration.GetOr("read_timeout", time.Duration(0))
	config.WriteTimeout = configuration.GetOr("write_timeout", time.Duration(0))
	config.IdleTimeout = configuration.GetOr("idle_timeout", time.Duration(0))
	config.MaxHeaderBytes = configuration.GetOr("max_header_bytes", 0)
}

// DefaultServerConfig returns the default server configuration with secure
// timeout values. Each call returns a fresh copy, preventing
// accidental mutation of shared defaults.
//
// Example:
//
//	config := framework.DefaultServerConfig()
//	config.Port = 9090
func DefaultServerConfig() ServerConfig {
	return ServerConfig{
		Host:              "0.0.0.0",
		Port:              8080,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       120 * time.Second,
		MaxHeaderBytes:    1 << 20, // 1 MB
	}
}

// withDefaults returns a copy of the config with zero values
// replaced by the corresponding [DefaultServerConfig] fields.
func (config ServerConfig) withDefaults() ServerConfig {
	defaults := DefaultServerConfig()

	if config.Host == "" {
		config.Host = defaults.Host
	}

	if config.Port == 0 {
		config.Port = defaults.Port
	}

	if config.ReadHeaderTimeout == 0 {
		config.ReadHeaderTimeout = defaults.ReadHeaderTimeout
	}

	if config.ReadTimeout == 0 {
		config.ReadTimeout = defaults.ReadTimeout
	}

	if config.WriteTimeout == 0 {
		config.WriteTimeout = defaults.WriteTimeout
	}

	if config.WriteTimeout < 0 {
		config.WriteTimeout = 0
	}

	if config.IdleTimeout == 0 {
		config.IdleTimeout = defaults.IdleTimeout
	}

	if config.MaxHeaderBytes == 0 {
		config.MaxHeaderBytes = defaults.MaxHeaderBytes
	}

	return config
}

// NewServer creates an [http.Server] with the provided
// configuration and handler. Zero-valued timeout fields are replaced
// with their secure defaults from [DefaultServerConfig].
//
// Example:
//
//	server := framework.NewServer(framework.ServerConfig{Host: "0.0.0.0", Port: 8443}, app)
//	_ = server
func NewServer(config ServerConfig, handler http.Handler) *http.Server {
	config = config.withDefaults()

	return &http.Server{
		Addr:              net.JoinHostPort(config.Host, strconv.Itoa(config.Port)),
		Handler:           handler,
		ReadHeaderTimeout: config.ReadHeaderTimeout,
		ReadTimeout:       config.ReadTimeout,
		WriteTimeout:      config.WriteTimeout,
		IdleTimeout:       config.IdleTimeout,
		MaxHeaderBytes:    config.MaxHeaderBytes,
	}
}
