package amqp

import (
	"testing"
	"time"

	"github.com/studiolambda/cosmos/contract"
	frameworkconfiguration "github.com/studiolambda/cosmos/framework/configuration"
	configuration "github.com/studiolambda/cosmos/framework/configuration/koanf"

	"github.com/stretchr/testify/require"
)

func TestDriverConfigFromConfiguration(t *testing.T) {
	t.Parallel()

	driver, err := configuration.New(frameworkconfiguration.Map(map[string]any{"jobs": map[string]any{"url": "amqps://example", "exchange": "jobs", "queues": map[string]any{"email": map[string]any{"name": "email-jobs", "routing_key": "email"}}, "prefetch": 2, "retry_delay": "3s", "max_retry_delay": "4s", "failure_queue": "failed-jobs"}}))
	require.NoError(t, err)

	config := DriverConfig{}
	config.FromConfiguration(contract.NewConfiguration(driver).Prefixed("jobs"))

	require.Equal(t, DriverConfig{URL: "amqps://example", Exchange: "jobs", Queues: map[string]QueueConfig{"email": {Name: "email-jobs", RoutingKey: "email"}}, Prefetch: 2, RetryDelay: 3 * time.Second, MaxRetryDelay: 4 * time.Second, FailureQueue: "failed-jobs"}, config)

	config = DriverConfig{}
	config.FromConfiguration(contract.NewConfiguration(driver).Prefixed("missing"))
	require.Equal(t, DefaultDriverConfig(), config)
}
