package koanf_test

import (
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/require"
	frameworkconfiguration "github.com/studiolambda/cosmos/framework/configuration/koanf"
)

func TestNewLoadsYAMLFilesInLexicalOrder(t *testing.T) {
	t.Parallel()

	driver, err := frameworkconfiguration.New(frameworkconfiguration.Config{
		Filesystem: fstest.MapFS{
			"config/base.yml":      &fstest.MapFile{Data: []byte("app:\n  name: cosmos\n  port: 8080\n")},
			"config/override.yaml": &fstest.MapFile{Data: []byte("app:\n  port: 9090\n")},
		},
	})

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

func TestNewRejectsNilFilesystem(t *testing.T) {
	t.Parallel()

	_, err := frameworkconfiguration.New(frameworkconfiguration.Config{})

	require.Error(t, err)
}
