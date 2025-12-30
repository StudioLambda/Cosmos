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

type RedisOptions = redis.Options

type RedisClient redis.Client

func NewRedis(options *RedisOptions) *RedisClient {
	return (*RedisClient)(redis.NewClient((*redis.Options)(options)))
}

// Get retrieves a value by key or returns contract.ErrNotFound if missing.
func (c *RedisClient) Get(ctx context.Context, key string) (any, error) {
	v, err := (*redis.Client)(c).Get(ctx, key).Result()

	if errors.Is(err, redis.Nil) {
		return nil, fmt.Errorf("%w: %s", contract.ErrCacheKeyNotFound, key)
	}

	if err != nil {
		return nil, err
	}

	return v, nil
}

// Put sets a key with value and TTL.
func (c *RedisClient) Put(ctx context.Context, key string, value any, ttl time.Duration) error {
	return (*redis.Client)(c).Set(ctx, key, value, ttl).Err()
}

// Delete removes a key.
func (c *RedisClient) Delete(ctx context.Context, key string) error {
	return (*redis.Client)(c).Del(ctx, key).Err()
}

// Has checks if key exists.
func (c *RedisClient) Has(ctx context.Context, key string) (bool, error) {
	n, err := (*redis.Client)(c).Exists(ctx, key).Result()

	if err != nil {
		return false, err
	}

	return n > 0, nil
}

// Pull retrieves and deletes a key atomically.
func (c *RedisClient) Pull(ctx context.Context, key string) (v any, e error) {
	encoded, err := (*redis.Client)(c).GetDel(ctx, key).Result()

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

// Forever stores a value indefinitely.
func (c *RedisClient) Forever(ctx context.Context, key string, value any) error {
	return c.Put(ctx, key, value, 0)
}

// Increment increases a key's integer value by 'by'.
func (c *RedisClient) Increment(ctx context.Context, key string, by int64) (int64, error) {
	return (*redis.Client)(c).IncrBy(ctx, key, by).Result()
}

// Decrement decreases a key's integer value by 'by'.
func (c *RedisClient) Decrement(ctx context.Context, key string, by int64) (int64, error) {
	return (*redis.Client)(c).DecrBy(ctx, key, by).Result()
}

// Remember gets or computes and caches a value with TTL.
func (c *RedisClient) Remember(ctx context.Context, key string, ttl time.Duration, compute func() (any, error)) (any, error) {
	val, err := c.Get(ctx, key)

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

	if err := c.Put(ctx, key, val, ttl); err != nil {
		return nil, err
	}

	return val, nil
}

// RememberForever caches a computed value indefinitely.
func (c *RedisClient) RememberForever(ctx context.Context, key string, compute func() (any, error)) (any, error) {
	return c.Remember(ctx, key, 0, compute)
}
