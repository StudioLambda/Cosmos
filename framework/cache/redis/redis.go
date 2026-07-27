// Package redis provides a Redis-backed [contract.CacheDriver].
package redis

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/studiolambda/cosmos/contract"

	"github.com/redis/go-redis/v9"
)

// RedisConfig configures the Redis cache driver.
type RedisConfig struct {
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

// RedisClient implements [contract.CacheDriver] and [contract.CacheCounter]
// using Redis as the backing store.
type RedisClient redis.Client

// DefaultRedisConfig holds the default Redis cache configuration.
var DefaultRedisConfig = RedisConfig{
	Network: "tcp",
	Addr:    "localhost:6379",
}

// NewRedis creates a RedisClient from the given connection configuration.
func NewRedis(config RedisConfig) *RedisClient {
	config = config.withDefaults()

	return NewRedisFrom(redis.NewClient(&redis.Options{
		Network:  config.Network,
		Addr:     config.Addr,
		Username: config.Username,
		Password: config.Password,
		DB:       config.DB,
	}))
}

func (config RedisConfig) withDefaults() RedisConfig {
	if config.Network == "" {
		config.Network = DefaultRedisConfig.Network
	}

	if config.Addr == "" {
		config.Addr = DefaultRedisConfig.Addr
	}

	return config
}

// NewRedisFrom wraps an existing redis.Client as a RedisClient,
// allowing reuse of a pre-configured connection.
func NewRedisFrom(client *redis.Client) *RedisClient {
	return (*RedisClient)(client)
}

// Get retrieves raw bytes by key. Returns
// [contract.ErrCacheKeyNotFound] when the key does not exist.
func (client *RedisClient) Get(ctx context.Context, key string) ([]byte, error) {
	value, err := (*redis.Client)(client).Get(ctx, key).Bytes()

	if errors.Is(err, redis.Nil) {
		return nil, contract.ErrCacheKeyNotFound
	}

	if err != nil {
		return nil, err
	}

	return value, nil
}

// Put stores raw bytes with the given TTL. A zero TTL means the key
// will not expire.
func (client *RedisClient) Put(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	return (*redis.Client)(client).Set(ctx, key, value, ttl).Err()
}

// Delete removes a key from Redis. Deleting a non-existent key is a no-op.
func (client *RedisClient) Delete(ctx context.Context, key string) error {
	return (*redis.Client)(client).Del(ctx, key).Err()
}

// Has reports whether the key exists in Redis and has not expired.
func (client *RedisClient) Has(ctx context.Context, key string) (bool, error) {
	count, err := (*redis.Client)(client).Exists(ctx, key).Result()

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// Add stores raw bytes only if the key does not already exist.
func (client *RedisClient) Add(ctx context.Context, key string, value []byte, ttl time.Duration) (bool, error) {
	return (*redis.Client)(client).SetNX(ctx, key, value, ttl).Result()
}

// Increment atomically increases the integer value stored at key by
// the given amount. Redis auto-creates the key with value 0 if it
// does not exist before incrementing.
func (client *RedisClient) Increment(ctx context.Context, key string, delta int64) (int64, error) {
	return (*redis.Client)(client).IncrBy(ctx, key, delta).Result()
}

// IncrementWithTTL atomically increments key and ensures it expires after ttl.
// A counter recreated after expiry receives a new expiration in the same Redis
// operation, preventing permanent rate-limit entries.
func (client *RedisClient) IncrementWithTTL(ctx context.Context, key string, delta int64, ttl time.Duration) (int64, time.Duration, error) {
	if ttl <= 0 {
		return 0, 0, fmt.Errorf("redis: counter TTL must be positive")
	}

	result, err := (*redis.Client)(client).Eval(ctx, `
local remaining = redis.call('PTTL', KEYS[1])
if remaining < 0 then
  redis.call('SET', KEYS[1], '0', 'PX', ARGV[1])
end
local count = redis.call('INCRBY', KEYS[1], ARGV[2])
remaining = redis.call('PTTL', KEYS[1])
return {count, remaining}
`, []string{key}, ttl.Milliseconds(), delta).Int64Slice()

	if err != nil {
		return 0, 0, err
	}

	return result[0], time.Duration(result[1]) * time.Millisecond, nil
}

// Decrement atomically decreases the integer value stored at key by
// the given amount. Redis auto-creates the key with value 0 if it
// does not exist before decrementing.
func (client *RedisClient) Decrement(ctx context.Context, key string, delta int64) (int64, error) {
	return (*redis.Client)(client).DecrBy(ctx, key, delta).Result()
}

// TTL returns the remaining lifetime for key.
func (client *RedisClient) TTL(ctx context.Context, key string) (time.Duration, error) {
	ttl, err := (*redis.Client)(client).TTL(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return 0, contract.ErrCacheKeyNotFound
	}

	if err != nil {
		return 0, err
	}

	if ttl == -2 {
		return 0, contract.ErrCacheKeyNotFound
	}

	if ttl == -1 {
		return 0, nil
	}

	return ttl, nil
}

// Ping verifies that the connection is still alive.
func (client *RedisClient) Ping(ctx context.Context) error {
	return (*redis.Client)(client).Ping(ctx).Err()
}
