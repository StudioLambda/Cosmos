package cache

import (
	"context"
	"errors"
	"time"

	"github.com/studiolambda/cosmos/contract"

	"github.com/redis/go-redis/v9"
)

// RedisOptions is an alias for redis.Options, exposing the full
// set of connection parameters without requiring a direct import
// of the go-redis package.
type RedisOptions = redis.Options

// RedisClient implements [contract.CacheDriver] and [contract.CacheCounter]
// using Redis as the backing store.
type RedisClient redis.Client

// NewRedis creates a RedisClient from the given connection options.
func NewRedis(options *RedisOptions) *RedisClient {
	return NewRedisFrom(redis.NewClient((*redis.Options)(options)))
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

// Increment atomically increases the integer value stored at key by
// the given amount. Redis auto-creates the key with value 0 if it
// does not exist before incrementing.
func (client *RedisClient) Increment(ctx context.Context, key string, delta int64) (int64, error) {
	return (*redis.Client)(client).IncrBy(ctx, key, delta).Result()
}

// Decrement atomically decreases the integer value stored at key by
// the given amount. Redis auto-creates the key with value 0 if it
// does not exist before decrementing.
func (client *RedisClient) Decrement(ctx context.Context, key string, delta int64) (int64, error) {
	return (*redis.Client)(client).DecrBy(ctx, key, delta).Result()
}

// Ping verifies that the connection is still alive.
func (client *RedisClient) Ping(ctx context.Context) error {
	return (*redis.Client)(client).Ping(ctx).Err()
}
