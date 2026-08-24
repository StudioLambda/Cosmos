package sqs

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/studiolambda/cosmos/contract"
	frameworkconfiguration "github.com/studiolambda/cosmos/framework/configuration"
	configuration "github.com/studiolambda/cosmos/framework/configuration/koanf"

	"github.com/stretchr/testify/require"
)

func TestSQSDriverConfigFromConfiguration(t *testing.T) {
	t.Parallel()

	driver, err := configuration.New(frameworkconfiguration.Map(map[string]any{"jobs": map[string]any{"queues": map[string]string{"email": "https://sqs.example/email"}, "failure_queue_url": "https://sqs.example/failed", "wait_time_seconds": 10, "max_messages": 2}}))
	require.NoError(t, err)

	config := SQSDriverConfig{AWSConfig: aws.Config{Region: "eu-west-1"}}
	config.FromConfiguration(contract.NewConfiguration(driver).Prefixed("jobs"))

	require.Equal(t, SQSDriverConfig{AWSConfig: aws.Config{Region: "eu-west-1"}, Queues: map[string]string{"email": "https://sqs.example/email"}, FailureQueueURL: "https://sqs.example/failed", WaitTimeSeconds: 10, MaxMessages: 2}, config)

	config = SQSDriverConfig{}
	config.FromConfiguration(contract.NewConfiguration(driver).Prefixed("missing"))
	require.Equal(t, DefaultSQSDriverConfig(), config)
}
