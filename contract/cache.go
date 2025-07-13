package contract

import (
	"context"
	"errors"
	"time"
)

// ErrCacheKeyNotFound should be returned when a key does not exist in the cache.
// Additional context should also be supplied, such as the cache key.
var ErrCacheKeyNotFound = errors.New("cache key not found")

// ErrCacheUnsupportedOperation should be returned when a method (e.g., Forever or atomic ops) is not supported by the backend.
// Additional context should also be supplied, such as the operation name.
var ErrCacheUnsupportedOperation = errors.New("cache unsupported operation")

// Cache defines a generic cache contract inspired by Laravel's Cache Repository.
// It supports basic CRUD, atomic operations, and lazy-loading via Remember patterns.
type Cache interface {
	// Get retrieves the value for the given key.
	// Returns nil and an error if the key is missing or retrieval fails.
	Get(ctx context.Context, key string) (any, error)

	// Put stores a value for the given key with a TTL.
	// Overwrites any existing value.
	Put(ctx context.Context, key string, value any, ttl time.Duration) error

	// Delete removes the cached value for the given key.
	// Does nothing if the key does not exist.
	Delete(ctx context.Context, key string) error

	// Has returns true if the key exists in the cache and is not expired.
	Has(ctx context.Context, key string) (bool, error)

	// Pull retrieves and removes the value for the given key.
	// Returns nil and an error if the key is missing.
	Pull(ctx context.Context, key string) (any, error)

	// Forever stores a value permanently (no TTL).
	// Only supported if the backend allows non-expiring keys.
	Forever(ctx context.Context, key string, value any) error

	// Increment atomically increases the integer value stored at key by the given
	// amount. Returns the new value or an error if the operation fails.
	Increment(ctx context.Context, key string, by int64) (int64, error)

	// Decrement atomically decreases the integer value stored at key by
	// the given amount. Returns the new value or an error if the operation fails.
	Decrement(ctx context.Context, key string, by int64) (int64, error)

	// Remember attempts to get the value for the given key.
	// If the key is missing, it calls the compute function, stores the result with the given TTL, and returns it.
	Remember(ctx context.Context, key string, ttl time.Duration, compute func() (any, error)) (any, error)

	// RememberForever is like Remember but stores the computed result without TTL (permanently).
	RememberForever(ctx context.Context, key string, compute func() (any, error)) (any, error)
}
