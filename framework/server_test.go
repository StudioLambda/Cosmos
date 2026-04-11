package framework_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/studiolambda/cosmos/framework"

	"github.com/stretchr/testify/require"
)

func TestNewServerUsesDefaultOptions(t *testing.T) {
	t.Parallel()

	mux := http.NewServeMux()
	server := framework.NewServer(":9090", mux)

	require.Equal(t, ":9090", server.Addr)
	require.Equal(t, mux, server.Handler)
	require.Equal(t, 10*time.Second, server.ReadHeaderTimeout)
	require.Equal(t, 30*time.Second, server.ReadTimeout)
	require.Equal(t, 60*time.Second, server.WriteTimeout)
	require.Equal(t, 120*time.Second, server.IdleTimeout)
	require.Equal(t, 1<<20, server.MaxHeaderBytes)
}

func TestNewServerWithCustomOptions(t *testing.T) {
	t.Parallel()

	mux := http.NewServeMux()
	opts := framework.ServerOptions{
		Addr:              ":3000",
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    512 << 10,
	}

	server := framework.NewServerWith(opts, mux)

	require.Equal(t, ":3000", server.Addr)
	require.Equal(t, mux, server.Handler)
	require.Equal(t, 5*time.Second, server.ReadHeaderTimeout)
	require.Equal(t, 15*time.Second, server.ReadTimeout)
	require.Equal(t, 30*time.Second, server.WriteTimeout)
	require.Equal(t, 60*time.Second, server.IdleTimeout)
	require.Equal(t, 512<<10, server.MaxHeaderBytes)
}

func TestWithDefaultsFillsZeroValues(t *testing.T) {
	t.Parallel()

	opts := framework.ServerOptions{
		Addr: ":4000",
	}

	server := framework.NewServerWith(opts, nil)

	require.Equal(t, ":4000", server.Addr)
	require.Equal(t, 10*time.Second, server.ReadHeaderTimeout)
	require.Equal(t, 30*time.Second, server.ReadTimeout)
	require.Equal(t, 60*time.Second, server.WriteTimeout)
	require.Equal(t, 120*time.Second, server.IdleTimeout)
	require.Equal(t, 1<<20, server.MaxHeaderBytes)
}

func TestWithDefaultsPreservesNonZeroValues(t *testing.T) {
	t.Parallel()

	opts := framework.ServerOptions{
		Addr:              ":5000",
		ReadHeaderTimeout: 1 * time.Second,
		ReadTimeout:       2 * time.Second,
		WriteTimeout:      3 * time.Second,
		IdleTimeout:       4 * time.Second,
		MaxHeaderBytes:    256,
	}

	server := framework.NewServerWith(opts, nil)

	require.Equal(t, 1*time.Second, server.ReadHeaderTimeout)
	require.Equal(t, 2*time.Second, server.ReadTimeout)
	require.Equal(t, 3*time.Second, server.WriteTimeout)
	require.Equal(t, 4*time.Second, server.IdleTimeout)
	require.Equal(t, 256, server.MaxHeaderBytes)
}

func TestDefaultServerOptionsValues(t *testing.T) {
	t.Parallel()

	defaults := framework.DefaultServerOptions

	require.Equal(t, ":8080", defaults.Addr)
	require.Equal(t, 10*time.Second, defaults.ReadHeaderTimeout)
	require.Equal(t, 30*time.Second, defaults.ReadTimeout)
	require.Equal(t, 60*time.Second, defaults.WriteTimeout)
	require.Equal(t, 120*time.Second, defaults.IdleTimeout)
	require.Equal(t, 1<<20, defaults.MaxHeaderBytes)
}
