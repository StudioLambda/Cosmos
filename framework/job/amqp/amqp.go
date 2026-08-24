package amqp

import (
	"context"
	"encoding/json/v2"
	"errors"
	"fmt"
	"maps"
	"math"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/studiolambda/cosmos/contract"

	amqp091 "github.com/rabbitmq/amqp091-go"
)

const (
	// DefaultExchange is the durable direct exchange used for Cosmos jobs.
	DefaultExchange   = "cosmos:jobs"
	defaultPrefetch   = 1
	defaultRetryDelay = time.Second
	attemptHeader     = "cosmos-attempt"
)

var (
	// ErrInvalidConfig is returned when [DriverConfig] cannot safely define job topology.
	ErrInvalidConfig = errors.New("invalid AMQP job driver configuration")

	// ErrUnknownQueue is returned when an application queue has no configured topology.
	ErrUnknownQueue = errors.New("unknown job queue")

	// ErrInvalidHandler is returned when Consume is given a nil handler.
	ErrInvalidHandler = errors.New("invalid job delivery handler")

	// ErrInvalidDelivery is returned when an AMQP delivery cannot be settled or decoded.
	ErrInvalidDelivery = errors.New("invalid AMQP job delivery")

	// ErrDeliverySettled is returned when a delivery is settled more than once.
	ErrDeliverySettled = errors.New("job delivery already settled")

	// ErrDriverClosed is returned when an operation is attempted after Close.
	ErrDriverClosed = errors.New("AMQP job driver is closed")

	// ErrPublishNotConfirmed is returned when RabbitMQ negatively acknowledges a publish.
	ErrPublishNotConfirmed = errors.New("AMQP job publish was not confirmed")

	// ErrUnsupportedRetryDelay is returned when a retry delay exceeds the configured bound.
	ErrUnsupportedRetryDelay = errors.New("unsupported AMQP job retry delay")
)

// QueueConfig maps a logical application queue to a fixed broker queue and routing key.
// Neither value is read from untrusted job message data.
type QueueConfig struct {
	// Name is the durable RabbitMQ queue name.
	Name string

	// RoutingKey binds Name to [DriverConfig.Exchange].
	RoutingKey string
}

// DriverConfig configures a durable RabbitMQ job driver.
type DriverConfig struct {
	// URL is used only by [New] to create an AMQP connection.
	URL string

	// Exchange is a durable direct exchange. An empty value uses [DefaultExchange].
	Exchange string

	// Queues maps trusted logical application queue names to RabbitMQ topology.
	Queues map[string]QueueConfig

	// Prefetch is the number of unacknowledged messages per Consume invocation.
	// A zero value uses one.
	Prefetch int

	// RetryDelay is used by Retry when its delay is zero. A zero value uses one second.
	RetryDelay time.Duration

	// MaxRetryDelay bounds Retry delays and therefore the number of retry queues that
	// a caller can create. A zero value permits RabbitMQ's maximum TTL, about 49 days.
	MaxRetryDelay time.Duration

	// FailureQueue receives durable failure envelopes before the original job is
	// acknowledged. When empty, Fail terminally rejects the original delivery; any
	// resulting dead-lettering depends on broker policy.
	FailureQueue string

	// Logger records delivery failures. A nil logger discards records.
	Logger *contract.Logger
}

// DefaultDriverConfig returns the default AMQP job driver configuration.
func DefaultDriverConfig() DriverConfig {
	return DriverConfig{
		Exchange:   DefaultExchange,
		Prefetch:   defaultPrefetch,
		RetryDelay: defaultRetryDelay,
	}
}

// FromConfiguration populates the declarative driver configuration from configuration.
func (config *DriverConfig) FromConfiguration(configuration *contract.Configuration) {
	logger := config.Logger
	*config = DefaultDriverConfig()
	config.Logger = logger
	config.URL = configuration.GetOr("url", config.URL)
	config.Exchange = configuration.GetOr("exchange", config.Exchange)
	config.Queues = configuration.GetOr("queues", config.Queues)
	config.Prefetch = configuration.GetOr("prefetch", config.Prefetch)
	config.RetryDelay = configuration.GetOr("retry_delay", config.RetryDelay)
	config.MaxRetryDelay = configuration.GetOr("max_retry_delay", config.MaxRetryDelay)
	config.FailureQueue = configuration.GetOr("failure_queue", config.FailureQueue)
}

// Driver implements [contract.JobDispatcherDriver] and [contract.JobConsumerDriver]
// with durable RabbitMQ work queues. Dispatch returns only after RabbitMQ publisher
// confirms the persistent publish. It does not confirm that a worker processed it.
//
// Each Consume call opens its own channel, so multiple consumers safely compete for
// messages on the same durable queue.
type Driver struct {
	connection *amqp091.Connection
	publisher  *amqp091.Channel
	confirms   <-chan amqp091.Confirmation
	config     DriverConfig
	queues     map[string]QueueConfig
	owned      bool

	mu     sync.Mutex
	closed bool
	logger *contract.Logger
}

// New connects to RabbitMQ, declares the configured durable topology, and returns a
// driver that owns the new connection. [Driver.Close] closes that connection.
func New(config DriverConfig) (*Driver, error) {
	config, err := normalizeConfig(config, true)
	if err != nil {
		return nil, err
	}

	connection, err := amqp091.Dial(config.URL)
	if err != nil {
		return nil, fmt.Errorf("dial AMQP: %w", err)
	}

	driver, err := newFrom(config, connection, true)
	if err != nil {
		_ = connection.Close()

		return nil, err
	}

	return driver, nil
}

// NewFrom declares the configured durable topology using connection. The caller
// retains ownership of connection: [Driver.Close] closes only the driver's publish
// channel, not the supplied connection. The caller must keep it open for the
// driver's lifetime.
func NewFrom(connection *amqp091.Connection, config DriverConfig) (*Driver, error) {
	config, err := normalizeConfig(config, false)
	if err != nil {
		return nil, err
	}

	if connection == nil {
		return nil, fmt.Errorf("%w: nil connection", ErrInvalidConfig)
	}

	return newFrom(config, connection, false)
}

func newFrom(config DriverConfig, connection *amqp091.Connection, owned bool) (*Driver, error) {
	publisher, err := connection.Channel()
	if err != nil {
		return nil, fmt.Errorf("open AMQP publisher channel: %w", err)
	}

	if err := declareTopology(publisher, config); err != nil {
		_ = publisher.Close()

		return nil, err
	}

	if err := publisher.Confirm(false); err != nil {
		_ = publisher.Close()

		return nil, fmt.Errorf("enable AMQP publisher confirms: %w", err)
	}

	return &Driver{connection: connection, publisher: publisher, confirms: publisher.NotifyPublish(make(chan amqp091.Confirmation, 1)), config: config, queues: maps.Clone(config.Queues), owned: owned, logger: driverLogger(config.Logger)}, nil
}

// Dispatch serializes message as a [contract.JobMessage] and persistently publishes
// it to the configured routing key. The message queue is always replaced by the
// configured logical queue, preventing transport data from selecting broker queues.
func (driver *Driver) Dispatch(ctx context.Context, message contract.JobMessage) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	queue, err := driver.queue(message.Queue)
	if err != nil {
		return err
	}

	message.Attempts = 0
	body, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("marshal job message: %w", err)
	}

	return driver.publish(ctx, driver.config.Exchange, queue.RoutingKey, publishing(message.ID, body, 1))
}

// Consume receives deliveries from logical queue until ctx is canceled or the AMQP
// channel fails. It uses manual acknowledgement. Handler errors are logged and left
// unsettled so RabbitMQ can redeliver them when the channel or connection closes.
func (driver *Driver) Consume(ctx context.Context, name string, handler contract.JobDeliveryHandler) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	if handler == nil {
		return ErrInvalidHandler
	}

	queue, err := driver.queue(name)
	if err != nil {
		return err
	}

	channel, err := driver.consumerChannel()
	if err != nil {
		return err
	}
	defer channel.Close()

	if err := channel.Qos(driver.config.Prefetch, 0, false); err != nil {
		return fmt.Errorf("set AMQP job QoS: %w", err)
	}

	deliveries, err := channel.ConsumeWithContext(ctx, queue.Name, "", false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("consume AMQP job queue: %w", err)
	}

	for received := range deliveries {
		driver.handle(ctx, name, channel, received, handler)
	}

	if err := ctx.Err(); err != nil {
		return err
	}

	return errors.New("AMQP job delivery channel closed")
}

func (driver *Driver) handle(ctx context.Context, queue string, channel *amqp091.Channel, received amqp091.Delivery, handler contract.JobDeliveryHandler) {
	message, err := deliveryMessage(queue, received)
	delivery := &delivery{driver: driver, channel: channel, received: received, message: message}
	if err != nil {
		delivery.message.ID = received.MessageId
		if failureErr := delivery.Fail(ctx, err); failureErr != nil {
			driver.logger.Error("failed malformed AMQP job delivery", "error", failureErr, "queue", queue)
		}

		return
	}

	if err := handler(ctx, delivery); err != nil {
		driver.logger.Error("AMQP job delivery handler failed", "error", err, "queue", queue, "job_id", message.ID)
	}
}

func driverLogger(logger *contract.Logger) *contract.Logger {
	if logger == nil {
		return contract.NewLogger(nil)
	}

	return logger
}

func (driver *Driver) queue(name string) (QueueConfig, error) {
	queue, ok := driver.queues[name]
	if !ok {
		return QueueConfig{}, fmt.Errorf("%w: %s", ErrUnknownQueue, name)
	}

	return queue, nil
}

func (driver *Driver) consumerChannel() (*amqp091.Channel, error) {
	driver.mu.Lock()
	defer driver.mu.Unlock()

	if driver.closed {
		return nil, ErrDriverClosed
	}

	channel, err := driver.connection.Channel()
	if err != nil {
		return nil, fmt.Errorf("open AMQP consumer channel: %w", err)
	}

	return channel, nil
}

func (driver *Driver) publish(ctx context.Context, exchange, routingKey string, message amqp091.Publishing) error {
	driver.mu.Lock()
	defer driver.mu.Unlock()

	if driver.closed {
		return ErrDriverClosed
	}

	sequence := driver.publisher.GetNextPublishSeqNo()
	if err := driver.publisher.PublishWithContext(ctx, exchange, routingKey, false, false, message); err != nil {
		return fmt.Errorf("publish AMQP job: %w", err)
	}

	return waitConfirm(ctx, driver.confirms, sequence)
}

// Ping verifies that a new AMQP channel can be opened. It does not publish data.
func (driver *Driver) Ping(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	channel, err := driver.consumerChannel()
	if err != nil {
		return err
	}

	return channel.Close()
}

// Close closes the publisher channel. For drivers created by [New], it also closes
// the owned connection. It does not attempt graceful worker shutdown; cancel each
// Consume context before closing the connection.
func (driver *Driver) Close() error {
	driver.mu.Lock()
	if driver.closed {
		driver.mu.Unlock()

		return ErrDriverClosed
	}
	driver.closed = true
	publisher := driver.publisher
	connection := driver.connection
	owned := driver.owned
	driver.mu.Unlock()

	err := publisher.Close()
	if owned {
		connectionErr := connection.Close()
		if err != nil {
			return err
		}

		return connectionErr
	}

	return err
}

func normalizeConfig(config DriverConfig, requireURL bool) (DriverConfig, error) {
	if requireURL && strings.TrimSpace(config.URL) == "" {
		return DriverConfig{}, fmt.Errorf("%w: URL is required", ErrInvalidConfig)
	}

	if config.Exchange == "" {
		config.Exchange = DefaultExchange
	}

	if strings.TrimSpace(config.Exchange) == "" || len(config.Queues) == 0 {
		return DriverConfig{}, fmt.Errorf("%w: exchange and queues are required", ErrInvalidConfig)
	}

	if config.Prefetch == 0 {
		config.Prefetch = defaultPrefetch
	}
	if config.Prefetch < 1 {
		return DriverConfig{}, fmt.Errorf("%w: prefetch must be positive", ErrInvalidConfig)
	}

	if config.RetryDelay == 0 {
		config.RetryDelay = defaultRetryDelay
	}
	if config.RetryDelay < 0 || retryMilliseconds(config.RetryDelay) > math.MaxUint32 {
		return DriverConfig{}, fmt.Errorf("%w: invalid retry delay", ErrInvalidConfig)
	}
	if config.MaxRetryDelay < 0 || (config.MaxRetryDelay > 0 && retryMilliseconds(config.MaxRetryDelay) > math.MaxUint32) {
		return DriverConfig{}, fmt.Errorf("%w: invalid maximum retry delay", ErrInvalidConfig)
	}

	queues := make(map[string]QueueConfig, len(config.Queues))
	for logical, queue := range config.Queues {
		if strings.TrimSpace(logical) == "" || strings.TrimSpace(queue.Name) == "" || strings.TrimSpace(queue.RoutingKey) == "" {
			return DriverConfig{}, fmt.Errorf("%w: logical queue, queue name, and routing key are required", ErrInvalidConfig)
		}

		queues[logical] = queue
	}
	config.Queues = queues

	return config, nil
}

func declareTopology(channel *amqp091.Channel, config DriverConfig) error {
	if err := channel.ExchangeDeclare(config.Exchange, "direct", true, false, false, false, nil); err != nil {
		return fmt.Errorf("declare AMQP job exchange: %w", err)
	}

	for _, queue := range config.Queues {
		if _, err := channel.QueueDeclare(queue.Name, true, false, false, false, nil); err != nil {
			return fmt.Errorf("declare AMQP job queue: %w", err)
		}
		if err := channel.QueueBind(queue.Name, queue.RoutingKey, config.Exchange, false, nil); err != nil {
			return fmt.Errorf("bind AMQP job queue: %w", err)
		}
	}

	if config.FailureQueue != "" {
		if _, err := channel.QueueDeclare(config.FailureQueue, true, false, false, false, nil); err != nil {
			return fmt.Errorf("declare AMQP job failure queue: %w", err)
		}
	}

	return nil
}

func deliveryMessage(queue string, received amqp091.Delivery) (contract.JobMessage, error) {
	var message contract.JobMessage
	if err := json.Unmarshal(received.Body, &message); err != nil {
		return contract.JobMessage{Queue: queue}, fmt.Errorf("%w: unmarshal job message: %w", ErrInvalidDelivery, err)
	}

	message.Queue = queue
	message.ID = received.MessageId
	message.Attempts = attempts(received.Headers, received.Redelivered)

	return message, nil
}

func attempts(headers amqp091.Table, redelivered bool) int {
	attempt := headerAttempt(headers)
	if attempt == 0 {
		attempt = 1
	}
	if redelivered {
		attempt++
	}

	return attempt
}

func headerAttempt(headers amqp091.Table) int {
	value, ok := headers[attemptHeader]
	if !ok {
		return 0
	}

	switch attempt := value.(type) {
	case int32:
		return int(attempt)
	case int64:
		if attempt <= 0 || attempt > math.MaxInt {
			return 0
		}

		return int(attempt)
	case string:
		parsed, err := strconv.Atoi(attempt)
		if err == nil && parsed > 0 {
			return parsed
		}
	}

	return 0
}

func publishing(id string, body []byte, attempt int) amqp091.Publishing {
	return amqp091.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp091.Persistent,
		MessageId:    id,
		Headers:      amqp091.Table{attemptHeader: int32(attempt)},
		Body:         slices.Clone(body),
	}
}

type delivery struct {
	driver   *Driver
	channel  *amqp091.Channel
	received amqp091.Delivery
	message  contract.JobMessage
	mu       sync.Mutex
	settling bool
	settled  bool
}

func (delivery *delivery) Message() contract.JobMessage {
	message := delivery.message
	message.Payload = slices.Clone(message.Payload)

	return message
}

func (delivery *delivery) Acknowledge(ctx context.Context) error {
	return delivery.settle(ctx, func() error { return delivery.received.Ack(false) })
}

func (delivery *delivery) Retry(ctx context.Context, delay time.Duration) error {
	if delay == 0 {
		delay = delivery.driver.config.RetryDelay
	}
	if delay < 0 || (delivery.driver.config.MaxRetryDelay > 0 && delay > delivery.driver.config.MaxRetryDelay) || retryMilliseconds(delay) > math.MaxUint32 {
		return fmt.Errorf("%w: %s", ErrUnsupportedRetryDelay, delay)
	}

	if delay == 0 {
		return delivery.settle(ctx, func() error { return delivery.received.Nack(false, true) })
	}

	return delivery.settle(ctx, func() error {
		if err := delivery.delayedRetry(ctx, delay); err != nil {
			return err
		}

		return delivery.received.Ack(false)
	})
}

func (delivery *delivery) Reject(ctx context.Context) error {
	return delivery.settle(ctx, func() error { return delivery.received.Reject(false) })
}

func (delivery *delivery) Fail(ctx context.Context, err error) error {
	return delivery.settle(ctx, func() error {
		if delivery.driver.config.FailureQueue == "" {
			return delivery.received.Reject(false)
		}

		body, marshalErr := json.Marshal(failureMessage{Message: delivery.message, Error: errorText(err), RawMessage: string(delivery.received.Body)})
		if marshalErr != nil {
			return fmt.Errorf("marshal AMQP job failure: %w", marshalErr)
		}
		if publishErr := delivery.driver.publish(ctx, "", delivery.driver.config.FailureQueue, publishing(delivery.message.ID, body, delivery.message.Attempts)); publishErr != nil {
			return fmt.Errorf("publish AMQP job failure: %w", publishErr)
		}

		return delivery.received.Ack(false)
	})
}

// delayedRetry first publishes the next attempt to a durable TTL queue and waits
// for a publisher confirm, then acknowledges the original. A process or connection
// failure after the confirmed publish and before the acknowledgement can duplicate
// work; reversing that order could lose work, so this adapter deliberately favors
// at-least-once delivery.
func (delivery *delivery) delayedRetry(ctx context.Context, delay time.Duration) error {
	queue, err := delivery.driver.queue(delivery.message.Queue)
	if err != nil {
		return err
	}

	body, err := json.Marshal(delivery.message)
	if err != nil {
		return fmt.Errorf("marshal retry job message: %w", err)
	}

	return delivery.driver.publishDelayed(ctx, queue, delivery.message.ID, body, delivery.message.Attempts+1, delay, delivery.received.Headers)
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

func (driver *Driver) publishDelayed(ctx context.Context, queue QueueConfig, id string, body []byte, attempt int, delay time.Duration, headers amqp091.Table) error {
	retryQueue := retryQueueName(queue.Name, delay)
	arguments := retryQueueArguments(driver.config.Exchange, queue.RoutingKey)

	driver.mu.Lock()
	if driver.closed {
		driver.mu.Unlock()

		return ErrDriverClosed
	}
	if _, err := driver.publisher.QueueDeclare(retryQueue, true, false, false, false, arguments); err != nil {
		driver.mu.Unlock()

		return fmt.Errorf("declare AMQP retry queue: %w", err)
	}
	message := delayedPublishing(id, body, attempt, delay, headers)
	sequence := driver.publisher.GetNextPublishSeqNo()
	if err := driver.publisher.PublishWithContext(ctx, "", retryQueue, false, false, message); err != nil {
		driver.mu.Unlock()

		return fmt.Errorf("publish delayed AMQP job: %w", err)
	}

	err := waitConfirm(ctx, driver.confirms, sequence)
	driver.mu.Unlock()

	return err
}

// waitConfirm ignores confirmations for earlier publishes whose caller stopped
// waiting. The publish channel is serialized, so the requested sequence identifies
// this publish without exposing AMQP delivery tags to job deliveries.
func waitConfirm(ctx context.Context, confirms <-chan amqp091.Confirmation, sequence uint64) error {
	for {
		select {
		case confirmation, ok := <-confirms:
			if !ok {
				return ErrPublishNotConfirmed
			}
			if confirmation.DeliveryTag != sequence {
				continue
			}
			if !confirmation.Ack {
				return ErrPublishNotConfirmed
			}

			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func retryQueueName(queue string, delay time.Duration) string {
	return queue + ".retry." + strconv.FormatInt(retryMilliseconds(delay), 10)
}

func retryQueueArguments(exchange, routingKey string) amqp091.Table {
	return amqp091.Table{
		"x-dead-letter-exchange":    exchange,
		"x-dead-letter-routing-key": routingKey,
	}
}

func delayedPublishing(id string, body []byte, attempt int, delay time.Duration, headers amqp091.Table) amqp091.Publishing {
	message := publishing(id, body, attempt)
	message.Headers = maps.Clone(headers)
	if message.Headers == nil {
		message.Headers = amqp091.Table{}
	}
	message.Headers[attemptHeader] = int32(attempt)
	message.Expiration = strconv.FormatInt(retryMilliseconds(delay), 10)

	return message
}

func retryMilliseconds(delay time.Duration) int64 {
	return int64(math.Ceil(float64(delay) / float64(time.Millisecond)))
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
