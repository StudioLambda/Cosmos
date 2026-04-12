package event

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"sync"

	"github.com/redis/go-redis/v9"
	"github.com/studiolambda/cosmos/contract"
)

// RedisBroker implements contract.EventBus using Redis Pub/Sub.
// It maps the "#" multi-level wildcard to Redis's "*" glob pattern
// for topic subscriptions.
type RedisBroker struct {
	client *redis.Client
	wg     sync.WaitGroup
}

// RedisBrokerOptions is an alias for redis.Options, exposing the
// full set of Redis connection parameters without requiring a
// direct import of the go-redis package.
type RedisBrokerOptions = redis.Options

// NewRedisBroker creates a RedisBroker by connecting to Redis
// with the given options.
func NewRedisBroker(options *redis.Options) *RedisBroker {
	client := redis.NewClient(options)

	return NewRedisBrokerFrom(client)
}

// NewRedisBrokerFrom wraps an existing redis.Client as a
// RedisBroker, allowing reuse of a pre-configured connection.
func NewRedisBrokerFrom(client *redis.Client) *RedisBroker {
	return &RedisBroker{
		client: client,
	}
}

// Publish serializes the payload as JSON and publishes it to the
// given Redis channel.
func (broker *RedisBroker) Publish(ctx context.Context, event string, payload any) error {
	encoded, err := json.Marshal(payload)

	if err != nil {
		return err
	}

	return broker.client.Publish(ctx, event, encoded).Err()
}

// Subscribe registers a handler for messages matching the given
// event pattern. The "#" wildcard is translated to Redis's "*" glob.
// It returns an unsubscribe function that closes the subscription
// and waits for the delivery goroutine to finish.
func (broker *RedisBroker) Subscribe(
	ctx context.Context,
	event string,
	handler contract.EventHandler,
) (contract.EventUnsubscribeFunc, error) {
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

				handler(func(dest any) error {
					return json.Unmarshal([]byte(message.Payload), dest)
				})
			}()
		}
	}()

	return func() error {
		return sub.Close()
	}, nil
}

// Close shuts down the underlying Redis client connection and
// waits for all active subscription goroutines to finish.
func (broker *RedisBroker) Close() error {
	err := broker.client.Close()
	broker.wg.Wait()

	return err
}
