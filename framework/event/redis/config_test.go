package redis

import (
	"testing"

	"github.com/studiolambda/cosmos/contract"
	frameworkconfiguration "github.com/studiolambda/cosmos/framework/configuration"
	configuration "github.com/studiolambda/cosmos/framework/configuration/koanf"

	"github.com/stretchr/testify/require"
)

func TestRedisBrokerConfigFromConfiguration(t *testing.T) {
	t.Parallel()

	driver, err := configuration.New(frameworkconfiguration.Map(map[string]any{"broker": map[string]any{"network": "unix", "addr": "/tmp/redis.sock", "username": "user", "password": "secret", "db": 2}}))
	require.NoError(t, err)

	config := RedisBrokerConfig{}
	config.FromConfiguration(contract.NewConfiguration(driver).Prefixed("broker"))

	require.Equal(t, RedisBrokerConfig{Network: "unix", Addr: "/tmp/redis.sock", Username: "user", Password: "secret", DB: 2}, config)

	config = RedisBrokerConfig{}
	config.FromConfiguration(contract.NewConfiguration(driver).Prefixed("missing"))
	require.Equal(t, DefaultRedisBrokerConfig(), config)
}
