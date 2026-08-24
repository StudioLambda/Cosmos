package environment

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValuesParsesPrefixedNestedVariables(t *testing.T) {
	t.Setenv("COSMOS__HTTP__SERVER__PORT", "9090")
	t.Setenv("COSMOS__CACHE__REDIS__PASSWORD", "secret")

	values := Values("COSMOS")

	require.Equal(t, "9090", values["http.server.port"])
	require.Equal(t, "secret", values["cache.redis.password"])
}

func TestValuesPreservesEmptyValues(t *testing.T) {
	t.Setenv("COSMOS__CACHE__REDIS__PASSWORD", "")

	values := Values("COSMOS")

	value, found := values["cache.redis.password"]
	require.True(t, found)
	require.Equal(t, "", value)
}

func TestValuesIgnoresVariablesOutsidePrefix(t *testing.T) {
	t.Setenv("OTHER__HTTP__SERVER__PORT", "9090")

	values := Values("COSMOS")

	_, found := values["http.server.port"]
	require.False(t, found)
}
