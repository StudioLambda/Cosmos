package session

import (
	"context"
	"net/http"
	"time"

	"github.com/studiolambda/cosmos/contract"
	"github.com/studiolambda/cosmos/contract/request"
	"github.com/studiolambda/cosmos/framework"
)

// MiddlewareOptions configures the session middleware behaviour
// including cookie attributes, session lifetime, and the context
// key used to store the session in the request.
type MiddlewareOptions struct {
	// Name is the cookie name sent to the client.
	Name string

	// Path restricts the cookie to the given URL path prefix.
	Path string

	// Domain restricts the cookie to the given domain.
	Domain string

	// Secure marks the cookie for HTTPS-only transmission.
	Secure bool

	// SameSite controls cross-site cookie behaviour.
	SameSite http.SameSite

	// Partitioned enables the CHIPS partitioned cookie attribute.
	Partitioned bool

	// TTL is the total lifetime of a session from creation or renewal.
	TTL time.Duration

	// ExpirationDelta is the remaining time threshold at which an
	// active session is automatically extended by a full TTL.
	ExpirationDelta time.Duration

	// Key is the context key under which the session is stored.
	// Defaults to contract.SessionKey when using Middleware.
	Key any
}

const (
	// DefaultCookie is the default cookie name for sessions.
	DefaultCookie = "cosmos.session"

	// DefaultExpirationDelta is the default remaining-time threshold
	// that triggers automatic session extension.
	DefaultExpirationDelta = 15 * time.Minute

	// DefaultTTL is the default total session lifetime.
	DefaultTTL = 2 * time.Hour
)

// currentSession loads an existing session from the cookie-provided
// ID or creates a fresh one when no valid session is found.
func currentSession(r *http.Request, driver contract.SessionDriver, options MiddlewareOptions) (contract.Session, error) {
	id := request.CookieValue(r, options.Name)

	if id != "" {
		if session, err := driver.Get(r.Context(), id); err == nil {
			session.MarkAsUnchanged()

			return session, nil
		}
	}

	return NewSession(time.Now().Add(options.TTL), map[string]any{})
}

// MiddlewareWith returns a session middleware configured with the
// given driver and options. It loads or creates a session per request,
// attaches it to the context, and persists changes via a
// BeforeWriteHeader hook that handles expiration, regeneration,
// and cookie updates.
func MiddlewareWith(driver contract.SessionDriver, options MiddlewareOptions) framework.Middleware {
	return func(next framework.Handler) framework.Handler {
		return func(w http.ResponseWriter, r *http.Request) error {
			session, err := currentSession(r, driver, options)

			if err != nil {
				return err
			}

			hooks := request.Hooks(r)
			hooks.BeforeWriteHeader(func(w http.ResponseWriter, status int) {
				if session.HasExpired() {
					_ = session.Regenerate()
					session.Extend(time.Now().Add(options.TTL))
				}

				if session.ExpiresSoon(options.ExpirationDelta) {
					session.Extend(time.Now().Add(options.TTL))
				}

				if session.HasRegenerated() {
					_ = driver.Delete(r.Context(), session.OriginalSessionID())
				}

				if session.HasChanged() {
					ttl := time.Until(session.ExpiresAt())

					_ = driver.Save(r.Context(), session, ttl)

					http.SetCookie(w, &http.Cookie{
						Name:        options.Name,
						Value:       session.SessionID(), // Will contain the new id if regenerated
						Path:        options.Path,
						Domain:      options.Domain,
						Expires:     session.ExpiresAt(), // Will be prior date if expired.
						MaxAge:      int(ttl.Seconds()),  // Will be negative if expired.
						Secure:      options.Secure,
						HttpOnly:    true,
						SameSite:    options.SameSite,
						Partitioned: options.Partitioned,
					})
				}
			})

			ctx := context.WithValue(r.Context(), options.Key, session)

			return next(w, r.WithContext(ctx))
		}
	}
}

// Middleware returns a session middleware using the given driver and
// sensible defaults: cookie name "cosmos.session", 2-hour TTL, 15-minute
// expiration delta, secure + HttpOnly + SameSite=Lax, stored under
// contract.SessionKey in the request context.
func Middleware(driver contract.SessionDriver) framework.Middleware {
	return MiddlewareWith(driver, MiddlewareOptions{
		Name:            DefaultCookie,
		Path:            "/",
		Domain:          "",
		Secure:          true,
		SameSite:        http.SameSiteLaxMode,
		Partitioned:     false,
		TTL:             DefaultTTL,
		ExpirationDelta: DefaultExpirationDelta,
		Key:             contract.SessionKey,
	})
}
