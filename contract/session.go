package contract

import (
	"context"
	"time"
)

// Session represents a user session with data storage and lifecycle management capabilities.
// It provides methods to store, retrieve, and manage session data, as well as control
// session expiration and regeneration for security purposes.
type Session interface {
	// SessionID returns the current session identifier. This may differ from the original
	// session ID if the session has been regenerated.
	SessionID() string

	// OriginalSessionID returns the session ID that was originally assigned to this session.
	// This remains constant even if the session is regenerated.
	OriginalSessionID() string

	// Get retrieves a value from the session by key. It returns the value and a boolean
	// indicating whether the key exists in the session. The value can be of any type.
	Get(key string) (any, bool)

	// Put stores a value in the session associated with the given key. If the key already
	// exists, its value is overwritten.
	Put(key string, value any)

	// Delete removes the value associated with the given key from the session.
	// If the key does not exist, this operation is a no-op.
	Delete(key string)

	// Extend updates the session's expiration time to the specified time. This is useful
	// for extending a session's lifetime during active use.
	Extend(expiresAt time.Time)

	// Regenerate creates a new session ID and associates it with this session.
	// This is commonly used after authentication to prevent session
	// fixation attacks. It returns an error if the regeneration process fails.
	Regenerate() error

	// Clear removes all data from the session while maintaining the session itself.
	Clear()

	// ExpiresAt returns the time at which the session will expire.
	ExpiresAt() time.Time

	// HasExpired returns true if the current time is past the session's expiration time.
	HasExpired() bool

	// ExpiresSoon returns true if the session will expire within the specified duration
	// from the current time. This is useful for triggering session renewal prompts.
	ExpiresSoon(delta time.Duration) bool

	// HasChanged returns true if the session data has been modified since it was loaded.
	// This is useful for determining whether the session needs to be persisted.
	HasChanged() bool

	// HasRegenerated returns true if the session has been regenerated (i.e., the session ID
	// has changed). This is useful for sending updated session identifiers to the client.
	HasRegenerated() bool

	// MarkAsUnchanged sets the session as if nothing has changed, therefore avoiding saving
	// the session when the request finishes.
	MarkAsUnchanged()
}

// SessionDriver defines the interface for persisting and retrieving session data.
// Implementations of SessionDriver are responsible for managing session storage,
// such as storing sessions in a database, cache, or file system.
type SessionDriver interface {
	// Get retrieves a session from persistent storage by its ID. It returns the session
	// and an error if the session cannot be found or if retrieval fails. If a session
	// with the given ID does not exist, an error should be returned.
	Get(ctx context.Context, id string) (Session, error)

	// Save persists a session to storage with the specified time-to-live (TTL).
	// The TTL parameter indicates how long the session should be retained in storage
	// before it can be automatically removed. It returns an error if the save operation fails.
	Save(ctx context.Context, session Session, ttl time.Duration) error

	// Delete removes a session from persistent storage by its ID. It returns an error
	// if the delete operation fails.
	Delete(ctx context.Context, id string) error
}
