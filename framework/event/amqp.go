package event

import (
	"context"
	"encoding/json"
	"sync"

	amqp091 "github.com/rabbitmq/amqp091-go"
	"github.com/studiolambda/cosmos/contract"
)

// AMQPBroker implements the EventBroker interface using RabbitMQ's
// AMQP protocol for publish/subscribe messaging. It uses a topic
// exchange to broadcast events to all subscribed consumers, with
// each subscriber receiving messages in their own exclusive queue.
//
// The broker maintains a single connection with one channel for
// publishing and creates individual channels for each subscriber,
// following RabbitMQ best practices for concurrent access.
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

// AMQPBrokerOptions configures the creation of a new AMQPBroker,
// allowing customization of the connection URL and exchange name.
type AMQPBrokerOptions struct {
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
// options for connection URL and exchange name. If no exchange
// name is specified in the options, DefaultAMQPExchange is used.
// The broker must be closed when no longer needed to release
// the connection and associated resources.
func NewAMQPBrokerWith(options *AMQPBrokerOptions) (*AMQPBroker, error) {
	conn, err := amqp091.Dial(options.URL)
	if err != nil {
		return nil, err
	}

	exchange := options.Exchange
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
		pubCh.Close()

		return nil, err
	}

	return &AMQPBroker{
		conn:     conn,
		pubCh:    pubCh,
		exchange: exchange,
	}, nil
}

// Publish sends an event with the given name and payload to all
// subscribers listening for that event. The payload is serialized
// to JSON before being published to the topic exchange using the
// event name as the routing key.
//
// Publishing is thread-safe and respects context cancellation.
// If the context is cancelled before the publish completes, the
// operation will be aborted and an error returned.
func (b *AMQPBroker) Publish(
	ctx context.Context,
	event string,
	payload any,
) error {
	encoded, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	return b.pubCh.PublishWithContext(
		ctx,
		b.exchange,
		event,
		false,
		false,
		amqp091.Publishing{
			ContentType: "application/json",
			Body:        encoded,
		},
	)
}

// Subscribe registers a handler to receive events with the given
// name. Each subscription creates its own exclusive, auto-delete
// queue that is bound to the topic exchange with the event name
// as the routing key. Messages are automatically acknowledged.
//
// The handler receives messages in a separate goroutine and will
// continue processing until the context is cancelled or the
// returned unsubscribe function is called. The handler receives
// an EventPayload function that can unmarshal the JSON message
// into the desired type.
//
// If subscription setup fails, the returned unsubscribe function
// will return the setup error when called.
func (b *AMQPBroker) Subscribe(
	ctx context.Context,
	event string,
	handler contract.EventHandler,
) (contract.EventUnsubscribeFunc, error) {
	ch, err := b.conn.Channel()
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
		ch.Close()

		return nil, err
	}

	err = ch.QueueBind(
		queue.Name,
		event,
		b.exchange,
		false,
		nil,
	)

	if err != nil {
		ch.Close()

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
		ch.Close()

		return nil, err
	}

	wg := sync.WaitGroup{}

	wg.Go(func() {
		defer wg.Done()

		for delivery := range deliveries {
			handler(func(dest any) error {
				return json.Unmarshal(delivery.Body, dest)
			})
		}
	})

	return func() error {
		defer wg.Wait()

		return ch.Close()
	}, nil
}

// Close closes the broker's publish channel and the underlying
// AMQP connection, releasing all associated resources. This will
// also cause all active subscriber channels to be closed.
//
// If closing the publish channel fails, the connection is still
// closed and the channel close error is returned.
func (b *AMQPBroker) Close() error {
	if b.pubCh != nil {
		if err := b.pubCh.Close(); err != nil {
			b.conn.Close()

			return err
		}
	}

	return b.conn.Close()
}
