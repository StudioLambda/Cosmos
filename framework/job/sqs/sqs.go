package sqs

import (
	"context"
	"encoding/json/v2"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awssqs "github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/studiolambda/cosmos/contract"
)

const maxRetryDelay = 15 * time.Minute

var (
	// ErrUnknownQueue is returned when a queue name has no configured SQS URL.
	ErrUnknownQueue = errors.New("unknown job queue")

	// ErrInvalidConfig is returned when SQS driver configuration is invalid.
	ErrInvalidConfig = errors.New("invalid SQS job driver configuration")

	// ErrUnsupportedRetryDelay is returned when an SQS visibility timeout cannot
	// represent the requested retry delay.
	ErrUnsupportedRetryDelay = errors.New("unsupported SQS job retry delay")

	// ErrInvalidDelivery is returned when SQS does not provide the data required
	// to settle a received delivery.
	ErrInvalidDelivery = errors.New("invalid SQS job delivery")

	// ErrInvalidHandler is returned when Consume is given a nil handler.
	ErrInvalidHandler = errors.New("invalid job delivery handler")

	// ErrDeliverySettled is returned when a delivery is settled more than once.
	ErrDeliverySettled = errors.New("job delivery already settled")
)

// SQSDriverConfig configures an SQS-backed job driver.
type SQSDriverConfig struct {
	// AWSConfig configures the AWS SDK client created by [New].
	AWSConfig aws.Config

	// Queues maps application queue names to their SQS queue URLs.
	Queues map[string]string

	// FailureQueueURL receives terminal failure reports when non-empty.
	FailureQueueURL string

	// WaitTimeSeconds controls SQS long polling. A zero value uses 20 seconds.
	WaitTimeSeconds int32

	// MaxMessages controls the maximum number of messages received per poll. A
	// zero value uses one message.
	MaxMessages int32
}

// DefaultSQSDriverConfig returns SQS driver defaults.
func DefaultSQSDriverConfig() SQSDriverConfig {
	return SQSDriverConfig{WaitTimeSeconds: 20, MaxMessages: 1}
}

type client interface {
	SendMessage(context.Context, *awssqs.SendMessageInput, ...func(*awssqs.Options)) (*awssqs.SendMessageOutput, error)
	ReceiveMessage(context.Context, *awssqs.ReceiveMessageInput, ...func(*awssqs.Options)) (*awssqs.ReceiveMessageOutput, error)
	DeleteMessage(context.Context, *awssqs.DeleteMessageInput, ...func(*awssqs.Options)) (*awssqs.DeleteMessageOutput, error)
	ChangeMessageVisibility(context.Context, *awssqs.ChangeMessageVisibilityInput, ...func(*awssqs.Options)) (*awssqs.ChangeMessageVisibilityOutput, error)
}

// Driver implements [contract.JobDispatcherDriver] and
// [contract.JobConsumerDriver] using Amazon SQS.
type Driver struct {
	client          client
	queues          map[string]string
	failureQueueURL string
	waitTimeSeconds int32
	maxMessages     int32
}

// New creates an SQS job driver and its AWS SDK client.
func New(config SQSDriverConfig) (*Driver, error) {
	return newFrom(awssqs.NewFromConfig(config.AWSConfig), config)
}

// NewFrom creates an SQS job driver using client. The caller retains ownership
// of client.
func NewFrom(client *awssqs.Client, config SQSDriverConfig) (*Driver, error) {
	if client == nil {
		return nil, fmt.Errorf("%w: nil client", ErrInvalidConfig)
	}

	return newFrom(client, config)
}

func newFrom(client client, config SQSDriverConfig) (*Driver, error) {

	if len(config.Queues) == 0 {
		return nil, fmt.Errorf("%w: no queues", ErrInvalidConfig)
	}

	queues := make(map[string]string, len(config.Queues))
	for name, url := range config.Queues {
		if strings.TrimSpace(name) == "" || strings.TrimSpace(url) == "" {
			return nil, fmt.Errorf("%w: queue name and URL are required", ErrInvalidConfig)
		}

		queues[name] = url
	}

	if config.WaitTimeSeconds == 0 {
		config.WaitTimeSeconds = 20
	}

	if config.WaitTimeSeconds < 1 || config.WaitTimeSeconds > 20 {
		return nil, fmt.Errorf("%w: wait time must be between 1 and 20 seconds", ErrInvalidConfig)
	}

	if config.MaxMessages == 0 {
		config.MaxMessages = 1
	}

	if config.MaxMessages < 1 || config.MaxMessages > 10 {
		return nil, fmt.Errorf("%w: maximum messages must be between 1 and 10", ErrInvalidConfig)
	}

	return &Driver{
		client:          client,
		queues:          queues,
		failureQueueURL: config.FailureQueueURL,
		waitTimeSeconds: config.WaitTimeSeconds,
		maxMessages:     config.MaxMessages,
	}, nil
}

// Dispatch serializes message and sends it to the configured URL for its queue.
func (driver *Driver) Dispatch(ctx context.Context, message contract.JobMessage) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	queueURL, err := driver.queueURL(message.Queue)
	if err != nil {
		return err
	}

	body, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("marshal job message: %w", err)
	}

	_, err = driver.client.SendMessage(ctx, &awssqs.SendMessageInput{
		MessageBody: aws.String(string(body)),
		QueueUrl:    aws.String(queueURL),
	})
	if err != nil {
		return fmt.Errorf("send job message: %w", err)
	}

	return nil
}

// Consume receives messages from queue until ctx is canceled or SQS returns an
// error. Handler errors are logged and left unsettled for SQS redelivery.
func (driver *Driver) Consume(ctx context.Context, queue string, handler contract.JobDeliveryHandler) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	if handler == nil {
		return ErrInvalidHandler
	}

	queueURL, err := driver.queueURL(queue)
	if err != nil {
		return err
	}

	for {
		output, err := driver.client.ReceiveMessage(ctx, &awssqs.ReceiveMessageInput{
			QueueUrl:                    aws.String(queueURL),
			WaitTimeSeconds:             driver.waitTimeSeconds,
			MaxNumberOfMessages:         driver.maxMessages,
			MessageSystemAttributeNames: []types.MessageSystemAttributeName{types.MessageSystemAttributeNameApproximateReceiveCount},
		})
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}

			return fmt.Errorf("receive job messages: %w", err)
		}

		for _, received := range output.Messages {
			driver.handle(ctx, queue, queueURL, received, handler)
		}
	}
}

func (driver *Driver) handle(ctx context.Context, queue, queueURL string, received types.Message, handler contract.JobDeliveryHandler) {
	message, err := parseMessage(queue, received)
	if err != nil {
		delivery := newDelivery(driver, queueURL, received, contract.JobMessage{})
		if delivery == nil {
			slog.Error("invalid SQS job delivery", "error", err, "queue", queue)

			return
		}

		if failureErr := delivery.fail(ctx, err, stringPointer(received.Body)); failureErr != nil {
			slog.Error("failed invalid SQS job message", "error", failureErr, "queue", queue)
		}

		return
	}

	delivery := newDelivery(driver, queueURL, received, message)
	if delivery == nil {
		slog.Error("invalid SQS job delivery", "error", ErrInvalidDelivery, "queue", queue, "job_id", message.ID)

		return
	}

	if err := handler(ctx, delivery); err != nil {
		slog.Error("SQS job delivery handler failed", "error", err, "queue", queue, "job_id", message.ID)
	}
}

func (driver *Driver) queueURL(queue string) (string, error) {
	url, ok := driver.queues[queue]
	if !ok {
		return "", fmt.Errorf("%w: %s", ErrUnknownQueue, queue)
	}

	return url, nil
}

func parseMessage(queue string, received types.Message) (contract.JobMessage, error) {
	if received.Body == nil {
		return contract.JobMessage{}, fmt.Errorf("%w: missing message body", ErrInvalidDelivery)
	}

	var message contract.JobMessage
	if err := json.Unmarshal([]byte(*received.Body), &message); err != nil {
		return contract.JobMessage{}, fmt.Errorf("unmarshal job message: %w", err)
	}

	message.Queue = queue
	message.ID = stringPointer(received.MessageId)
	message.Attempts = attempts(received.Attributes)

	return message, nil
}

func attempts(attributes map[string]string) int {
	count, err := strconv.Atoi(attributes["ApproximateReceiveCount"])
	if err != nil || count < 1 {
		return 0
	}

	return count
}

func stringPointer(value *string) string {
	if value == nil {
		return ""
	}

	return *value
}

type delivery struct {
	driver        *Driver
	queueURL      string
	receiptHandle string
	message       contract.JobMessage
	mu            sync.Mutex
	settling      bool
	settled       bool
}

func newDelivery(driver *Driver, queueURL string, received types.Message, message contract.JobMessage) *delivery {
	receiptHandle := stringPointer(received.ReceiptHandle)
	if receiptHandle == "" {
		return nil
	}

	return &delivery{driver: driver, queueURL: queueURL, receiptHandle: receiptHandle, message: message}
}

func (delivery *delivery) Message() contract.JobMessage {
	message := delivery.message
	message.Payload = slices.Clone(message.Payload)

	return message
}

func (delivery *delivery) Acknowledge(ctx context.Context) error {
	return delivery.settle(ctx, func() error { return delivery.delete(ctx) })
}

func (delivery *delivery) Retry(ctx context.Context, delay time.Duration) error {
	if delay < 0 || delay > maxRetryDelay {
		return fmt.Errorf("%w: %s exceeds %s", ErrUnsupportedRetryDelay, delay, maxRetryDelay)
	}

	seconds := int32(delay / time.Second)
	if delay%time.Second != 0 {
		seconds++
	}

	return delivery.settle(ctx, func() error {
		_, err := delivery.driver.client.ChangeMessageVisibility(ctx, &awssqs.ChangeMessageVisibilityInput{
			QueueUrl:          aws.String(delivery.queueURL),
			ReceiptHandle:     aws.String(delivery.receiptHandle),
			VisibilityTimeout: seconds,
		})

		return err
	})
}

func (delivery *delivery) Reject(ctx context.Context) error {
	return delivery.settle(ctx, func() error { return delivery.delete(ctx) })
}

func (delivery *delivery) Fail(ctx context.Context, err error) error {
	return delivery.settle(ctx, func() error { return delivery.fail(ctx, err, "") })
}

func (delivery *delivery) fail(ctx context.Context, err error, rawMessage string) error {
	if delivery.driver.failureQueueURL != "" {
		body, marshalErr := json.Marshal(failureMessage{Message: delivery.message, Error: errorText(err), RawMessage: rawMessage})
		if marshalErr != nil {
			return fmt.Errorf("marshal job failure: %w", marshalErr)
		}

		_, sendErr := delivery.driver.client.SendMessage(ctx, &awssqs.SendMessageInput{
			MessageBody: aws.String(string(body)),
			QueueUrl:    aws.String(delivery.driver.failureQueueURL),
		})
		if sendErr != nil {
			return fmt.Errorf("send job failure: %w", sendErr)
		}
	} else {
		slog.Error("SQS job delivery failed", "error", err, "queue", delivery.message.Queue, "job_id", delivery.message.ID)
	}

	return delivery.delete(ctx)
}

func (delivery *delivery) delete(ctx context.Context) error {
	_, err := delivery.driver.client.DeleteMessage(ctx, &awssqs.DeleteMessageInput{
		QueueUrl:      aws.String(delivery.queueURL),
		ReceiptHandle: aws.String(delivery.receiptHandle),
	})

	return err
}

func (delivery *delivery) settle(ctx context.Context, operation func() error) error {
	delivery.mu.Lock()
	if delivery.settled || delivery.settling {
		delivery.mu.Unlock()

		return ErrDeliverySettled
	}

	if err := ctx.Err(); err != nil {
		delivery.mu.Unlock()

		return err
	}

	delivery.settling = true
	delivery.mu.Unlock()

	if err := operation(); err != nil {
		delivery.mu.Lock()
		delivery.settling = false
		delivery.mu.Unlock()

		return err
	}

	delivery.mu.Lock()
	delivery.settling = false
	delivery.settled = true
	delivery.mu.Unlock()

	return nil
}

type failureMessage struct {
	Message    contract.JobMessage `json:"message"`
	Error      string              `json:"error"`
	RawMessage string              `json:"raw_message,omitzero"`
}

func errorText(err error) string {
	if err == nil {
		return "job failed"
	}

	return err.Error()
}
