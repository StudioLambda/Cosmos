package contract

import (
	"context"
	"encoding/json/v2"
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

	// Ping verifies that the connection is still alive.
	Ping(ctx context.Context) error
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
// as Pull, Forever, Remember, and atomic counters.
type Cache struct {
	driver CacheDriver
}

// NewCache creates a new [Cache] that delegates storage to the given driver.
//
// Example:
//
//	ctx := context.Background()
//	cache := contract.NewCache(driver)
//	if err := cache.Put(ctx, "users:1", map[string]any{"id": 1, "name": "Alice"}, time.Minute); err != nil {
//		return err
//	}
func NewCache(driver CacheDriver) *Cache {
	return &Cache{driver: driver}
}

// Driver returns the underlying [CacheDriver].
//
// Example:
//
//	cache := contract.NewCache(driver)
//	rawDriver := cache.Driver()
//	_ = rawDriver
func (cache *Cache) Driver() CacheDriver {
	return cache.driver
}

// Get retrieves the value for the given key and JSON-decodes it into
// the provided destination. Returns [ErrCacheKeyNotFound] when the
// key is missing.
//
// Example:
//
//	ctx := context.Background()
//	cache := contract.NewCache(driver)
//	user, err := cache.Get[User](ctx, "users:1")
//	if err != nil {
//		return err
//	}
//	_ = user
func (cache *Cache) Get[T any](ctx context.Context, key string) (res T, err error) {
	raw, err := cache.driver.Get(ctx, key)

	if err != nil {
		return res, err
	}

	if err := json.Unmarshal(raw, &res); err != nil {
		return res, err
	}

	return res, nil
}

// Put JSON-encodes the value and stores it with the given TTL.
//
// Example:
//
//	ctx := context.Background()
//	cache := contract.NewCache(driver)
//	if err := cache.Put(ctx, "users:1", User{ID: 1}, 5*time.Minute); err != nil {
//		return err
//	}
func (cache *Cache) Put[T any](ctx context.Context, key string, value T, ttl time.Duration) error {
	raw, err := json.Marshal(value)

	if err != nil {
		return err
	}

	return cache.driver.Put(ctx, key, raw, ttl)
}

// Delete removes the cached value for the given key.
//
// Example:
//
//	ctx := context.Background()
//	cache := contract.NewCache(driver)
//	if err := cache.Delete(ctx, "users:1"); err != nil {
//		return err
//	}
func (cache *Cache) Delete(ctx context.Context, key string) error {
	return cache.driver.Delete(ctx, key)
}

// Has returns true if the key exists in the cache and is not expired.
//
// Example:
//
//	ctx := context.Background()
//	cache := contract.NewCache(driver)
//	exists, err := cache.Has(ctx, "users:1")
//	if err != nil {
//		return err
//	}
//	if exists {
//		// use cached value
//	}
func (cache *Cache) Has(ctx context.Context, key string) (bool, error) {
	return cache.driver.Has(ctx, key)
}

// Pull retrieves and removes the value for the given key, decoding
// it into dest. Returns [ErrCacheKeyNotFound] when the key is missing.
//
// Example:
//
//	ctx := context.Background()
//	cache := contract.NewCache(driver)
//	token, err := cache.Pull[string](ctx, "password-reset:abc")
//	if err != nil {
//		return err
//	}
//	_ = token
func (cache *Cache) Pull[T any](ctx context.Context, key string) (res T, err error) {
	raw, err := cache.driver.Get(ctx, key)

	if err != nil {
		return res, err
	}

	if err := cache.driver.Delete(ctx, key); err != nil {
		return res, err
	}

	if err := json.Unmarshal(raw, &res); err != nil {
		return res, err
	}

	return res, nil
}

// Forever stores a value permanently (zero TTL).
//
// Example:
//
//	ctx := context.Background()
//	cache := contract.NewCache(driver)
//	if err := cache.Forever(ctx, "feature-flags", map[string]bool{"beta": true}); err != nil {
//		return err
//	}
func (cache *Cache) Forever[T any](ctx context.Context, key string, value T) error {
	return cache.Put(ctx, key, value, 0)
}

// Increment atomically increases the integer value at key by the
// given delta. Returns [ErrCacheUnsupportedOperation] if the driver
// does not implement [CacheCounter].
//
// Example:
//
//	ctx := context.Background()
//	cache := contract.NewCache(driver)
//	count, err := cache.Increment(ctx, "api:hits", 1)
//	if err != nil {
//		return err
//	}
//	_ = count
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
//
// Example:
//
//	ctx := context.Background()
//	cache := contract.NewCache(driver)
//	count, err := cache.Decrement(ctx, "jobs:pending", 1)
//	if err != nil {
//		return err
//	}
//	_ = count
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
//
// Example:
//
//	ctx := context.Background()
//	cache := contract.NewCache(driver)
//	profile, err := cache.Remember(ctx, "profiles:1", time.Minute, func() (Profile, error) {
//		return fetchProfile(ctx, 1)
//	})
//	if err != nil {
//		return err
//	}
//	_ = profile
func (cache *Cache) Remember[T any](ctx context.Context, key string, ttl time.Duration, compute func() (T, error)) (res T, err error) {
	val, err := cache.Get[T](ctx, key)

	if err == nil {
		return val, nil
	}

	if !errors.Is(err, ErrCacheKeyNotFound) {
		return res, err
	}

	value, err := compute()

	if err != nil {
		return res, err
	}

	if err := cache.Put(ctx, key, value, ttl); err != nil {
		return res, err
	}

	return value, nil
}

// RememberForever is like [Cache.Remember] but stores the computed
// result without expiration.
//
// Example:
//
//	ctx := context.Background()
//	cache := contract.NewCache(driver)
//	settings, err := cache.RememberForever(ctx, "settings:public", func() (Settings, error) {
//		return loadSettings(ctx)
//	})
//	if err != nil {
//		return err
//	}
//	_ = settings
func (cache *Cache) RememberForever[T any](ctx context.Context, key string, compute func() (T, error)) (res T, err error) {
	return cache.Remember(ctx, key, 0, compute)
}
