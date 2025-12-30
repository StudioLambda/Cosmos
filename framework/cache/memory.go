package cache

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/studiolambda/cosmos/contract"
)

type Memory struct {
	mux   sync.Mutex
	store *cache.Cache
}

func NewMemory(expiration time.Duration, cleanup time.Duration) *Memory {
	return &Memory{
		store: cache.New(expiration, cleanup),
		mux:   sync.Mutex{},
	}
}

// Get retrieves the value for the given key.
func (m *Memory) Get(_ context.Context, key string) (any, error) {
	val, found := m.store.Get(key)

	if !found {
		return nil, fmt.Errorf("%w: %s", contract.ErrCacheKeyNotFound, key)
	}

	return val, nil
}

// Put stores a value for the given key with a TTL.
func (m *Memory) Put(_ context.Context, key string, value any, ttl time.Duration) error {
	m.store.Set(key, value, ttl)

	return nil
}

// Delete removes the cached value for the given key.
func (m *Memory) Delete(_ context.Context, key string) error {
	m.store.Delete(key)

	return nil
}

// Has returns true if the key exists in the cache and is not expired.
func (m *Memory) Has(_ context.Context, key string) (bool, error) {
	_, found := m.store.Get(key)

	return found, nil
}

// Pull retrieves and removes the value for the given key.
func (m *Memory) Pull(ctx context.Context, key string) (any, error) {
	m.mux.Lock()
	defer m.mux.Unlock()

	val, err := m.Get(ctx, key)

	if err != nil {
		return nil, err
	}

	if err := m.Delete(ctx, key); err != nil {
		return nil, err
	}

	return val, nil
}

// Forever stores a value permanently (no TTL).
func (m *Memory) Forever(_ context.Context, key string, value any) error {
	m.store.Set(key, value, cache.NoExpiration)

	return nil
}

// Increment atomically increases the integer value stored at key by the given amount.
func (m *Memory) Increment(ctx context.Context, key string, by int64) (int64, error) {
	m.mux.Lock()
	defer m.mux.Unlock()

	if found, _ := m.Has(ctx, key); !found {
		return 0, fmt.Errorf("%w: %s", contract.ErrCacheKeyNotFound, key)
	}

	next, err := m.store.IncrementInt64(key, by)

	return next, err
}

// Decrement atomically decreases the integer value stored at key by the given amount.
func (m *Memory) Decrement(ctx context.Context, key string, by int64) (int64, error) {
	m.mux.Lock()
	defer m.mux.Unlock()

	if found, _ := m.Has(ctx, key); !found {
		return 0, fmt.Errorf("%w: %s", contract.ErrCacheKeyNotFound, key)
	}

	next, err := m.store.DecrementInt64(key, by)

	return next, err
}

// Remember tries to get the value or compute and store it if not found.
func (m *Memory) Remember(ctx context.Context, key string, ttl time.Duration, compute func() (any, error)) (any, error) {
	val, err := m.Get(ctx, key)

	if err == nil {
		return val, nil
	}

	val, err = compute()

	if err != nil {
		return nil, err
	}

	_ = m.Put(ctx, key, val, ttl)

	return val, nil
}

// RememberForever tries to get the value or compute and store it forever if not found.
func (m *Memory) RememberForever(ctx context.Context, key string, compute func() (any, error)) (any, error) {
	val, err := m.Get(ctx, key)

	if err == nil {
		return val, nil
	}

	val, err = compute()

	if err != nil {
		return nil, err
	}

	_ = m.Forever(ctx, key, val)

	return val, nil
}
