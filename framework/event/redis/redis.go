package redis

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"

	"github.com/studiolambda/cosmos/contract"
	core "github.com/studiolambda/cosmos/framework/event/internal/event"

	"github.com/redis/go-redis/v9"
)

// RedisBroker implements [contract.EventPublisherDriver] and
// [contract.EventSubscriberDriver] using Redis Pub/Sub.
// It uses Redis pattern subscriptions and verifies each delivery against the
// portable event-pattern grammar.
type RedisBroker struct {
	client *redis.Client
	wg     sync.WaitGroup
}

// RedisBrokerConfig configures the Redis pub/sub event broker.
type RedisBrokerConfig struct {
	// Network is the network type passed to the Redis client.
	// Common values are "tcp" and "unix".
	Network string

	// Addr is the Redis server address.
	Addr string

	// Username is the optional ACL username.
	Username string

	// Password is the optional ACL password.
	Password string

	// DB is the Redis logical database number.
	DB int
}

// DefaultRedisBrokerConfig holds the default Redis event broker configuration.
var DefaultRedisBrokerConfig = RedisBrokerConfig{
	Network: "tcp",
	Addr:    "localhost:6379",
}

// NewRedisBroker creates a RedisBroker by connecting to Redis
// with the given configuration.
func NewRedisBroker(config RedisBrokerConfig) *RedisBroker {
	config = config.withDefaults()

	client := redis.NewClient(&redis.Options{
		Network:  config.Network,
		Addr:     config.Addr,
		Username: config.Username,
		Password: config.Password,
		DB:       config.DB,
	})

	return NewRedisBrokerFrom(client)
}

func (config RedisBrokerConfig) withDefaults() RedisBrokerConfig {
	if config.Network == "" {
		config.Network = DefaultRedisBrokerConfig.Network
	}

	if config.Addr == "" {
		config.Addr = DefaultRedisBrokerConfig.Addr
	}

	return config
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
	if err := core.ValidateName(event); err != nil {
		return err
	}

	return broker.client.Publish(ctx, event, payload).Err()
}

// Subscribe registers a handler for messages matching the given event pattern.
func (broker *RedisBroker) Subscribe(
	ctx context.Context,
	event string,
	handler contract.EventHandler,
) (contract.EventUnsubscribeFunc, error) {
	if err := core.ValidatePattern(event); err != nil {
		return nil, err
	}

	pattern := event
	sub := broker.client.PSubscribe(ctx, redisPattern(pattern))

	broker.wg.Go(func() {

		for message := range sub.Channel() {
			if !core.Match(pattern, message.Channel) {
				continue
			}

			func() {
				defer func() {
					if r := recover(); r != nil {
						slog.Error("panic in redis event handler", "event", event, "panic", fmt.Sprint(r))
					}
				}()

				handler([]byte(message.Payload))
			}()
		}
	})

	return func() error {
		return sub.Close()
	}, nil
}

// redisPattern returns a Redis glob broad enough to receive every channel that
// may match pattern. Core matching filters Redis's broader glob semantics.
func redisPattern(pattern string) string {
	if pattern == "#" {
		return "*"
	}

	if !strings.HasSuffix(pattern, "#") {
		return pattern
	}

	return strings.TrimSuffix(pattern, ".#") + "*"
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

// Shutdown closes the Redis client unless ctx has already expired.
func (broker *RedisBroker) Shutdown(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	return broker.Close()
}
