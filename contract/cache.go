package contract

import (
	"context"
	"encoding/json"
	"errors"
	"time"
)

var (
	// ErrCacheKeyNotFound is returned when a key does not exist in the cache.
	ErrCacheKeyNotFound = errors.New("cache key not found")

	// ErrCacheUnsupportedOperation is returned when a method such as
	// atomic increment/decrement is not supported by the cache driver.
	ErrCacheUnsupportedOperation = errors.New("cache unsupported operation")
)

// CacheDriver defines the minimal contract that cache backends must
// implement. Drivers operate on raw bytes, leaving serialization to
// the [Cache] wrapper. A zero TTL means the entry should never expire.
type CacheDriver interface {
	// Get retrieves the raw bytes for the given key.
	// Returns [ErrCacheKeyNotFound] when the key is missing or expired.
	Get(ctx context.Context, key string) ([]byte, error)

	// Put stores raw bytes for the given key with a TTL.
	// A zero TTL stores the entry without expiration.
	Put(ctx context.Context, key string, value []byte, ttl time.Duration) error

	// Delete removes the entry for the given key.
	// Does nothing if the key does not exist.
	Delete(ctx context.Context, key string) error

	// Has returns true if the key exists and is not expired.
	Has(ctx context.Context, key string) (bool, error)
}

// CacheCounter is an optional interface that cache drivers may
// implement to support atomic increment and decrement operations.
// When the driver does not implement this interface, the [Cache]
// wrapper returns [ErrCacheUnsupportedOperation].
type CacheCounter interface {
	// Increment atomically increases the integer value at key by
	// the given delta. Returns the new value.
	Increment(ctx context.Context, key string, delta int64) (int64, error)

	// Decrement atomically decreases the integer value at key by
	// the given delta. Returns the new value.
	Decrement(ctx context.Context, key string, delta int64) (int64, error)
}

// Cache provides a type-safe caching layer over a [CacheDriver].
// It handles JSON serialization and offers convenience methods such
// as Pull, Forever, Remember, and atomic counters. When generic
// methods become available in Go, the Get and Remember methods will
// be updated to return typed values directly.
type Cache struct {
	driver CacheDriver
}

// NewCache creates a new [Cache] that delegates storage to the given driver.
func NewCache(driver CacheDriver) *Cache {
	return &Cache{driver: driver}
}

// Driver returns the underlying [CacheDriver].
func (cache *Cache) Driver() CacheDriver {
	return cache.driver
}

// Get retrieves the value for the given key and JSON-decodes it into
// the provided destination. Returns [ErrCacheKeyNotFound] when the
// key is missing.
func (cache *Cache) Get(ctx context.Context, key string, dest any) error {
	raw, err := cache.driver.Get(ctx, key)

	if err != nil {
		return err
	}

	return json.Unmarshal(raw, dest)
}

// Put JSON-encodes the value and stores it with the given TTL.
func (cache *Cache) Put(ctx context.Context, key string, value any, ttl time.Duration) error {
	raw, err := json.Marshal(value)

	if err != nil {
		return err
	}

	return cache.driver.Put(ctx, key, raw, ttl)
}

// Delete removes the cached value for the given key.
func (cache *Cache) Delete(ctx context.Context, key string) error {
	return cache.driver.Delete(ctx, key)
}

// Has returns true if the key exists in the cache and is not expired.
func (cache *Cache) Has(ctx context.Context, key string) (bool, error) {
	return cache.driver.Has(ctx, key)
}

// Pull retrieves and removes the value for the given key, decoding
// it into dest. Returns [ErrCacheKeyNotFound] when the key is missing.
func (cache *Cache) Pull(ctx context.Context, key string, dest any) error {
	raw, err := cache.driver.Get(ctx, key)

	if err != nil {
		return err
	}

	if err := cache.driver.Delete(ctx, key); err != nil {
		return err
	}

	return json.Unmarshal(raw, dest)
}

// Forever stores a value permanently (zero TTL).
func (cache *Cache) Forever(ctx context.Context, key string, value any) error {
	return cache.Put(ctx, key, value, 0)
}

// Increment atomically increases the integer value at key by the
// given delta. Returns [ErrCacheUnsupportedOperation] if the driver
// does not implement [CacheCounter].
func (cache *Cache) Increment(ctx context.Context, key string, delta int64) (int64, error) {
	counter, ok := cache.driver.(CacheCounter)

	if !ok {
		return 0, ErrCacheUnsupportedOperation
	}

	return counter.Increment(ctx, key, delta)
}

// Decrement atomically decreases the integer value at key by the
// given delta. Returns [ErrCacheUnsupportedOperation] if the driver
// does not implement [CacheCounter].
func (cache *Cache) Decrement(ctx context.Context, key string, delta int64) (int64, error) {
	counter, ok := cache.driver.(CacheCounter)

	if !ok {
		return 0, ErrCacheUnsupportedOperation
	}

	return counter.Decrement(ctx, key, delta)
}

// Remember retrieves the cached value for the given key, decoding it
// into dest. If the key is not found, it calls compute, stores the
// result with the given TTL, and decodes it into dest.
func (cache *Cache) Remember(ctx context.Context, key string, ttl time.Duration, dest any, compute func() (any, error)) error {
	raw, err := cache.driver.Get(ctx, key)

	if err == nil {
		return json.Unmarshal(raw, dest)
	}

	if !errors.Is(err, ErrCacheKeyNotFound) {
		return err
	}

	value, err := compute()

	if err != nil {
		return err
	}

	encoded, err := json.Marshal(value)

	if err != nil {
		return err
	}

	if err := cache.driver.Put(ctx, key, encoded, ttl); err != nil {
		return err
	}

	return json.Unmarshal(encoded, dest)
}

// RememberForever is like [Cache.Remember] but stores the computed
// result without expiration.
func (cache *Cache) RememberForever(ctx context.Context, key string, dest any, compute func() (any, error)) error {
	return cache.Remember(ctx, key, 0, dest, compute)
}
