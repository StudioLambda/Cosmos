package mqtt

import (
	"testing"

	"github.com/studiolambda/cosmos/contract"
	frameworkconfiguration "github.com/studiolambda/cosmos/framework/configuration"
	configuration "github.com/studiolambda/cosmos/framework/configuration/koanf"

	"github.com/stretchr/testify/require"
)

func TestMQTTBrokerConfigFromConfiguration(t *testing.T) {
	t.Parallel()

	driver, err := configuration.New(frameworkconfiguration.Map(map[string]any{"broker": map[string]any{"urls": []string{"mqtts://one", "mqtts://two"}, "qos": 2, "username": "user", "password": "secret", "keep_alive": 45}}))
	require.NoError(t, err)

	config := MQTTBrokerConfig{}
	config.FromConfiguration(contract.NewConfiguration(driver).Prefixed("broker"))

	require.Equal(t, MQTTBrokerConfig{URLs: []string{"mqtts://one", "mqtts://two"}, QoS: 2, Username: "user", Password: "secret", KeepAlive: 45}, config)

	config = MQTTBrokerConfig{}
	config.FromConfiguration(contract.NewConfiguration(driver).Prefixed("missing"))
	require.Equal(t, DefaultMQTTBrokerConfig(), config)
}
