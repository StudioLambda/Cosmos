package redis

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/studiolambda/cosmos/contract"
)

type Options redis.Options

type Client redis.Client

func New(options *Options) *Client {
	return (*Client)(redis.NewClient((*redis.Options)(options)))
}

// Get retrieves a value by key or returns contract.ErrNotFound if missing.
func (c *Client) Get(ctx context.Context, key string) (any, error) {
	val, err := (*redis.Client)(c).Get(ctx, key).Result()

	if errors.Is(err, redis.Nil) {
		return nil, fmt.Errorf("%w: %s", contract.ErrCacheKeyNotFound, key)
	}

	if err != nil {
		return nil, err
	}

	return val, nil
}

// Put sets a key with value and TTL.
func (c *Client) Put(ctx context.Context, key string, value any, ttl time.Duration) error {
	return (*redis.Client)(c).Set(ctx, key, value, ttl).Err()
}

// Delete removes a key.
func (c *Client) Delete(ctx context.Context, key string) error {
	return (*redis.Client)(c).Del(ctx, key).Err()
}

// Has checks if key exists.
func (c *Client) Has(ctx context.Context, key string) (bool, error) {
	n, err := (*redis.Client)(c).Exists(ctx, key).Result()

	if err != nil {
		return false, err
	}

	return n > 0, nil
}

// Pull retrieves and deletes a key atomically.
// NOTE: Redis lacks a native atomic get+del, so a transaction is used.
func (c *Client) Pull(ctx context.Context, key string) (any, error) {
	var val string

	txf := func(tx *redis.Tx) error {
		v, err := tx.Get(ctx, key).Result()

		if errors.Is(err, redis.Nil) {
			return fmt.Errorf("%w: %s", contract.ErrCacheKeyNotFound, key)
		}

		if err != nil {
			return err
		}

		val = v

		_, err = tx.Del(ctx, key).Result()

		return err
	}

	if err := (*redis.Client)(c).Watch(ctx, txf, key); err != nil {
		return nil, err
	}

	return val, nil
}

// Forever stores a value indefinitely.
func (c *Client) Forever(ctx context.Context, key string, value any) error {
	return (*redis.Client)(c).Set(ctx, key, value, 0).Err()
}

// Increment increases a key's integer value by 'by'.
func (c *Client) Increment(ctx context.Context, key string, by int64) (int64, error) {
	return (*redis.Client)(c).IncrBy(ctx, key, by).Result()
}

// Decrement decreases a key's integer value by 'by'.
func (c *Client) Decrement(ctx context.Context, key string, by int64) (int64, error) {
	return (*redis.Client)(c).DecrBy(ctx, key, by).Result()
}

// Remember gets or computes and caches a value with TTL.
func (c *Client) Remember(ctx context.Context, key string, ttl time.Duration, compute func() (any, error)) (any, error) {
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
func (c *Client) RememberForever(ctx context.Context, key string, compute func() (any, error)) (any, error) {
	return c.Remember(ctx, key, 0, compute)
}
