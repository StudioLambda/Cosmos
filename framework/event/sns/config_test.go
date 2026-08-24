package sns

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/studiolambda/cosmos/contract"
	frameworkconfiguration "github.com/studiolambda/cosmos/framework/configuration"
	configuration "github.com/studiolambda/cosmos/framework/configuration/koanf"

	"github.com/stretchr/testify/require"
)

func TestSNSPublisherConfigFromConfiguration(t *testing.T) {
	t.Parallel()

	driver, err := configuration.New(frameworkconfiguration.Map(map[string]any{"publisher": map[string]any{"topics": map[string]string{"users.created": "arn:aws:sns:example:users"}}}))
	require.NoError(t, err)

	config := SNSPublisherConfig{AWSConfig: aws.Config{Region: "eu-west-1"}}
	config.FromConfiguration(contract.NewConfiguration(driver).Prefixed("publisher"))

	require.Equal(t, SNSPublisherConfig{AWSConfig: aws.Config{Region: "eu-west-1"}, Topics: map[string]string{"users.created": "arn:aws:sns:example:users"}}, config)

	config = SNSPublisherConfig{}
	config.FromConfiguration(contract.NewConfiguration(driver).Prefixed("missing"))
	require.Equal(t, DefaultSNSPublisherConfig(), config)
}
