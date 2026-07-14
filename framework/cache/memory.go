package cache

import (
	"context"
	"time"

	"github.com/studiolambda/cosmos/contract"

	"github.com/patrickmn/go-cache"
)

// Memory implements [contract.CacheDriver] and [contract.CacheCounter]
// using an in-memory store backed by patrickmn/go-cache.
//
// Memory is suitable for single-process applications and testing scenarios
// where persistence across restarts is not required.
type Memory struct {
	store *cache.Cache
}

// NewMemory creates a Memory cache with the given default expiration
// and cleanup interval. Items without an explicit TTL use the default
// expiration, and expired items are purged at the cleanup interval.
func NewMemory(expiration time.Duration, cleanup time.Duration) *Memory {
	return &Memory{
		store: cache.New(expiration, cleanup),
	}
}

// Get retrieves the raw bytes for the given key from the in-memory
// store. Returns [contract.ErrCacheKeyNotFound] when the key does
// not exist or has expired.
func (memory *Memory) Get(_ context.Context, key string) ([]byte, error) {
	val, found := memory.store.Get(key)

	if !found {
		return nil, contract.ErrCacheKeyNotFound
	}

	raw, ok := val.([]byte)

	if !ok {
		return nil, contract.ErrCacheKeyNotFound
	}

	return raw, nil
}

// Put stores raw bytes in the in-memory cache with the given TTL.
// A zero TTL uses the default expiration configured at creation.
func (memory *Memory) Put(_ context.Context, key string, value []byte, ttl time.Duration) error {
	memory.store.Set(key, value, ttl)

	return nil
}

// Delete removes the cached value for the given key. Deleting a
// non-existent key is a no-op.
func (memory *Memory) Delete(_ context.Context, key string) error {
	memory.store.Delete(key)

	return nil
}

// Has reports whether the key exists in the cache and has not expired.
func (memory *Memory) Has(_ context.Context, key string) (bool, error) {
	_, found := memory.store.Get(key)

	return found, nil
}

// Store returns the underlying go-cache instance. This is primarily
// useful for testing scenarios where direct access to the store is
// needed, such as storing int64 values for increment/decrement tests.
func (memory *Memory) Store() *cache.Cache {
	return memory.store
}

// Increment atomically increases the integer value stored at key by
// the given amount. Returns [contract.ErrCacheKeyNotFound] if the key
// does not exist.
func (memory *Memory) Increment(_ context.Context, key string, delta int64) (int64, error) {
	result, err := memory.store.IncrementInt64(key, delta)

	if err != nil {
		return 0, contract.ErrCacheKeyNotFound
	}

	return result, nil
}

// Decrement atomically decreases the integer value stored at key by
// the given amount. Returns [contract.ErrCacheKeyNotFound] if the key
// does not exist.
func (memory *Memory) Decrement(_ context.Context, key string, delta int64) (int64, error) {
	result, err := memory.store.DecrementInt64(key, delta)

	if err != nil {
		return 0, contract.ErrCacheKeyNotFound
	}

	return result, nil
}

// Ping verifies that the connection is still alive.
func (memory *Memory) Ping(ctx context.Context) error {
	return nil // always connected, in memory
}
