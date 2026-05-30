package session

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/studiolambda/cosmos/contract"
)

// CacheDriver implements [contract.SessionDriver] by storing sessions
// in any [contract.CacheDriver] backend. Sessions are JSON-serialized
// and keyed with a configurable prefix to avoid collisions.
//
// WARNING: Session data is stored without encryption. When using a
// remote backend such as Redis, session contents are transmitted and
// persisted in plaintext. For sensitive data, callers should encrypt
// values before calling Put or use a backend with transport/at-rest
// encryption.
type CacheDriver struct {
	cache   contract.CacheDriver
	options CacheDriverOptions
}

// CacheDriverOptions holds configuration for the CacheDriver.
type CacheDriverOptions struct {
	Prefix string
}

// sessionData is the serializable representation of a session for
// storage in the cache backend.
type sessionData struct {
	ID        string         `json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	ExpiresAt time.Time      `json:"expires_at"`
	Storage   map[string]any `json:"storage"`
}

// NewCacheDriver creates a CacheDriver with the default key prefix
// "cosmos.sessions".
func NewCacheDriver(cache contract.CacheDriver) *CacheDriver {
	return NewCacheDriverWith(cache, CacheDriverOptions{
		Prefix: "cosmos.sessions",
	})
}

// NewCacheDriverWith creates a CacheDriver with the given cache
// backend and options.
func NewCacheDriverWith(cache contract.CacheDriver, options CacheDriverOptions) *CacheDriver {
	return &CacheDriver{
		cache:   cache,
		options: options,
	}
}

// key builds the full cache key by joining the configured prefix
// with the session ID.
func (driver *CacheDriver) key(id string) string {
	return fmt.Sprintf("%s.%s", driver.options.Prefix, id)
}

// Get retrieves a session from the cache by its ID.
func (driver *CacheDriver) Get(ctx context.Context, id string) (*contract.Session, error) {
	raw, err := driver.cache.Get(ctx, driver.key(id))

	if err != nil {
		return nil, err
	}

	var data sessionData

	if err := json.Unmarshal(raw, &data); err != nil {
		return nil, err
	}

	return contract.NewSessionFrom(data.ID, data.CreatedAt, data.ExpiresAt, data.Storage), nil
}

// Save persists a session in the cache with the given TTL.
func (driver *CacheDriver) Save(ctx context.Context, session *contract.Session, ttl time.Duration) error {
	data := sessionData{
		ID:        session.SessionID(),
		CreatedAt: session.CreatedAt(),
		ExpiresAt: session.ExpiresAt(),
		Storage:   session.All(),
	}

	raw, err := json.Marshal(data)

	if err != nil {
		return err
	}

	return driver.cache.Put(ctx, driver.key(session.SessionID()), raw, ttl)
}

// Delete removes a session from the cache by its ID.
func (driver *CacheDriver) Delete(ctx context.Context, id string) error {
	return driver.cache.Delete(ctx, driver.key(id))
}
