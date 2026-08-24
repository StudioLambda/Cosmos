package contract

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json/v2"
	"errors"
	"fmt"
	"maps"
	"sync"
	"time"
)

// sessionKey is a private type used as a context key to avoid collisions.
type sessionKey struct{}

// SessionDriver defines the interface for persisting and retrieving
// session data. Implementations manage session storage in backends
// such as databases, caches, or file systems.
type SessionDriver interface {
	// Get retrieves a session from persistent storage by its ID.
	Get(ctx context.Context, id string) (*Session, error)

	// Save persists a session to storage with the specified TTL.
	Save(ctx context.Context, session *Session, ttl time.Duration) error

	// Delete removes a session from persistent storage by its ID.
	Delete(ctx context.Context, id string) error
}

// Session represents a user session with data storage and lifecycle
// management capabilities. It provides methods to store, retrieve,
// and manage session data, as well as control session expiration
// and regeneration for security purposes. Access is protected by
// a mutex to ensure thread-safe operations.
type Session struct {
	// originalID is the session ID assigned when the session was
	// first created. This value remains constant even if the
	// session is regenerated.
	originalID string

	// id is the current session identifier. This may differ from
	// originalID if the session has been regenerated.
	id string

	// createdAt records the absolute time the session was first created.
	createdAt time.Time

	// expiresAt is the time at which the session will expire.
	expiresAt time.Time

	// storage holds the session data as key-value pairs.
	storage map[string]any

	// mutex protects concurrent access to the session fields.
	mutex sync.Mutex

	// changed tracks whether the session data has been modified.
	changed bool
}

// sessionIDLength is the number of random bytes used to generate
// a session ID. 32 bytes provides 256 bits of entropy.
const sessionIDLength = 32

// SessionKey is the context key used to store and retrieve the session from a context.Context.
var SessionKey = sessionKey{}

// ErrSessionKeyNotFound indicates that a session has no value for a key.
var ErrSessionKeyNotFound = errors.New("session key not found")

// ErrSessionInvalidValueType indicates that a session value cannot be decoded
// into its requested type.
var ErrSessionInvalidValueType = errors.New("session invalid value type")

// generateSessionID generates a cryptographically random session
// ID using crypto/rand and base64url encoding (43 characters).
func generateSessionID() string {
	value := make([]byte, sessionIDLength)

	// crypto/rand.Read always fills value or terminates the process when the
	// operating system cannot provide cryptographically secure randomness.
	_, _ = rand.Read(value)

	return base64.RawURLEncoding.EncodeToString(value)
}

// NewSession creates a new session with the specified expiration
// time and initial storage data. It generates a cryptographically
// random session ID. The session is marked as changed to ensure it
// is persisted on first save. Its error result is retained for compatibility
// and is always nil because crypto/rand.Read is infallible.
//
// Example:
//
//	session, err := contract.NewSession(time.Now().Add(24*time.Hour), map[string]any{"role": "user"})
//	if err != nil {
//		return err
//	}
//	_ = session
func NewSession(expiresAt time.Time, storage map[string]any) (*Session, error) {
	id := generateSessionID()

	return &Session{
		originalID: id,
		id:         id,
		createdAt:  time.Now(),
		expiresAt:  expiresAt,
		storage:    storage,
		changed:    true,
	}, nil
}

// NewSessionFrom reconstructs a session from persisted data. Unlike
// [NewSession], it does not generate a new ID or mark the session as
// changed. This is used by session drivers when loading from storage.
//
// Example:
//
//	session := contract.NewSessionFrom(id, createdAt, expiresAt, storedValues)
//	_ = session
func NewSessionFrom(id string, createdAt time.Time, expiresAt time.Time, storage map[string]any) *Session {
	return &Session{
		originalID: id,
		id:         id,
		createdAt:  createdAt,
		expiresAt:  expiresAt,
		storage:    storage,
		changed:    false,
	}
}

// All returns a copy of all session data. This is used by session
// drivers when serializing the session for storage.
func (session *Session) All() map[string]any {
	session.mutex.Lock()
	defer session.mutex.Unlock()

	return maps.Clone(session.storage)
}

// SessionID returns the current session identifier. This may differ
// from the original session ID if the session has been regenerated.
func (session *Session) SessionID() string {
	session.mutex.Lock()
	defer session.mutex.Unlock()

	return session.id
}

// OriginalSessionID returns the session identifier that was assigned
// when the session was initially created.
func (session *Session) OriginalSessionID() string {
	session.mutex.Lock()
	defer session.mutex.Unlock()

	return session.originalID
}

// Get retrieves a value from the session storage by key. It returns
// the value and a boolean indicating whether the key exists.
//
// Example:
//
//	userID, err := session.Get[int]("user_id")
//	if err != nil {
//		return err
//	}
//	_ = userID
func (session *Session) Get[T any](key string) (res T, err error) {
	session.mutex.Lock()
	defer session.mutex.Unlock()

	raw, ok := session.storage[key]

	if !ok {
		return res, fmt.Errorf("%w for key %q: expected %T but got %T", ErrSessionKeyNotFound, key, res, raw)
	}

	if value, ok := raw.(T); ok {
		return value, nil
	}

	// Session drivers may serialize values before persistence. Re-decode the
	// stored representation so JSON-compatible values retain their API type.
	encoded, err := json.Marshal(raw)
	if err != nil {
		return res, fmt.Errorf("%w for key %q: %w", ErrSessionInvalidValueType, key, err)
	}

	if err := json.Unmarshal(encoded, &res); err == nil {
		return res, nil
	}

	return res, fmt.Errorf("%w for key %q", ErrSessionInvalidValueType, key)
}

// Put stores a value in the session associated with the given key.
// This operation marks the session as changed.
//
// WARNING: When storing authentication-related state, callers MUST
// call [Session.Regenerate] immediately after to prevent session
// fixation attacks.
//
// Example:
//
//	session.Put("user_id", 42)
//	session.Regenerate()
func (session *Session) Put[T any](key string, value T) {
	session.mutex.Lock()
	defer session.mutex.Unlock()

	session.storage[key] = value
	session.changed = true
}

// Delete removes a value from the session storage by key.
// This operation marks the session as changed.
func (session *Session) Delete(key string) {
	session.mutex.Lock()
	defer session.mutex.Unlock()

	delete(session.storage, key)
	session.changed = true
}

// Extend updates the session's expiration time. This operation
// marks the session as changed.
func (session *Session) Extend(expiresAt time.Time) {
	session.mutex.Lock()
	defer session.mutex.Unlock()

	session.expiresAt = expiresAt
	session.changed = true
}

// Regenerate generates a new cryptographically random session ID. The original
// session ID is preserved for cleanup and this operation marks the session as
// changed. When using middleware.Session, the middleware persists the new
// session, deletes the old record, and issues the replacement cookie before
// response headers are written.
//
// WARNING: This method MUST be called after any authentication
// state change (login, logout, privilege escalation).
//
// Example:
//
//	session.Regenerate()
func (session *Session) Regenerate() {
	id := generateSessionID()

	session.mutex.Lock()
	defer session.mutex.Unlock()

	session.id = id
	session.changed = true
}

// Clear removes all data from the session while maintaining the
// session itself. This operation marks the session as changed.
func (session *Session) Clear() {
	session.mutex.Lock()
	defer session.mutex.Unlock()

	clear(session.storage)
	session.changed = true
}

// CreatedAt returns the absolute time the session was first created.
func (session *Session) CreatedAt() time.Time {
	session.mutex.Lock()
	defer session.mutex.Unlock()

	return session.createdAt
}

// ExpiresAt returns the time at which the session will expire.
func (session *Session) ExpiresAt() time.Time {
	session.mutex.Lock()
	defer session.mutex.Unlock()

	return session.expiresAt
}

// HasExpired returns true if the current time is past the session's
// expiration time.
func (session *Session) HasExpired() bool {
	session.mutex.Lock()
	defer session.mutex.Unlock()

	return time.Now().After(session.expiresAt)
}

// ExpiresSoon returns true if the session will expire within the
// specified duration from the current time.
func (session *Session) ExpiresSoon(delta time.Duration) bool {
	session.mutex.Lock()
	defer session.mutex.Unlock()

	now := time.Now()
	warningTime := session.expiresAt.Add(-delta)

	return now.After(warningTime) && now.Before(session.expiresAt)
}

// HasChanged returns true if the session data has been modified
// since it was loaded.
func (session *Session) HasChanged() bool {
	session.mutex.Lock()
	defer session.mutex.Unlock()

	return session.changed
}

// HasRegenerated returns true if the session has been regenerated
// (i.e., the session ID has changed).
func (session *Session) HasRegenerated() bool {
	session.mutex.Lock()
	defer session.mutex.Unlock()

	return session.id != session.originalID
}

// MarkAsUnchanged resets the change tracking flag, preventing
// the session from being persisted on the current request.
func (session *Session) MarkAsUnchanged() {
	session.mutex.Lock()
	defer session.mutex.Unlock()

	session.changed = false
}
