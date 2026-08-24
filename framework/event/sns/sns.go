package sns

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awssns "github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/studiolambda/cosmos/contract"
)

var (
	// ErrUnknownEvent is returned when an event has no configured SNS topic.
	ErrUnknownEvent = errors.New("unknown event")

	// ErrInvalidConfig is returned when SNS publisher configuration is invalid.
	ErrInvalidConfig = errors.New("invalid SNS event publisher configuration")
)

// SNSPublisherConfig configures an SNS-backed event publisher.
type SNSPublisherConfig struct {
	// AWSConfig configures the AWS SDK client created by [New].
	AWSConfig aws.Config

	// Topics maps application event names to SNS topic ARNs.
	Topics map[string]string
}

// DefaultSNSPublisherConfig returns the default SNS event publisher configuration.
func DefaultSNSPublisherConfig() SNSPublisherConfig {
	return SNSPublisherConfig{}
}

// FromConfiguration populates the declarative publisher configuration from configuration.
func (config *SNSPublisherConfig) FromConfiguration(configuration *contract.Configuration) {
	awsConfig := config.AWSConfig
	*config = DefaultSNSPublisherConfig()
	config.AWSConfig = awsConfig
	config.Topics = configuration.GetOr("topics", config.Topics)
}

type client interface {
	Publish(context.Context, *awssns.PublishInput, ...func(*awssns.Options)) (*awssns.PublishOutput, error)
}

// Driver implements [contract.EventPublisherDriver] using Amazon SNS.
//
// Driver sends each payload as the SNS message without a Subject, preserving
// raw payload bytes and allowing configured event names that SNS Subjects do
// not support.
type Driver struct {
	client client
	topics map[string]string
}

// New creates an SNS event publisher and its AWS SDK client.
func New(config SNSPublisherConfig) (*Driver, error) {
	return newFrom(awssns.NewFromConfig(config.AWSConfig), config)
}

// NewFrom creates an SNS event publisher using client. The caller retains
// ownership of client.
func NewFrom(client *awssns.Client, config SNSPublisherConfig) (*Driver, error) {
	if client == nil {
		return nil, fmt.Errorf("%w: nil client", ErrInvalidConfig)
	}

	return newFrom(client, config)
}

func newFrom(client client, config SNSPublisherConfig) (*Driver, error) {

	if len(config.Topics) == 0 {
		return nil, fmt.Errorf("%w: no topics", ErrInvalidConfig)
	}

	for event, topicARN := range config.Topics {
		if strings.TrimSpace(event) == "" || strings.TrimSpace(topicARN) == "" {
			return nil, fmt.Errorf("%w: event name and topic ARN are required", ErrInvalidConfig)
		}
	}

	return &Driver{client: client, topics: maps.Clone(config.Topics)}, nil
}

// Publish sends payload to the configured SNS topic for event. A successful
// result means SNS accepted the message, not that a subscriber processed it.
func (driver *Driver) Publish(ctx context.Context, event string, payload []byte) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	topicARN, ok := driver.topics[event]
	if !ok {
		return fmt.Errorf("%w: %s", ErrUnknownEvent, event)
	}

	_, err := driver.client.Publish(ctx, &awssns.PublishInput{
		Message:  aws.String(string(payload)),
		TopicArn: aws.String(topicARN),
	})
	if err != nil {
		return fmt.Errorf("publish SNS event: %w", err)
	}

	return nil
}
