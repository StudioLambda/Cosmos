package framework_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/studiolambda/cosmos/framework"

	"github.com/stretchr/testify/require"
)

func TestNewServerUsesDefaultConfig(t *testing.T) {
	t.Parallel()

	mux := http.NewServeMux()
	config := framework.DefaultServerConfig()
	config.Port = 9090
	server := framework.NewServer(config, mux)

	require.Equal(t, "0.0.0.0:9090", server.Addr)
	require.Equal(t, mux, server.Handler)
	require.Equal(t, 10*time.Second, server.ReadHeaderTimeout)
	require.Equal(t, 30*time.Second, server.ReadTimeout)
	require.Equal(t, 60*time.Second, server.WriteTimeout)
	require.Equal(t, 120*time.Second, server.IdleTimeout)
	require.Equal(t, 1<<20, server.MaxHeaderBytes)
}

func TestNewServerWithCustomConfig(t *testing.T) {
	t.Parallel()

	mux := http.NewServeMux()
	config := framework.ServerConfig{
		Host:              "127.0.0.1",
		Port:              3000,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    512 << 10,
	}

	server := framework.NewServer(config, mux)

	require.Equal(t, "127.0.0.1:3000", server.Addr)
	require.Equal(t, mux, server.Handler)
	require.Equal(t, 5*time.Second, server.ReadHeaderTimeout)
	require.Equal(t, 15*time.Second, server.ReadTimeout)
	require.Equal(t, 30*time.Second, server.WriteTimeout)
	require.Equal(t, 60*time.Second, server.IdleTimeout)
	require.Equal(t, 512<<10, server.MaxHeaderBytes)
}

func TestWithDefaultsFillsZeroValues(t *testing.T) {
	t.Parallel()

	config := framework.ServerConfig{
		Port: 4000,
	}

	server := framework.NewServer(config, nil)

	require.Equal(t, "0.0.0.0:4000", server.Addr)
	require.Equal(t, 10*time.Second, server.ReadHeaderTimeout)
	require.Equal(t, 30*time.Second, server.ReadTimeout)
	require.Equal(t, 60*time.Second, server.WriteTimeout)
	require.Equal(t, 120*time.Second, server.IdleTimeout)
	require.Equal(t, 1<<20, server.MaxHeaderBytes)
}

func TestWithDefaultsPreservesNonZeroValues(t *testing.T) {
	t.Parallel()

	config := framework.ServerConfig{
		Host:              "127.0.0.1",
		Port:              5000,
		ReadHeaderTimeout: 1 * time.Second,
		ReadTimeout:       2 * time.Second,
		WriteTimeout:      3 * time.Second,
		IdleTimeout:       4 * time.Second,
		MaxHeaderBytes:    256,
	}

	server := framework.NewServer(config, nil)

	require.Equal(t, 1*time.Second, server.ReadHeaderTimeout)
	require.Equal(t, 2*time.Second, server.ReadTimeout)
	require.Equal(t, 3*time.Second, server.WriteTimeout)
	require.Equal(t, 4*time.Second, server.IdleTimeout)
	require.Equal(t, 256, server.MaxHeaderBytes)
}

func TestDefaultServerConfigValues(t *testing.T) {
	t.Parallel()

	defaults := framework.DefaultServerConfig()

	require.Equal(t, "0.0.0.0", defaults.Host)
	require.Equal(t, 8080, defaults.Port)
	require.Equal(t, 10*time.Second, defaults.ReadHeaderTimeout)
	require.Equal(t, 30*time.Second, defaults.ReadTimeout)
	require.Equal(t, 60*time.Second, defaults.WriteTimeout)
	require.Equal(t, 120*time.Second, defaults.IdleTimeout)
	require.Equal(t, 1<<20, defaults.MaxHeaderBytes)
}

func TestDefaultServerConfigReturnsNewCopy(t *testing.T) {
	t.Parallel()

	first := framework.DefaultServerConfig()
	second := framework.DefaultServerConfig()

	first.Host = "127.0.0.1"
	first.Port = 9999
	first.ReadTimeout = 999 * time.Second

	require.Equal(t, "0.0.0.0", second.Host)
	require.Equal(t, 8080, second.Port)
	require.Equal(t, 30*time.Second, second.ReadTimeout)
}
