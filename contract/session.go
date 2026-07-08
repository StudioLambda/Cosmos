package contract

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"maps"
	"sync"
	"time"
)

// sessionKey is a private type used as a context key to avoid collisions.
type sessionKey struct{}

// SessionKey is the context key used to store and retrieve the session from a context.Context.
var SessionKey = sessionKey{}

// sessionIDLength is the number of random bytes used to generate
// a session ID. 32 bytes provides 256 bits of entropy.
const sessionIDLength = 32

// generateSessionID generates a cryptographically random session
// ID using crypto/rand and base64url encoding (43 characters).
func generateSessionID() (string, error) {
	b := make([]byte, sessionIDLength)

	_, err := rand.Read(b)

	if err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(b), nil
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

// NewSession creates a new session with the specified expiration
// time and initial storage data. It generates a cryptographically
// random session ID. The session is marked as changed to ensure it
// is persisted on first save. Returns an error if ID generation fails.
func NewSession(expiresAt time.Time, storage map[string]any) (*Session, error) {
	id, err := generateSessionID()

	if err != nil {
		return nil, err
	}

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

	result := make(map[string]any, len(session.storage))

	maps.Copy(result, session.storage)

	return result
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
func (session *Session) Get(key string) (any, bool) {
	session.mutex.Lock()
	defer session.mutex.Unlock()

	value, ok := session.storage[key]

	return value, ok
}

// Put stores a value in the session associated with the given key.
// This operation marks the session as changed.
//
// WARNING: When storing authentication-related state, callers MUST
// call [Session.Regenerate] immediately after to prevent session
// fixation attacks.
func (session *Session) Put(key string, value any) {
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

// Regenerate generates a new cryptographically random session ID.
// The original session ID is preserved for cleanup. This operation
// marks the session as changed.
//
// WARNING: This method MUST be called after any authentication
// state change (login, logout, privilege escalation).
func (session *Session) Regenerate() error {
	id, err := generateSessionID()

	if err != nil {
		return err
	}

	session.mutex.Lock()
	defer session.mutex.Unlock()

	session.id = id
	session.changed = true

	return nil
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
