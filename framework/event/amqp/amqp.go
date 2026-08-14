package amqp

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/studiolambda/cosmos/contract"
	core "github.com/studiolambda/cosmos/framework/event/internal/event"

	amqp091 "github.com/rabbitmq/amqp091-go"
)

// AMQPBroker implements [contract.EventPublisherDriver] and
// [contract.EventSubscriberDriver] using RabbitMQ's
// AMQP protocol for publish/subscribe messaging. It uses a topic
// exchange to broadcast events to all subscribed consumers, with
// each subscriber receiving messages in their own exclusive queue.
//
// The broker maintains a single connection with one channel for
// publishing and creates individual channels for each subscriber,
// following RabbitMQ best practices for concurrent access.
//
// Wildcard patterns: '*' matches a single dot-separated word and '#'
// matches zero or more words, following AMQP topic exchange semantics.
type AMQPBroker struct {
	// conn is the shared AMQP connection used for all operations.
	conn *amqp091.Connection

	// pubCh is the dedicated channel for publishing messages.
	pubCh *amqp091.Channel

	// exchange is the name of the topic exchange where events are published.
	exchange string

	// owned reports whether Close must close conn.
	owned bool

	// mu protects concurrent access to the publish channel.
	mu sync.Mutex

	logger *contract.Logger
}

// AMQPBrokerConfig configures the creation of a new AMQPBroker,
// allowing customization of the connection URL and exchange name.
//
// WARNING: The URL field typically contains credentials in the
// format amqp://username:password@host:port/vhost. These are
// stored as a plain string in memory for the lifetime of this
// struct. Callers should:
//  1. Always use TLS (amqps://) to protect credentials in
//     transit.
//  2. Load the connection URL from environment variables or a
//     secret manager rather than hard-coding it.
//  3. Consider short-lived credentials or external auth
//     mechanisms where the broker supports them.
type AMQPBrokerConfig struct {
	// URL is the AMQP connection string in the format:
	// amqp://username:password@host:port/vhost
	URL string

	// Exchange is the name of the topic exchange to use for events.
	// If empty, DefaultAMQPExchange is used.
	Exchange string

	// Logger records recovered handler panics. A nil logger discards records.
	Logger *contract.Logger
}

// DefaultAMQPExchange is the default name for the topic exchange
// used by AMQPBroker when no custom exchange is specified.
const DefaultAMQPExchange = "cosmos:events"

// DefaultAMQPBrokerConfig returns the default AMQP broker configuration.
func DefaultAMQPBrokerConfig() AMQPBrokerConfig {
	return AMQPBrokerConfig{Exchange: DefaultAMQPExchange}
}

// NewAMQPBroker creates a new AMQPBroker using the provided
// configuration for connection URL and exchange name. If no exchange
// name is specified in the configuration, DefaultAMQPExchange is used.
// The broker must be closed when no longer needed to release
// the connection and associated resources.
func NewAMQPBroker(config AMQPBrokerConfig) (*AMQPBroker, error) {
	conn, err := amqp091.Dial(config.URL)
	if err != nil {
		return nil, err
	}

	exchange := config.Exchange
	if exchange == "" {
		exchange = DefaultAMQPExchange
	}

	broker, err := newAMQPBrokerFrom(conn, AMQPBrokerConfig{Exchange: exchange}, true)
	if err != nil {
		_ = conn.Close()

		return nil, err
	}

	return broker, nil
}

// NewAMQPBrokerFrom creates a new AMQPBroker using an existing
// AMQP connection and broker configuration. This constructor is useful
// when you need to share a connection across multiple brokers
// or have custom connection configuration requirements.
//
// The function creates a dedicated channel for publishing and
// declares the topic exchange. If the exchange already exists
// with matching configuration, the declaration is idempotent.
// The caller retains ownership of conn.
func NewAMQPBrokerFrom(
	conn *amqp091.Connection,
	config AMQPBrokerConfig,
) (*AMQPBroker, error) {
	return newAMQPBrokerFrom(conn, config, false)
}

func newAMQPBrokerFrom(conn *amqp091.Connection, config AMQPBrokerConfig, owned bool) (*AMQPBroker, error) {
	if conn == nil {
		return nil, errors.New("nil AMQP connection")
	}

	exchange := config.Exchange
	if exchange == "" {
		exchange = DefaultAMQPExchange
	}

	pubCh, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	err = pubCh.ExchangeDeclare(
		exchange,
		"topic",
		true,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		// Close is best-effort: the connection is being abandoned due to
		// the exchange declaration failure above.
		_ = pubCh.Close()

		return nil, err
	}

	return &AMQPBroker{
		conn:     conn,
		pubCh:    pubCh,
		exchange: exchange,
		owned:    owned,
		logger:   brokerLogger(config.Logger),
	}, nil
}

// Publish sends raw payload bytes to all subscribers of the named event.
func (broker *AMQPBroker) Publish(
	ctx context.Context,
	event string,
	payload []byte,
) error {
	if err := core.ValidateName(event); err != nil {
		return err
	}

	broker.mu.Lock()
	defer broker.mu.Unlock()

	err := broker.pubCh.PublishWithContext(
		ctx,
		broker.exchange,
		event,
		false,
		false,
		amqp091.Publishing{
			ContentType: "application/octet-stream",
			Body:        payload,
		},
	)

	if err != nil {
		newCh, chErr := broker.conn.Channel()
		if chErr != nil {
			return errors.Join(err, chErr)
		}

		broker.pubCh = newCh

		return broker.pubCh.PublishWithContext(
			ctx,
			broker.exchange,
			event,
			false,
			false,
			amqp091.Publishing{
				ContentType: "application/octet-stream",
				Body:        payload,
			},
		)
	}

	return nil
}

// Subscribe registers a handler to receive events with the given
// name. Each subscription creates its own exclusive, auto-delete
// queue that is bound to the topic exchange with the event name
// as the routing key. Messages are automatically acknowledged.
//
// The handler receives messages in a separate goroutine and will
// continue processing until the context is cancelled or the
// returned unsubscribe function is called. The handler receives
// raw payload bytes, which callers can decode with json.Unmarshal
// (or via [contract.NewEventSubscriber] for typed decoding).
//
// If subscription setup fails, the returned unsubscribe function
// will return the setup error when called.
func (broker *AMQPBroker) Subscribe(
	ctx context.Context,
	event string,
	handler contract.EventHandler,
) (contract.EventUnsubscribeFunc, error) {
	if err := core.ValidatePattern(event); err != nil {
		return nil, err
	}

	ch, err := broker.conn.Channel()
	if err != nil {
		return nil, err
	}

	queue, err := ch.QueueDeclare(
		"",
		false,
		true,
		true,
		false,
		nil,
	)
	if err != nil {
		// Close is best-effort: the channel is being abandoned due to
		// the queue declaration failure above.
		_ = ch.Close()

		return nil, err
	}

	err = ch.QueueBind(
		queue.Name,
		event,
		broker.exchange,
		false,
		nil,
	)

	if err != nil {
		// Close is best-effort: the channel is being abandoned due to
		// the queue bind failure above.
		_ = ch.Close()

		return nil, err
	}

	deliveries, err := ch.ConsumeWithContext(
		ctx,
		queue.Name,
		"",
		true,
		true,
		false,
		false,
		nil,
	)
	if err != nil {
		// Close is best-effort: the channel is being abandoned due to
		// the consume setup failure above.
		_ = ch.Close()

		return nil, err
	}

	var wg sync.WaitGroup

	wg.Go(func() {
		for delivery := range deliveries {
			func() {
				defer func() {
					if r := recover(); r != nil {
						broker.logger.Error("panic in amqp event handler", "event", event, "panic", fmt.Sprint(r))
					}
				}()

				handler(delivery.Body)
			}()
		}
	})

	return func() error {
		defer wg.Wait()

		return ch.Close()
	}, nil
}

func brokerLogger(logger *contract.Logger) *contract.Logger {
	if logger == nil {
		return contract.NewLogger(nil)
	}

	return logger
}

// Ping verifies that the AMQP connection is still alive.
func (broker *AMQPBroker) Ping(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	ch, err := broker.conn.Channel()
	if err != nil {
		return err
	}

	return ch.Close()
}

// Close closes the broker's publish channel. It also closes the AMQP connection
// when the broker created it, which closes active subscriber channels.
//
// If closing the publish channel fails, the connection is still
// closed and the channel close error is returned.
func (broker *AMQPBroker) Close() error {
	if broker.pubCh != nil {
		if err := broker.pubCh.Close(); err != nil {
			if broker.owned {
				_ = broker.conn.Close()
			}

			return err
		}
	}

	if broker.owned {
		return broker.conn.Close()
	}

	return nil
}
