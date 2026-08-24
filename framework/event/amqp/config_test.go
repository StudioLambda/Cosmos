package amqp

import (
	"testing"

	"github.com/studiolambda/cosmos/contract"
	frameworkconfiguration "github.com/studiolambda/cosmos/framework/configuration"
	configuration "github.com/studiolambda/cosmos/framework/configuration/koanf"

	"github.com/stretchr/testify/require"
)

func TestAMQPBrokerConfigFromConfiguration(t *testing.T) {
	t.Parallel()

	driver, err := configuration.New(frameworkconfiguration.Map(map[string]any{"broker": map[string]any{"url": "amqps://example", "exchange": "events"}}))
	require.NoError(t, err)

	config := AMQPBrokerConfig{}
	config.FromConfiguration(contract.NewConfiguration(driver).Prefixed("broker"))

	require.Equal(t, AMQPBrokerConfig{URL: "amqps://example", Exchange: "events"}, config)

	config = AMQPBrokerConfig{}
	config.FromConfiguration(contract.NewConfiguration(driver).Prefixed("missing"))
	require.Equal(t, DefaultAMQPBrokerConfig(), config)
}
