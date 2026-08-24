package jetstream

import (
	"testing"
	"time"

	"github.com/studiolambda/cosmos/contract"
	frameworkconfiguration "github.com/studiolambda/cosmos/framework/configuration"
	configuration "github.com/studiolambda/cosmos/framework/configuration/koanf"

	"github.com/stretchr/testify/require"
)

func TestConfigFromConfiguration(t *testing.T) {
	t.Parallel()

	driver, err := configuration.New(frameworkconfiguration.Map(map[string]any{"jobs": map[string]any{"urls": []string{"nats://one"}, "stream": "JOBS", "queues": map[string]any{"email": map[string]any{"subject": "jobs.email", "durable": "email-worker"}}, "failure_subject": "jobs.failed", "provision": true, "batch": 2, "max_ack_pending": 4, "ack_wait": "5s", "max_deliver": 3}}))
	require.NoError(t, err)

	config := Config{}
	config.FromConfiguration(contract.NewConfiguration(driver).Prefixed("jobs"))

	require.Equal(t, Config{URLs: []string{"nats://one"}, Stream: "JOBS", Queues: map[string]QueueConfig{"email": {Subject: "jobs.email", Durable: "email-worker"}}, FailureSubject: "jobs.failed", Provision: true, Batch: 2, MaxAckPending: 4, AckWait: 5 * time.Second, MaxDeliver: 3}, config)

	config = Config{}
	config.FromConfiguration(contract.NewConfiguration(driver).Prefixed("missing"))
	require.Equal(t, DefaultConfig(), config)
}
