package memory

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/studiolambda/cosmos/contract"

	"github.com/patrickmn/go-cache"
)

// Memory implements [contract.CacheDriver]
// using an in-memory store backed by patrickmn/go-cache.
//
// Memory is suitable for single-process applications and testing scenarios
// where persistence across restarts is not required.
type Memory struct {
	store *cache.Cache
	mu    sync.Mutex
}

// MemoryConfig configures the in-memory cache.
type MemoryConfig struct {
	// Expiration is retained for compatibility with the underlying cache.
	// Cosmos operations always use their explicit TTL; zero means no expiry.
	Expiration time.Duration

	// Cleanup is the interval used to purge expired items.
	Cleanup time.Duration
}

// DefaultMemoryConfig returns the default in-memory cache configuration.
func DefaultMemoryConfig() MemoryConfig {
	return MemoryConfig{}
}

// FromConfiguration populates the memory-cache configuration from configuration.
func (config *MemoryConfig) FromConfiguration(configuration *contract.Configuration) {
	*config = DefaultMemoryConfig()
	config.Expiration = configuration.GetOr("expiration", config.Expiration)
	config.Cleanup = configuration.GetOr("cleanup", config.Cleanup)
}

// NewMemory creates a Memory cache with the given configuration.
// Zero TTL entries never expire, consistent with [contract.CacheDriver].
// Expired entries are purged at the cleanup interval.
func NewMemory(config MemoryConfig) *Memory {
	return &Memory{
		store: cache.New(config.Expiration, config.Cleanup),
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
// A zero TTL stores the entry without expiration.
func (memory *Memory) Put(_ context.Context, key string, value []byte, ttl time.Duration) error {
	if ttl == 0 {
		ttl = cache.NoExpiration
	}

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

// Add stores raw bytes only if the key does not already exist.
func (memory *Memory) Add(_ context.Context, key string, value []byte, ttl time.Duration) (bool, error) {
	memory.mu.Lock()
	defer memory.mu.Unlock()

	if _, found := memory.store.Get(key); found {
		return false, nil
	}

	if ttl == 0 {
		ttl = cache.NoExpiration
	}

	memory.store.Set(key, value, ttl)

	return true, nil
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
	return memory.adjust(key, delta)
}

// Decrement atomically decreases the integer value stored at key by
// the given amount. Returns [contract.ErrCacheKeyNotFound] if the key
// does not exist.
func (memory *Memory) Decrement(_ context.Context, key string, delta int64) (int64, error) {
	return memory.adjust(key, -delta)
}

// TTL returns the remaining lifetime for key.
func (memory *Memory) TTL(_ context.Context, key string) (time.Duration, error) {
	memory.mu.Lock()
	defer memory.mu.Unlock()

	item, found := memory.store.Items()[key]
	if !found {
		return 0, contract.ErrCacheKeyNotFound
	}

	if item.Expiration <= 0 {
		return 0, nil
	}

	remaining := time.Until(time.Unix(0, item.Expiration))
	if remaining <= 0 {
		return 0, contract.ErrCacheKeyNotFound
	}

	return remaining, nil
}

func (memory *Memory) adjust(key string, delta int64) (int64, error) {
	memory.mu.Lock()
	defer memory.mu.Unlock()

	value, found := memory.store.Get(key)
	if !found {
		return 0, contract.ErrCacheKeyNotFound
	}

	raw, ok := value.([]byte)
	if !ok {
		return 0, contract.ErrCacheUnsupportedOperation
	}

	current, err := strconv.ParseInt(string(raw), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("memory: parse counter value: %w", err)
	}

	next := current + delta

	item, found := memory.store.Items()[key]
	if !found {
		return 0, contract.ErrCacheKeyNotFound
	}

	ttl := time.Duration(0)
	if item.Expiration > 0 {
		ttl = time.Until(time.Unix(0, item.Expiration))
		if ttl <= 0 {
			return 0, contract.ErrCacheKeyNotFound
		}
	}

	memory.store.Set(key, []byte(strconv.FormatInt(next, 10)), ttl)

	return next, nil
}
