package jetstream

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json/v2"
	"errors"
	"fmt"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/studiolambda/cosmos/contract"

	"github.com/nats-io/nats.go"
)

var (
	// ErrInvalidConfig is returned when a JetStream job driver configuration is invalid.
	ErrInvalidConfig = errors.New("invalid JetStream job driver configuration")

	// ErrUnknownQueue is returned when a queue has no configured JetStream mapping.
	ErrUnknownQueue = errors.New("unknown job queue")

	// ErrResourcesNotProvisioned is returned when required JetStream resources do not exist.
	ErrResourcesNotProvisioned = errors.New("JetStream job resources are not provisioned")

	// ErrInvalidHandler is returned when Consume is given a nil handler.
	ErrInvalidHandler = errors.New("invalid job delivery handler")

	// ErrInvalidDelivery is returned when a JetStream message cannot be converted to a job delivery.
	ErrInvalidDelivery = errors.New("invalid JetStream job delivery")

	// ErrUnsupportedRetryDelay is returned when a retry delay is negative.
	ErrUnsupportedRetryDelay = errors.New("unsupported JetStream job retry delay")

	// ErrDeliverySettled is returned when a delivery is settled more than once.
	ErrDeliverySettled = errors.New("job delivery already settled")
)

// QueueConfig maps one application queue to a JetStream subject and durable pull consumer.
type QueueConfig struct {
	// Subject receives dispatched jobs for this queue.
	Subject string

	// Durable is the durable pull consumer name for this queue.
	Durable string
}

// Config configures a JetStream-backed job driver.
type Config struct {
	// URLs are the NATS server URLs used by [New]. At least one URL is required.
	URLs []string

	// Options are passed to nats.Connect by [New].
	Options []nats.Option

	// Stream is the JetStream stream containing all configured queue subjects.
	Stream string

	// Queues maps application queue names to subjects and durable pull consumers.
	Queues map[string]QueueConfig

	// FailureSubject receives application-level terminal failure envelopes. An
	// empty value logs failures and acknowledges terminal deliveries.
	FailureSubject string

	// Provision creates missing stream and durable consumers. When false,
	// resources must already exist and are validated by New.
	Provision bool

	// Batch is the number of messages each pull request asks for. Zero uses one.
	Batch int

	// MaxAckPending caps unacknowledged messages for each durable consumer. Zero uses Batch.
	MaxAckPending int

	// AckWait is the JetStream acknowledgement deadline. Zero uses 30 seconds.
	AckWait time.Duration

	// MaxDeliver caps deliveries for provisioned consumers. Zero leaves the
	// JetStream server default in effect.
	MaxDeliver int

	// Logger records delivery failures. A nil logger discards records.
	Logger *contract.Logger
}

// DefaultConfig returns JetStream job driver defaults. Callers must set Stream and Queues.
func DefaultConfig() Config {
	return Config{Batch: 1, AckWait: 30 * time.Second}
}

// FromConfiguration populates the declarative driver configuration from configuration.
func (config *Config) FromConfiguration(configuration *contract.Configuration) {
	options := config.Options
	logger := config.Logger
	*config = DefaultConfig()
	config.Options = options
	config.Logger = logger
	config.URLs = configuration.GetOr("urls", config.URLs)
	config.Stream = configuration.GetOr("stream", config.Stream)
	config.Queues = configuration.GetOr("queues", config.Queues)
	config.FailureSubject = configuration.GetOr("failure_subject", config.FailureSubject)
	config.Provision = configuration.GetOr("provision", config.Provision)
	config.Batch = configuration.GetOr("batch", config.Batch)
	config.MaxAckPending = configuration.GetOr("max_ack_pending", config.MaxAckPending)
	config.AckWait = configuration.GetOr("ack_wait", config.AckWait)
	config.MaxDeliver = configuration.GetOr("max_deliver", config.MaxDeliver)
}

type pullJetStream interface {
	nats.JetStreamContext
	PullSubscribe(string, string, ...nats.SubOpt) (*nats.Subscription, error)
}

// Driver implements [contract.JobDispatcherDriver], [contract.JobConsumerDriver], and [contract.Pinger].
// It uses durable JetStream pull consumers, so concurrent workers safely compete for one queue's messages.
type Driver struct {
	stream         string
	queues         map[string]QueueConfig
	failureSubject string
	batch          int
	jetStream      pullJetStream
	connection     *nats.Conn
	owned          bool
	mu             sync.Mutex
	closed         bool
	logger         *contract.Logger
}

// New connects to NATS, provisions or validates JetStream resources, and
// returns a driver that owns the new connection.
func New(config Config) (*Driver, error) {
	config, err := normalizeConfig(config)
	if err != nil {
		return nil, err
	}

	if len(config.URLs) == 0 {
		return nil, fmt.Errorf("%w: NATS URLs are required", ErrInvalidConfig)
	}

	connection, err := nats.Connect(strings.Join(config.URLs, ","), config.Options...)
	if err != nil {
		return nil, fmt.Errorf("connect to NATS: %w", err)
	}

	driver, err := newFrom(connection, config, true)
	if err != nil {
		connection.Close()

		return nil, err
	}

	return driver, nil
}

// NewFrom configures a JetStream job driver using conn. The caller retains
// ownership of conn and must keep it open for the driver's lifetime.
func NewFrom(conn *nats.Conn, config Config) (*Driver, error) {
	if conn == nil {
		return nil, fmt.Errorf("%w: nil NATS connection", ErrInvalidConfig)
	}

	return newFrom(conn, config, false)
}

func newFrom(conn *nats.Conn, config Config, owned bool) (*Driver, error) {
	config, err := normalizeConfig(config)
	if err != nil {
		return nil, err
	}

	jetStreamContext, err := conn.JetStream()
	if err != nil {
		return nil, fmt.Errorf("create JetStream context: %w", err)
	}

	jetStream, ok := jetStreamContext.(pullJetStream)
	if !ok {
		return nil, fmt.Errorf("%w: installed NATS client does not support pull consumers", ErrInvalidConfig)
	}

	if err := ensureResources(jetStream, config); err != nil {
		return nil, err
	}

	return &Driver{
		stream:         config.Stream,
		queues:         config.Queues,
		failureSubject: config.FailureSubject,
		batch:          config.Batch,
		jetStream:      jetStream,
		connection:     conn,
		owned:          owned,
		logger:         driverLogger(config.Logger),
	}, nil
}

// Dispatch JSON-encodes message and publishes it to its configured JetStream subject.
// A non-empty message ID is sent as Nats-Msg-Id for JetStream deduplication.
func (driver *Driver) Dispatch(ctx context.Context, message contract.JobMessage) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	queue, err := driver.queue(message.Queue)
	if err != nil {
		return err
	}
	if message.ID == "" {
		message.ID, err = newMessageID()
		if err != nil {
			return err
		}
	}

	body, err := marshalMessage(message)
	if err != nil {
		return err
	}

	publication := nats.NewMsg(queue.Subject)
	publication.Data = body
	if message.ID != "" {
		publication.Header.Set(nats.MsgIdHdr, message.ID)
	}

	_, err = driver.jetStream.PublishMsg(publication, nats.Context(ctx))
	if err != nil {
		return fmt.Errorf("publish JetStream job message: %w", err)
	}

	return nil
}

// Consume receives jobs from queue until ctx is canceled or JetStream returns an error.
// Handler errors are logged and deliberately left unacknowledged for redelivery.
func (driver *Driver) Consume(ctx context.Context, queue string, handler contract.JobDeliveryHandler) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	if handler == nil {
		return ErrInvalidHandler
	}

	configuredQueue, err := driver.queue(queue)
	if err != nil {
		return err
	}

	subscription, err := driver.jetStream.PullSubscribe(configuredQueue.Subject, configuredQueue.Durable, nats.Bind(driver.stream, configuredQueue.Durable))
	if err != nil {
		return fmt.Errorf("bind JetStream pull consumer: %w", err)
	}
	defer subscription.Unsubscribe()

	for {
		messages, err := subscription.Fetch(driver.batch, nats.Context(ctx))
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}

			return fmt.Errorf("fetch JetStream job messages: %w", err)
		}

		for _, received := range messages {
			driver.handle(ctx, queue, received, handler)
		}
	}
}

// Ping verifies that the NATS connection and JetStream API are available.
func (driver *Driver) Ping(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	_, err := driver.jetStream.AccountInfo(nats.Context(ctx))
	if err != nil {
		return fmt.Errorf("ping JetStream: %w", err)
	}

	return nil
}

// Close drains the NATS connection when the driver created it. Drivers created
// with [NewFrom] leave the caller-owned connection open. Drain is NATS's
// graceful shutdown operation and closes the connection after pending work.
func (driver *Driver) Close() error {
	driver.mu.Lock()
	defer driver.mu.Unlock()

	if driver.closed {
		return nil
	}
	driver.closed = true

	if !driver.owned {
		return nil
	}

	if err := driver.connection.Drain(); err != nil {
		return fmt.Errorf("drain NATS connection: %w", err)
	}

	return nil
}

func (driver *Driver) handle(ctx context.Context, queue string, received *nats.Msg, handler contract.JobDeliveryHandler) {
	message, err := decodeMessage(queue, received)
	if err != nil {
		delivery := newDelivery(driver, received, contract.JobMessage{})
		if delivery == nil {
			driver.logger.Error("invalid JetStream job delivery", "error", err, "queue", queue)

			return
		}
		delivery.rawMessage = string(received.Data)

		if failureErr := delivery.Fail(ctx, err); failureErr != nil {
			driver.logger.Error("failed invalid JetStream job message", "error", failureErr, "queue", queue)
		}

		return
	}

	delivery := newDelivery(driver, received, message)
	if delivery == nil {
		driver.logger.Error("invalid JetStream job delivery", "error", ErrInvalidDelivery, "queue", queue, "job_id", message.ID)

		return
	}

	if err := handler(ctx, delivery); err != nil {
		driver.logger.Error("JetStream job delivery handler failed", "error", err, "queue", queue, "job_id", message.ID)
	}
}

func (driver *Driver) queue(name string) (QueueConfig, error) {
	queue, ok := driver.queues[name]
	if !ok {
		return QueueConfig{}, fmt.Errorf("%w: %s", ErrUnknownQueue, name)
	}

	return queue, nil
}

func normalizeConfig(config Config) (Config, error) {
	if strings.TrimSpace(config.Stream) == "" || len(config.Queues) == 0 {
		return Config{}, fmt.Errorf("%w: stream and queues are required", ErrInvalidConfig)
	}

	if config.Batch == 0 {
		config.Batch = 1
	}
	if config.Batch < 1 {
		return Config{}, fmt.Errorf("%w: batch must be positive", ErrInvalidConfig)
	}

	if config.MaxAckPending == 0 {
		config.MaxAckPending = config.Batch
	}
	if config.MaxAckPending < config.Batch {
		return Config{}, fmt.Errorf("%w: max ack pending must be at least batch", ErrInvalidConfig)
	}

	if config.AckWait == 0 {
		config.AckWait = 30 * time.Second
	}
	if config.AckWait <= 0 || config.MaxDeliver < 0 {
		return Config{}, fmt.Errorf("%w: acknowledgement settings are invalid", ErrInvalidConfig)
	}

	queues := make(map[string]QueueConfig, len(config.Queues))
	subjects := make(map[string]struct{}, len(config.Queues))
	durables := make(map[string]struct{}, len(config.Queues))
	for name, queue := range config.Queues {
		if strings.TrimSpace(name) == "" || strings.TrimSpace(queue.Subject) == "" || strings.TrimSpace(queue.Durable) == "" {
			return Config{}, fmt.Errorf("%w: queue name, subject, and durable are required", ErrInvalidConfig)
		}
		if _, ok := subjects[queue.Subject]; ok {
			return Config{}, fmt.Errorf("%w: duplicate queue subject %q", ErrInvalidConfig, queue.Subject)
		}
		if _, ok := durables[queue.Durable]; ok {
			return Config{}, fmt.Errorf("%w: duplicate durable consumer %q", ErrInvalidConfig, queue.Durable)
		}

		queues[name] = queue
		subjects[queue.Subject] = struct{}{}
		durables[queue.Durable] = struct{}{}
	}
	config.Queues = queues

	return config, nil
}

func ensureResources(jetStream nats.JetStreamContext, config Config) error {
	if err := ensureStream(jetStream, config); err != nil {
		return err
	}

	for _, queue := range config.Queues {
		if err := ensureConsumer(jetStream, config, queue); err != nil {
			return err
		}
	}

	return nil
}

func ensureStream(jetStream nats.JetStreamContext, config Config) error {
	stream, err := jetStream.StreamInfo(config.Stream)
	if err == nil {
		return validateStream(config, stream.Config)
	}
	if !errors.Is(err, nats.ErrStreamNotFound) || !config.Provision {
		return fmt.Errorf("%w: stream %q: %w", ErrResourcesNotProvisioned, config.Stream, err)
	}

	subjects := make([]string, 0, len(config.Queues))
	for _, queue := range config.Queues {
		subjects = append(subjects, queue.Subject)
	}

	_, err = jetStream.AddStream(&nats.StreamConfig{Name: config.Stream, Subjects: subjects, Retention: nats.WorkQueuePolicy})
	if err != nil {
		return fmt.Errorf("provision JetStream stream %q: %w", config.Stream, err)
	}

	return nil
}

func validateStream(config Config, stream nats.StreamConfig) error {
	subjects := make(map[string]struct{}, len(stream.Subjects))
	for _, subject := range stream.Subjects {
		subjects[subject] = struct{}{}
	}

	for _, queue := range config.Queues {
		if _, ok := subjects[queue.Subject]; !ok {
			return fmt.Errorf("%w: stream %q does not include subject %q", ErrResourcesNotProvisioned, config.Stream, queue.Subject)
		}
	}

	return nil
}

func ensureConsumer(jetStream nats.JetStreamContext, config Config, queue QueueConfig) error {
	consumerInfo, err := jetStream.ConsumerInfo(config.Stream, queue.Durable)
	if err == nil {
		return validateConsumer(config, queue, consumerInfo.Config)
	}
	if !errors.Is(err, nats.ErrConsumerNotFound) || !config.Provision {
		return fmt.Errorf("%w: consumer %q: %w", ErrResourcesNotProvisioned, queue.Durable, err)
	}

	consumer := &nats.ConsumerConfig{
		Durable:           queue.Durable,
		DeliverPolicy:     nats.DeliverAllPolicy,
		AckPolicy:         nats.AckExplicitPolicy,
		AckWait:           config.AckWait,
		MaxDeliver:        config.MaxDeliver,
		FilterSubject:     queue.Subject,
		MaxAckPending:     config.MaxAckPending,
		MaxRequestBatch:   config.Batch,
		MaxRequestExpires: config.AckWait,
	}

	_, err = jetStream.AddConsumer(config.Stream, consumer)
	if err != nil {
		return fmt.Errorf("provision JetStream consumer %q: %w", queue.Durable, err)
	}

	return nil
}

func validateConsumer(config Config, queue QueueConfig, consumer nats.ConsumerConfig) error {
	if consumer.Durable != queue.Durable || consumer.DeliverSubject != "" || consumer.AckPolicy != nats.AckExplicitPolicy || consumer.FilterSubject != queue.Subject {
		return fmt.Errorf("%w: consumer %q does not match the required durable pull configuration", ErrResourcesNotProvisioned, queue.Durable)
	}
	if consumer.AckWait != config.AckWait || consumer.MaxAckPending != config.MaxAckPending || consumer.MaxRequestBatch < config.Batch {
		return fmt.Errorf("%w: consumer %q does not match the required acknowledgement or batch configuration", ErrResourcesNotProvisioned, queue.Durable)
	}
	if config.MaxDeliver > 0 && consumer.MaxDeliver != config.MaxDeliver {
		return fmt.Errorf("%w: consumer %q does not match the required maximum deliveries", ErrResourcesNotProvisioned, queue.Durable)
	}

	return nil
}

func marshalMessage(message contract.JobMessage) ([]byte, error) {
	body, err := json.Marshal(message)
	if err != nil {
		return nil, fmt.Errorf("marshal job message: %w", err)
	}

	return body, nil
}

// newMessageID generates an opaque, transport-local id for JetStream deduplication.
func newMessageID() (string, error) {
	value := make([]byte, 16)
	if _, err := rand.Read(value); err != nil {
		return "", fmt.Errorf("generate job message ID: %w", err)
	}

	return base64.RawURLEncoding.EncodeToString(value), nil
}

func decodeMessage(queue string, received *nats.Msg) (contract.JobMessage, error) {
	if received == nil {
		return contract.JobMessage{}, fmt.Errorf("%w: nil message", ErrInvalidDelivery)
	}

	var message contract.JobMessage
	if err := json.Unmarshal(received.Data, &message); err != nil {
		return contract.JobMessage{}, fmt.Errorf("%w: unmarshal job message: %w", ErrInvalidDelivery, err)
	}

	attempts, err := deliveryAttempts(received)
	if err != nil {
		return contract.JobMessage{}, err
	}

	message.Queue = queue
	message.Attempts = attempts
	if message.ID == "" {
		message.ID = received.Header.Get(nats.MsgIdHdr)
	}

	return message, nil
}

func deliveryAttempts(received *nats.Msg) (int, error) {
	metadata, err := received.Metadata()
	if err != nil {
		return 0, fmt.Errorf("%w: JetStream metadata: %w", ErrInvalidDelivery, err)
	}
	if metadata.NumDelivered > uint64(^uint(0)>>1) {
		return 0, fmt.Errorf("%w: delivery attempt overflow", ErrInvalidDelivery)
	}

	return int(metadata.NumDelivered), nil
}

type delivery struct {
	driver     *Driver
	received   *nats.Msg
	message    contract.JobMessage
	rawMessage string
	mu         sync.Mutex
	settling   bool
	settled    bool
}

func newDelivery(driver *Driver, received *nats.Msg, message contract.JobMessage) *delivery {
	if received == nil {
		return nil
	}

	return &delivery{driver: driver, received: received, message: message}
}

func (delivery *delivery) Message() contract.JobMessage {
	message := delivery.message
	message.Payload = slices.Clone(message.Payload)

	return message
}

func (delivery *delivery) Acknowledge(ctx context.Context) error {
	return delivery.settle(ctx, func() error { return delivery.received.Ack(nats.Context(ctx)) })
}

func (delivery *delivery) Retry(ctx context.Context, delay time.Duration) error {
	if delay < 0 {
		return fmt.Errorf("%w: %s", ErrUnsupportedRetryDelay, delay)
	}

	return delivery.settle(ctx, func() error {
		if delay == 0 {
			return delivery.received.Nak(nats.Context(ctx))
		}

		return delivery.received.NakWithDelay(delay, nats.Context(ctx))
	})
}

func (delivery *delivery) Reject(ctx context.Context) error {
	return delivery.Acknowledge(ctx)
}

func (delivery *delivery) Fail(ctx context.Context, err error) error {
	return delivery.settle(ctx, func() error {
		if err := delivery.publishFailure(ctx, err); err != nil {
			return err
		}

		return delivery.received.Ack(nats.Context(ctx))
	})
}

func (delivery *delivery) publishFailure(ctx context.Context, err error) error {
	if delivery.driver.failureSubject == "" {
		delivery.driver.logger.Error("JetStream job delivery failed", "error", err, "queue", delivery.message.Queue, "job_id", delivery.message.ID)

		return nil
	}

	body, marshalErr := json.Marshal(failureMessage{Message: delivery.message, Error: errorText(err), RawMessage: delivery.rawMessage})
	if marshalErr != nil {
		return fmt.Errorf("marshal job failure: %w", marshalErr)
	}

	publication := nats.NewMsg(delivery.driver.failureSubject)
	publication.Data = body
	if delivery.message.ID != "" {
		publication.Header.Set(nats.MsgIdHdr, delivery.message.ID+":failure")
	}

	_, publishErr := delivery.driver.jetStream.PublishMsg(publication, nats.Context(ctx))
	if publishErr != nil {
		return fmt.Errorf("publish job failure: %w", publishErr)
	}

	return nil
}

func driverLogger(logger *contract.Logger) *contract.Logger {
	if logger == nil {
		return contract.NewLogger(nil)
	}

	return logger
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
