package configuration_test

import (
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/require"
	"github.com/studiolambda/cosmos/framework/configuration"
)

func TestMapAcceptsDottedAndNestedKeys(t *testing.T) {
	t.Parallel()

	for range 100 {
		values, err := configuration.Resolve(configuration.Map(map[string]any{
			"http.server.port": 8080,
			"http": map[string]any{
				"server": map[string]any{
					"host": "0.0.0.0",
				},
			},
		}))

		require.NoError(t, err)
		require.Equal(t, map[string]any{
			"http": map[string]any{
				"server": map[string]any{
					"host": "0.0.0.0",
					"port": 8080,
				},
			},
		}, values)
	}
}

func TestResolveLaterProvidersOverrideEarlierValues(t *testing.T) {
	t.Parallel()

	values, err := configuration.Resolve(
		configuration.Map(map[string]any{
			"http": map[string]any{
				"server": map[string]any{
					"host": "0.0.0.0",
					"port": 8080,
				},
			},
		}),
		configuration.Map(map[string]any{"http.server.port": 9090}),
	)

	require.NoError(t, err)
	require.Equal(t, map[string]any{
		"http": map[string]any{
			"server": map[string]any{
				"host": "0.0.0.0",
				"port": 9090,
			},
		},
	}, values)
}

func TestFilesystemLoadsYAMLFilesInLexicalOrder(t *testing.T) {
	t.Parallel()

	values, err := configuration.Resolve(configuration.Filesystem(fstest.MapFS{
		"config/base.yml":      &fstest.MapFile{Data: []byte("app:\n  name: cosmos\n  port: 8080\n")},
		"config/override.yaml": &fstest.MapFile{Data: []byte("app:\n  port: 9090\n")},
	}))

	require.NoError(t, err)
	require.Equal(t, map[string]any{
		"app": map[string]any{
			"name": "cosmos",
			"port": 9090,
		},
	}, values)
}

func TestEnvironmentOverridesEarlierProviders(t *testing.T) {
	t.Setenv("COSMOS__HTTP__SERVER__PORT", "9090")

	values, err := configuration.Resolve(
		configuration.Map(map[string]any{"http.server.port": 8080}),
		configuration.Environment("COSMOS"),
	)

	require.NoError(t, err)
	require.Equal(t, "9090", values["http"].(map[string]any)["server"].(map[string]any)["port"])
}

func TestResolveWithoutProvidersReturnsEmptyValues(t *testing.T) {
	t.Parallel()

	values, err := configuration.Resolve()

	require.NoError(t, err)
	require.Empty(t, values)
}
