package nats

import (
	"testing"
	"time"

	"github.com/studiolambda/cosmos/contract"
	frameworkconfiguration "github.com/studiolambda/cosmos/framework/configuration"
	configuration "github.com/studiolambda/cosmos/framework/configuration/koanf"

	"github.com/stretchr/testify/require"
)

func TestNATSBrokerConfigFromConfiguration(t *testing.T) {
	t.Parallel()

	driver, err := configuration.New(frameworkconfiguration.Map(map[string]any{"broker": map[string]any{"urls": []string{"nats://one", "nats://two"}, "name": "events", "max_reconnects": 3, "reconnect_wait": "4s", "timeout": "5s", "username": "user", "password": "secret", "token": "token", "nkey_seed": "seed", "credentials_file": "credentials.creds", "root_cas": []string{"ca.pem"}}}))
	require.NoError(t, err)

	config := NATSBrokerConfig{}
	config.FromConfiguration(contract.NewConfiguration(driver).Prefixed("broker"))

	require.Equal(t, NATSBrokerConfig{URLs: []string{"nats://one", "nats://two"}, Name: "events", MaxReconnects: 3, ReconnectWait: 4 * time.Second, Timeout: 5 * time.Second, Username: "user", Password: "secret", Token: "token", NKeySeed: "seed", CredentialsFile: "credentials.creds", RootCAs: []string{"ca.pem"}}, config)

	config = NATSBrokerConfig{}
	config.FromConfiguration(contract.NewConfiguration(driver).Prefixed("missing"))
	require.Equal(t, DefaultNATSBrokerConfig(), config)
}
