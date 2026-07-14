package event

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"

	"github.com/studiolambda/cosmos/contract"

	amqp091 "github.com/rabbitmq/amqp091-go"
)

// AMQPBroker implements the EventBroker interface using RabbitMQ's
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

	// mu protects concurrent access to the publish channel.
	mu sync.Mutex
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
}

// DefaultAMQPExchange is the default name for the topic exchange
// used by AMQPBroker when no custom exchange is specified.
const DefaultAMQPExchange = "cosmos.events"

// NewAMQPBroker creates a new AMQPBroker by establishing a connection
// to the RabbitMQ server at the given URL and using the default
// exchange name. The broker must be closed when no longer needed
// to release the connection and associated resources.
//
// The URL should be in the format:
// amqp://username:password@host:port/vhost
func NewAMQPBroker(url string) (*AMQPBroker, error) {
	conn, err := amqp091.Dial(url)
	if err != nil {
		return nil, err
	}

	return NewAMQPBrokerFrom(conn, DefaultAMQPExchange)
}

// NewAMQPBrokerWith creates a new AMQPBroker using the provided
// configuration for connection URL and exchange name. If no exchange
// name is specified in the configuration, DefaultAMQPExchange is used.
// The broker must be closed when no longer needed to release
// the connection and associated resources.
func NewAMQPBrokerWith(config *AMQPBrokerConfig) (*AMQPBroker, error) {
	conn, err := amqp091.Dial(config.URL)
	if err != nil {
		return nil, err
	}

	exchange := config.Exchange
	if exchange == "" {
		exchange = DefaultAMQPExchange
	}

	return NewAMQPBrokerFrom(conn, exchange)
}

// NewAMQPBrokerFrom creates a new AMQPBroker using an existing
// AMQP connection and exchange name. This constructor is useful
// when you need to share a connection across multiple brokers
// or have custom connection configuration requirements.
//
// The function creates a dedicated channel for publishing and
// declares the topic exchange. If the exchange already exists
// with matching configuration, the declaration is idempotent.
// The broker takes ownership of managing the connection lifecycle.
func NewAMQPBrokerFrom(
	conn *amqp091.Connection,
	exchange string,
) (*AMQPBroker, error) {
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
	}, nil
}

// Publish sends raw payload bytes to all subscribers of the named event.
func (broker *AMQPBroker) Publish(
	ctx context.Context,
	event string,
	payload []byte,
) error {
	if err := validateEvent(event); err != nil {
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
// (or via [contract.NewEvents] for typed decoding).
//
// If subscription setup fails, the returned unsubscribe function
// will return the setup error when called.
func (broker *AMQPBroker) Subscribe(
	ctx context.Context,
	event string,
	handler contract.EventHandler,
) (contract.EventUnsubscribeFunc, error) {
	if err := validateEvent(event); err != nil {
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
						slog.Error("panic in amqp event handler", "event", event, "panic", fmt.Sprint(r))
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

// Close closes the broker's publish channel and the underlying
// AMQP connection, releasing all associated resources. This will
// also cause all active subscriber channels to be closed.
//
// If closing the publish channel fails, the connection is still
// closed and the channel close error is returned.
func (broker *AMQPBroker) Close() error {
	if broker.pubCh != nil {
		if err := broker.pubCh.Close(); err != nil {
			broker.conn.Close()

			return err
		}
	}

	return broker.conn.Close()
}
