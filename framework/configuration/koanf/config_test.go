package koanf_test

import (
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/require"
	"github.com/studiolambda/cosmos/framework/configuration"
	frameworkconfiguration "github.com/studiolambda/cosmos/framework/configuration/koanf"
)

func TestNewLoadsYAMLFilesInLexicalOrder(t *testing.T) {
	t.Parallel()

	driver, err := frameworkconfiguration.New(
		configuration.Filesystem(fstest.MapFS{
			"config/base.yml":      &fstest.MapFile{Data: []byte("app:\n  name: cosmos\n  port: 8080\n")},
			"config/override.yaml": &fstest.MapFile{Data: []byte("app:\n  port: 9090\n")},
		}),
	)

	require.NoError(t, err)

	var name string
	err = driver.Unmarshal("app.name", &name)
	require.NoError(t, err)
	require.Equal(t, "cosmos", name)

	var port int
	err = driver.Unmarshal("app.port", &port)
	require.NoError(t, err)
	require.Equal(t, 9090, port)
}

func TestNewRejectsNilFilesystemProvider(t *testing.T) {
	t.Parallel()

	_, err := frameworkconfiguration.New(configuration.Filesystem(nil))

	require.Error(t, err)
}

func TestNewEnvironmentOverridesYAML(t *testing.T) {
	t.Setenv("COSMOS__APP__PORT", "9090")

	driver, err := frameworkconfiguration.New(
		configuration.Filesystem(fstest.MapFS{
			"config/app.yml": &fstest.MapFile{Data: []byte("app:\n  port: 8080\n")},
		}),
		configuration.Environment("COSMOS"),
	)
	require.NoError(t, err)

	var port int
	err = driver.Unmarshal("app.port", &port)

	require.NoError(t, err)
	require.Equal(t, 9090, port)
}
