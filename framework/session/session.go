package session

import (
	"crypto/rand"
	"encoding/base64"
	"sync"
	"time"
)

// Session represents an HTTP session with data storage, expiration
// management, and change tracking. It maintains the current and
// original session IDs, supports key-value storage, and tracks
// modifications for persistence decisions. Access to the session
// is protected by a mutex to ensure thread-safe operations.
type Session struct {
	// originalID is the session ID assigned when the session was
	// first created. This value remains constant even if the
	// session is regenerated.
	originalID string

	// id is the current session identifier. This may differ from
	// OriginalID if the session has been regenerated (e.g., for
	// security purposes after authentication).
	id string

	// createdAt records the absolute time the session was first
	// created. This timestamp is immutable and used by the
	// middleware to enforce a maximum absolute session lifetime,
	// preventing indefinite session extension through activity.
	createdAt time.Time

	// expiresAt is the time at which the session will expire
	// and be considered invalid.
	expiresAt time.Time

	// storage holds the session data as key-value pairs. Keys
	// are strings and values can be of any type.
	storage map[string]any

	// mutex protects concurrent access to the session fields.
	mutex sync.Mutex

	// changed tracks whether the session data has been modified
	// since it was loaded, indicating whether it needs to be
	// persisted.
	changed bool
}

// sessionIDLength is the number of random bytes used to generate
// a session ID. 32 bytes provides 256 bits of entropy, which is
// encoded to a 43-character base64url string.
const sessionIDLength = 32

// generateSessionID generates a cryptographically random session
// ID using crypto/rand and base64url encoding. The resulting ID
// is 43 characters long with 256 bits of entropy and contains no
// embedded timestamp, unlike UUID v7.
func generateSessionID() (string, error) {
	b := make([]byte, sessionIDLength)

	_, err := rand.Read(b)

	if err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(b), nil
}

// NewSession creates a new session with the specified expiration
// time and initial storage data. It generates a cryptographically
// random session ID for both the original and current session IDs.
// The session is marked as changed to ensure it is persisted on
// first save. It returns an error if ID generation fails.
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
		mutex:      sync.Mutex{},
		changed:    true, // this will make sure first time its saved
	}, nil
}

// SessionID returns the current session identifier. This may differ from the original
// session ID if the session has been regenerated.
func (session *Session) SessionID() string {
	session.mutex.Lock()
	defer session.mutex.Unlock()

	return session.id
}

// OriginalSessionID returns the session identifier that was assigned when the session
// was initially created. This value does not change even if the session is regenerated.
func (session *Session) OriginalSessionID() string {
	session.mutex.Lock()
	defer session.mutex.Unlock()

	return session.originalID
}

// Get retrieves a value from the session storage by key. It returns the value and a
// boolean indicating whether the key exists in the session.
func (session *Session) Get(key string) (any, bool) {
	session.mutex.Lock()
	defer session.mutex.Unlock()

	value, ok := session.storage[key]

	return value, ok
}

// Put stores a value in the session storage associated with the
// given key. If the key already exists, its value is overwritten.
// This operation marks the session as changed.
//
// WARNING: When storing authentication-related state (e.g.,
// user IDs, roles, or privilege levels), callers MUST call
// [Session.Regenerate] immediately after to prevent session
// fixation attacks. Failing to regenerate the session ID
// after authentication changes allows an attacker who knows
// the pre-authentication session ID to hijack the elevated
// session.
func (session *Session) Put(key string, value any) {
	session.mutex.Lock()
	defer session.mutex.Unlock()

	session.storage[key] = value
	session.changed = true
}

// Delete removes a value from the session storage by key. If the key does not exist,
// this operation is a no-op. This operation marks the session as changed.
func (session *Session) Delete(key string) {
	session.mutex.Lock()
	defer session.mutex.Unlock()

	delete(session.storage, key)
	session.changed = true
}

// Extend updates the session's expiration time to the specified time. This is useful
// for extending the session's lifetime during active use. This operation marks the
// session as changed.
func (session *Session) Extend(expiresAt time.Time) {
	session.mutex.Lock()
	defer session.mutex.Unlock()

	session.expiresAt = expiresAt
	session.changed = true
}

// Regenerate generates a new session ID and updates the expiration
// time. This is commonly used for security purposes such as
// preventing session fixation attacks after user authentication.
// The original session ID is preserved and can be retrieved via
// OriginalSessionID. This operation marks the session as changed.
// It returns an error if ID generation fails.
//
// WARNING: This method MUST be called after any authentication
// state change (login, logout, privilege escalation). Without
// regeneration, an attacker who obtained the session ID before
// authentication can use it to access the authenticated session.
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

// Clear removes all data from the session storage while keeping the session itself intact.
// The session ID and expiration time remain unchanged. This operation marks the session
// as changed.
func (session *Session) Clear() {
	session.mutex.Lock()
	defer session.mutex.Unlock()

	clear(session.storage)
	session.changed = true
}

// ExpiresAt returns the time at which the session will expire
// and be considered invalid.
func (session *Session) ExpiresAt() time.Time {
	session.mutex.Lock()
	defer session.mutex.Unlock()

	return session.expiresAt
}

// CreatedAt returns the absolute time the session was first
// created. This value never changes, even if the session is
// regenerated or extended. It is used by the middleware to
// enforce [MiddlewareOptions.MaxLifetime].
func (session *Session) CreatedAt() time.Time {
	session.mutex.Lock()
	defer session.mutex.Unlock()

	return session.createdAt
}

// HasExpired checks whether the session has already expired by comparing the current
// time against the expiration time.
func (session *Session) HasExpired() bool {
	session.mutex.Lock()
	defer session.mutex.Unlock()

	return time.Now().After(session.expiresAt)
}

// ExpiresSoon checks whether the session will expire within the specified duration
// from the current time. This is useful for triggering session renewal prompts or
// warnings before the session actually expires.
func (session *Session) ExpiresSoon(delta time.Duration) bool {
	session.mutex.Lock()
	defer session.mutex.Unlock()

	now := time.Now()
	warningTime := session.expiresAt.Add(-delta)

	return now.After(warningTime) && now.Before(session.expiresAt)
}

// HasChanged returns true if the session data has been modified since it was loaded.
// This is useful for determining whether the session needs to be persisted to storage.
func (session *Session) HasChanged() bool {
	session.mutex.Lock()
	defer session.mutex.Unlock()

	return session.changed
}

// HasRegenerated returns true if the session has been regenerated, meaning the current
// session ID differs from the original session ID. This is useful for determining whether
// an updated session identifier should be sent to the client.
func (session *Session) HasRegenerated() bool {
	session.mutex.Lock()
	defer session.mutex.Unlock()

	return session.id != session.originalID
}

// MarkAsUnchanged sets the session as if nothing has changed, therefore avoiding saving
// the session when the request finishes.
func (session *Session) MarkAsUnchanged() {
	session.mutex.Lock()
	defer session.mutex.Unlock()

	session.changed = false
}
