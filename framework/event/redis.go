package event

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"

	"github.com/studiolambda/cosmos/contract"

	"github.com/redis/go-redis/v9"
)

// RedisBroker implements [contract.EventDriver] using Redis Pub/Sub.
// It maps the "#" multi-level wildcard to Redis's "*" glob pattern
// for topic subscriptions.
type RedisBroker struct {
	client *redis.Client
	wg     sync.WaitGroup
}

// RedisBrokerConfig is an alias for redis.Options, exposing the
// full set of Redis connection parameters without requiring a
// direct import of the go-redis package.
type RedisBrokerConfig = redis.Options

// NewRedisBroker creates a RedisBroker by connecting to Redis
// with the given configuration.
func NewRedisBroker(config *RedisBrokerConfig) *RedisBroker {
	client := redis.NewClient((*redis.Options)(config))

	return NewRedisBrokerFrom(client)
}

// NewRedisBrokerFrom wraps an existing redis.Client as a
// RedisBroker, allowing reuse of a pre-configured connection.
func NewRedisBrokerFrom(client *redis.Client) *RedisBroker {
	return &RedisBroker{
		client: client,
	}
}

// Publish sends raw payload bytes to the given Redis channel.
func (broker *RedisBroker) Publish(ctx context.Context, event string, payload []byte) error {
	if err := validateEvent(event); err != nil {
		return err
	}

	return broker.client.Publish(ctx, event, payload).Err()
}

// Subscribe registers a handler for messages matching the given
// event pattern. The "#" wildcard is translated to Redis's "*" glob.
func (broker *RedisBroker) Subscribe(
	ctx context.Context,
	event string,
	handler contract.EventHandler,
) (contract.EventUnsubscribeFunc, error) {
	if err := validateEvent(event); err != nil {
		return nil, err
	}

	event = strings.ReplaceAll(event, "#", "*")
	sub := broker.client.PSubscribe(ctx, event)

	broker.wg.Add(1)

	go func() {
		defer broker.wg.Done()

		for message := range sub.Channel() {
			func() {
				defer func() {
					if r := recover(); r != nil {
						slog.Error("panic in redis event handler", "event", event, "panic", fmt.Sprint(r))
					}
				}()

				handler([]byte(message.Payload))
			}()
		}
	}()

	return func() error {
		return sub.Close()
	}, nil
}

// Ping verifies that the Redis connection is still alive.
func (broker *RedisBroker) Ping(ctx context.Context) error {
	return broker.client.Ping(ctx).Err()
}

// Close shuts down the underlying Redis client connection and
// waits for all active subscription goroutines to finish.
func (broker *RedisBroker) Close() error {
	err := broker.client.Close()
	broker.wg.Wait()

	return err
}
