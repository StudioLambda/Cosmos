package session

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/studiolambda/cosmos/contract"
)

// CacheDriver implements contract.SessionDriver by storing sessions
// in any contract.Cache backend. Sessions are keyed with a configurable
// prefix to avoid collisions with other cached data.
type CacheDriver struct {
	cache   contract.Cache
	options CacheDriverOptions
}

// CacheDriverOptions holds configuration for the CacheDriver.
// The prefix is prepended to session IDs when forming cache keys.
type CacheDriverOptions struct {
	prefix string
}

// ErrCacheDriverInvalidType is returned when a value retrieved from the
// cache cannot be type-asserted to contract.Session, indicating a
// serialization mismatch or cache key collision.
var ErrCacheDriverInvalidType = errors.New("invalid cache type")

// NewCacheDriver creates a CacheDriver with the default key prefix
// "cosmos.sessions". Use NewCacheDriverWith for custom options.
func NewCacheDriver(cache contract.Cache) *CacheDriver {
	return NewCacheDriverWith(cache, CacheDriverOptions{
		prefix: "cosmos.sessions",
	})
}

// NewCacheDriverWith creates a CacheDriver with the given cache
// backend and options, allowing a custom key prefix.
func NewCacheDriverWith(cache contract.Cache, options CacheDriverOptions) *CacheDriver {
	return &CacheDriver{
		cache:   cache,
		options: options,
	}
}

// key builds the full cache key by joining the configured prefix
// with the session ID.
func (driver *CacheDriver) key(id string) string {
	return fmt.Sprintf("%s.%s", driver.options.prefix, id)
}

// Get retrieves a session from the cache by its ID. It returns
// ErrCacheDriverInvalidType if the cached value is not a valid
// contract.Session.
func (driver *CacheDriver) Get(ctx context.Context, id string) (contract.Session, error) {
	cacheKey := driver.key(id)
	value, err := driver.cache.Get(ctx, cacheKey)

	if err != nil {
		return nil, err
	}

	if session, ok := value.(contract.Session); ok {
		return session, nil
	}

	return nil, ErrCacheDriverInvalidType
}

// Save persists a session in the cache with the given TTL.
func (driver *CacheDriver) Save(ctx context.Context, session contract.Session, ttl time.Duration) error {
	return driver.cache.Put(ctx, driver.key(session.SessionID()), session, ttl)
}

// Delete removes a session from the cache by its ID.
func (driver *CacheDriver) Delete(ctx context.Context, id string) error {
	return driver.cache.Delete(ctx, driver.key(id))
}
