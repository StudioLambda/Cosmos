package contract

import (
	"context"
	"net/http"
)

// Authenticatable is the contract that any user struct (e.g., User, Admin)
// must implement to be used in the authentication system. It ensures the
// authentication logic can access the user's unique identifier and password hash.
type Authenticatable interface {
	// AuthID returns the unique identifier for the user (e.g., primary key).
	AuthID() []byte

	// AuthPassword returns the hashed password for credential validation.
	AuthPassword() []byte
}

// Guard defines the strategy for authenticating and managing a user's session
// or state over HTTP. All methods are request-aware, reading from the Request
// (*http.Request) and potentially writing to the ResponseWriter (http.ResponseWriter).
type AuthGuard[User Authenticatable] interface {
	// User checks the request (e.g., looks for a valid cookie/header)
	// and returns the authenticated user if found. It does not perform an explicit login.
	User(r *http.Request) (User, error)

	// Check is a simplified method that validates existence. It typically
	// calls User() and returns true if the user is found and no error occurs.
	Check(r *http.Request) (bool, error)

	// Attempt tries to log in using raw credentials (identifier and password).
	// It relies on the UserProvider to find and validate the user, and on
	// successful validation, it calls Login() to establish the session/state.
	// It writes authentication data (e.g., cookie, token) to the ResponseWriter.
	Attempt(w http.ResponseWriter, r *http.Request, identifier []byte, password []byte) (User, error)

	// Login manually marks a specific user as authenticated. This is typically
	// used after registration or manual checks. It writes authentication
	// data (e.g., cookie, token) to the ResponseWriter.
	Login(w http.ResponseWriter, r *http.Request, user User) error

	// Logout invalidates the current user's session/state. It removes
	// authentication data (e.g., clears cookie, revokes token) from the ResponseWriter.
	Logout(w http.ResponseWriter, r *http.Request) error
}

// Provider defines the contract for retrieving and validating user data from
// a persistence layer (e.g., database, cache, LDAP). It is HTTP-agnostic.
// Concrete Guards will depend on a concrete Provider implementation.
type AuthProvider[User Authenticatable] interface {
	// UserByID retrieves a user based on their unique authentication identifier (AuthID).
	UserByID(ctx context.Context, id []byte) (User, error)

	// UserByCredentials retrieves a user based on a non-ID credential
	// (e.g., username, email) used for the initial login attempt.
	RetrieveByCredentials(ctx context.Context, identifier []byte) (User, error)

	// Validate checks the raw, unhashed password against the user's stored
	// hashed password (AuthPassword). The implementation handles hashing
	// and comparison logic.
	Validate(ctx context.Context, user User, password []byte) bool
}
