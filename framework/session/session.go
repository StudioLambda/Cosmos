package session

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

// Session represents an HTTP session with data storage, expiration management, and change tracking.
// It maintains the current and original session IDs, supports key-value storage, and tracks
// modifications for persistence decisions. Access to the session is protected by a mutex
// to ensure thread-safe operations.
type Session struct {
	// originalID is the session ID assigned when the session was first created.
	// This value remains constant even if the session is regenerated.
	originalID string

	// id is the current session identifier. This may differ from OriginalID if the session
	// has been regenerated (e.g., for security purposes after authentication).
	id string

	// expiration is the time at which the session will expire and be considered invalid.
	expiration time.Time

	// storage holds the session data as key-value pairs. Keys are strings and values
	// can be of any type.
	storage map[string]any

	// mutex protects concurrent access to the session fields.
	mutex sync.Mutex

	// changed tracks whether the session data has been modified since it was loaded,
	// indicating whether it needs to be persisted.
	changed bool
}

// NewSession creates a new session with the specified expiration time and initial storage data.
// It generates a new UUID v7 for both the original and current session IDs.
// The session is marked as changed to ensure it is persisted on first save.
// It returns an error if the UUID generation fails.
func NewSession(expiresAt time.Time, storage map[string]any) (*Session, error) {
	id, err := uuid.NewV7()

	if err != nil {
		return nil, err
	}

	return &Session{
		originalID: id.String(),
		id:         id.String(),
		expiration: expiresAt,
		storage:    storage,
		mutex:      sync.Mutex{},
		changed:    true, // this will make sure first time its saved
	}, nil
}

// SessionID returns the current session identifier. This may differ from the original
// session ID if the session has been regenerated.
func (s *Session) SessionID() string {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.id
}

// OriginalSessionID returns the session identifier that was assigned when the session
// was initially created. This value does not change even if the session is regenerated.
func (s *Session) OriginalSessionID() string {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.originalID
}

// Get retrieves a value from the session storage by key. It returns the value and a
// boolean indicating whether the key exists in the session.
func (s *Session) Get(key string) (any, bool) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	v, ok := s.storage[key]

	return v, ok
}

// Put stores a value in the session storage associated with the given key. If the key
// already exists, its value is overwritten. This operation marks the session as changed.
func (s *Session) Put(key string, value any) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.storage[key] = value
	s.changed = true
}

// Delete removes a value from the session storage by key. If the key does not exist,
// this operation is a no-op. This operation marks the session as changed.
func (s *Session) Delete(key string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	delete(s.storage, key)
	s.changed = true
}

// Extend updates the session's expiration time to the specified time. This is useful
// for extending the session's lifetime during active use. This operation marks the
// session as changed.
func (s *Session) Extend(expiresAt time.Time) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.expiration = expiresAt
	s.changed = true
}

// Regenerate generates a new session ID and updates the expiration time. This is
// commonly used for security purposes such as preventing session fixation attacks
// after user authentication. The original session ID is preserved and can be retrieved
// via OriginalSessionID. This operation marks the session as changed. It returns an
// error if ID generation fails.
func (s *Session) Regenerate() error {
	id, err := uuid.NewV7()

	if err != nil {
		return err
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.id = id.String()
	s.changed = true

	return nil
}

// Clear removes all data from the session storage while keeping the session itself intact.
// The session ID and expiration time remain unchanged. This operation marks the session
// as changed.
func (s *Session) Clear() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	clear(s.storage)
	s.changed = true
}

// ExpiresAt returns the time at which the session will expire and be considered invalid.
func (s *Session) ExpiresAt() time.Time {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.expiration
}

// HasExpired checks whether the session has already expired by comparing the current
// time against the expiration time.
func (s *Session) HasExpired() bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return time.Now().After(s.expiration)
}

// ExpiresSoon checks whether the session will expire within the specified duration
// from the current time. This is useful for triggering session renewal prompts or
// warnings before the session actually expires.
func (s *Session) ExpiresSoon(delta time.Duration) bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	now := time.Now()
	warningTime := s.expiration.Add(-delta)

	return now.After(warningTime) && now.Before(s.expiration)
}

// HasChanged returns true if the session data has been modified since it was loaded.
// This is useful for determining whether the session needs to be persisted to storage.
func (s *Session) HasChanged() bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.changed
}

// HasRegenerated returns true if the session has been regenerated, meaning the current
// session ID differs from the original session ID. This is useful for determining whether
// an updated session identifier should be sent to the client.
func (s *Session) HasRegenerated() bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.id != s.originalID
}

// MarkAsUnchanged sets the session as if nothing has changed, therefore avoiding saving
// the session when the request finishes.
func (s *Session) MarkAsUnchanged() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.changed = false
}
