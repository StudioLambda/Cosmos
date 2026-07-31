package viper_test

import (
	"testing"
	"testing/fstest"
	"time"

	"github.com/spf13/viper"
	"github.com/studiolambda/cosmos/contract"
	frameworkconfiguration "github.com/studiolambda/cosmos/framework/configuration"
	configuration "github.com/studiolambda/cosmos/framework/configuration/viper"

	"github.com/stretchr/testify/require"
)

func TestNewLoadsYAMLFilesInLexicalOrder(t *testing.T) {
	t.Parallel()

	driver, err := configuration.New(frameworkconfiguration.Filesystem(fstest.MapFS{
		"config/base.yml":      &fstest.MapFile{Data: []byte("app:\n  name: cosmos\n  port: 8080\n")},
		"config/override.yaml": &fstest.MapFile{Data: []byte("app:\n  port: 9090\n")},
	}))
	require.NoError(t, err)

	var port int
	err = driver.Unmarshal("app.port", &port)
	require.NoError(t, err)
	require.Equal(t, 9090, port)
}

func TestNewEnvironmentOverridesYAML(t *testing.T) {
	t.Setenv("COSMOS__HTTP__SERVER__PORT", "9090")

	driver, err := configuration.New(
		frameworkconfiguration.Filesystem(fstest.MapFS{
			"config/app.yml": &fstest.MapFile{Data: []byte("http:\n  server:\n    port: 8080\n")},
		}),
		frameworkconfiguration.Environment("COSMOS"),
	)
	require.NoError(t, err)

	var port int
	err = driver.Unmarshal("http.server.port", &port)
	require.NoError(t, err)
	require.Equal(t, 9090, port)
}

func TestViperUnmarshalDecodesDuration(t *testing.T) {
	t.Parallel()

	driver, err := configuration.New(frameworkconfiguration.Filesystem(fstest.MapFS{
		"config/app.yml": &fstest.MapFile{Data: []byte("http:\n  timeout: 30s\n")},
	}))
	require.NoError(t, err)

	var timeout time.Duration
	err = driver.Unmarshal("http.timeout", &timeout)
	require.NoError(t, err)
	require.Equal(t, 30*time.Second, timeout)
}

func TestViperUnmarshalReturnsNotFoundWhenMissing(t *testing.T) {
	t.Parallel()

	driver, err := configuration.New(frameworkconfiguration.Filesystem(fstest.MapFS{
		"config/app.yml": &fstest.MapFile{Data: []byte("app: {}\n")},
	}))
	require.NoError(t, err)

	var value string
	err = driver.Unmarshal("app.missing", &value)
	require.ErrorIs(t, err, contract.ErrConfigurationKeyNotFound)
}

func TestNewViperFromUsesProvidedInstance(t *testing.T) {
	t.Parallel()

	instance := viper.New()
	driver := configuration.NewViperFrom(instance, ".")

	require.Same(t, instance, driver.Instance())
}
