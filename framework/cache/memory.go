package cache

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/studiolambda/cosmos/contract"
)

// Memory implements contract.Cache using an in-memory store backed
// by patrickmn/go-cache. It is suitable for single-process
// applications and testing scenarios where persistence across
// restarts is not required.
type Memory struct {
	mux   sync.Mutex
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

// Get retrieves the value for the given key from the in-memory store.
// Returns contract.ErrCacheKeyNotFound wrapped with the key name
// when the key does not exist or has expired.
func (memory *Memory) Get(_ context.Context, key string) (any, error) {
	val, found := memory.store.Get(key)

	if !found {
		return nil, fmt.Errorf("%w: %s", contract.ErrCacheKeyNotFound, key)
	}

	return val, nil
}

// Put stores a value in the in-memory cache with the given TTL.
// A zero TTL uses the default expiration configured at creation.
func (memory *Memory) Put(_ context.Context, key string, value any, ttl time.Duration) error {
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

// Pull atomically retrieves and removes the value for the given key.
// It holds a mutex to prevent races between the get and delete steps.
func (memory *Memory) Pull(ctx context.Context, key string) (any, error) {
	memory.mux.Lock()
	defer memory.mux.Unlock()

	val, err := memory.Get(ctx, key)

	if err != nil {
		return nil, err
	}

	if err := memory.Delete(ctx, key); err != nil {
		return nil, err
	}

	return val, nil
}

// Forever stores a value permanently with no expiration.
func (memory *Memory) Forever(_ context.Context, key string, value any) error {
	memory.store.Set(key, value, cache.NoExpiration)

	return nil
}

// Increment atomically increases the integer value stored at key by
// the given amount. Returns contract.ErrCacheKeyNotFound if the key
// does not exist.
func (memory *Memory) Increment(ctx context.Context, key string, by int64) (int64, error) {
	memory.mux.Lock()
	defer memory.mux.Unlock()

	if found, _ := memory.Has(ctx, key); !found {
		return 0, fmt.Errorf("%w: %s", contract.ErrCacheKeyNotFound, key)
	}

	return memory.store.IncrementInt64(key, by)
}

// Decrement atomically decreases the integer value stored at key by
// the given amount. Returns contract.ErrCacheKeyNotFound if the key
// does not exist.
func (memory *Memory) Decrement(ctx context.Context, key string, by int64) (int64, error) {
	memory.mux.Lock()
	defer memory.mux.Unlock()

	if found, _ := memory.Has(ctx, key); !found {
		return 0, fmt.Errorf("%w: %s", contract.ErrCacheKeyNotFound, key)
	}

	return memory.store.DecrementInt64(key, by)
}

// Remember retrieves the cached value for the given key, or computes
// and stores it if the key is not found. The compute function is only
// called on a cache miss.
func (memory *Memory) Remember(ctx context.Context, key string, ttl time.Duration, compute func() (any, error)) (any, error) {
	val, err := memory.Get(ctx, key)

	if err == nil {
		return val, nil
	}

	val, err = compute()

	if err != nil {
		return nil, err
	}

	_ = memory.Put(ctx, key, val, ttl)

	return val, nil
}

// RememberForever retrieves the cached value for the given key, or
// computes and stores it permanently if the key is not found.
func (memory *Memory) RememberForever(ctx context.Context, key string, compute func() (any, error)) (any, error) {
	val, err := memory.Get(ctx, key)

	if err == nil {
		return val, nil
	}

	val, err = compute()

	if err != nil {
		return nil, err
	}

	_ = memory.Forever(ctx, key, val)

	return val, nil
}
