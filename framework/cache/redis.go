package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/studiolambda/cosmos/contract"
)

// RedisOptions is an alias for redis.Options, exposing the full
// set of connection parameters without requiring a direct import
// of the go-redis package.
type RedisOptions = redis.Options

// RedisClient implements contract.Cache using Redis as the backing
// store. It is defined as a type conversion of redis.Client so that
// cache methods can be attached without wrapping in a separate struct.
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

// Get retrieves a value by key. Returns contract.ErrCacheKeyNotFound
// wrapped with the key name when the key does not exist.
func (client *RedisClient) Get(ctx context.Context, key string) (any, error) {
	v, err := (*redis.Client)(client).Get(ctx, key).Result()

	if errors.Is(err, redis.Nil) {
		return nil, fmt.Errorf("%w: %s", contract.ErrCacheKeyNotFound, key)
	}

	if err != nil {
		return nil, err
	}

	return v, nil
}

// Put stores a value with the given TTL. A zero TTL means the key
// will not expire.
func (client *RedisClient) Put(ctx context.Context, key string, value any, ttl time.Duration) error {
	return (*redis.Client)(client).Set(ctx, key, value, ttl).Err()
}

// Delete removes a key from Redis.
func (client *RedisClient) Delete(ctx context.Context, key string) error {
	return (*redis.Client)(client).Del(ctx, key).Err()
}

// Has reports whether the key exists in Redis.
func (client *RedisClient) Has(ctx context.Context, key string) (bool, error) {
	n, err := (*redis.Client)(client).Exists(ctx, key).Result()

	if err != nil {
		return false, err
	}

	return n > 0, nil
}

// Pull atomically retrieves and deletes a key using Redis GETDEL.
// The stored value is JSON-decoded into the return value.
func (client *RedisClient) Pull(ctx context.Context, key string) (v any, err error) {
	encoded, err := (*redis.Client)(client).GetDel(ctx, key).Result()

	if errors.Is(err, redis.Nil) {
		return nil, fmt.Errorf("%w: %s", contract.ErrCacheKeyNotFound, key)
	}

	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal([]byte(encoded), &v); err != nil {
		return nil, err
	}

	return v, nil
}

// Forever stores a value with no expiration.
func (client *RedisClient) Forever(ctx context.Context, key string, value any) error {
	return client.Put(ctx, key, value, 0)
}

// Increment increases the integer value at key by the given amount.
func (client *RedisClient) Increment(ctx context.Context, key string, by int64) (int64, error) {
	return (*redis.Client)(client).IncrBy(ctx, key, by).Result()
}

// Decrement decreases the integer value at key by the given amount.
func (client *RedisClient) Decrement(ctx context.Context, key string, by int64) (int64, error) {
	return (*redis.Client)(client).DecrBy(ctx, key, by).Result()
}

// Remember retrieves the cached value for key, or computes and stores
// it with the given TTL on a cache miss. Non-"key not found" errors
// from Get are returned immediately without calling compute.
func (client *RedisClient) Remember(ctx context.Context, key string, ttl time.Duration, compute func() (any, error)) (any, error) {
	val, err := client.Get(ctx, key)

	if err == nil {
		return val, nil
	}

	if !errors.Is(err, contract.ErrCacheKeyNotFound) {
		return nil, err
	}

	val, err = compute()

	if err != nil {
		return nil, err
	}

	if err := client.Put(ctx, key, val, ttl); err != nil {
		return nil, err
	}

	return val, nil
}

// RememberForever retrieves the cached value for key, or computes
// and stores it permanently on a cache miss.
func (client *RedisClient) RememberForever(ctx context.Context, key string, compute func() (any, error)) (any, error) {
	return client.Remember(ctx, key, 0, compute)
}
