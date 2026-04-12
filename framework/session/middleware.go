package session

import (
	"context"
	"net/http"
	"regexp"
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

	// TTL is the total lifetime of a session from creation or
	// renewal.
	TTL time.Duration

	// MaxLifetime is the absolute maximum duration a session may
	// exist from its initial creation, regardless of activity.
	// When a session exceeds this age, it is force-regenerated
	// on the next request. This prevents indefinite session
	// extension through continued use.
	//
	// A zero value disables the absolute lifetime check.
	// Default: 24 hours.
	MaxLifetime time.Duration

	// ExpirationDelta is the remaining time threshold at which
	// an active session is automatically extended by a full TTL.
	ExpirationDelta time.Duration

	// Key is the context key under which the session is stored.
	// Defaults to contract.SessionKey when using Middleware.
	Key any

	// ErrorHandler is an optional callback invoked when internal
	// session operations (save, delete, regenerate) fail. When
	// nil, errors from these operations are silently discarded
	// to preserve backward compatibility.
	//
	// Callers should use this to log or monitor session driver
	// errors in production.
	ErrorHandler func(error)
}

const (
	// DefaultCookie is the default cookie name for sessions.
	DefaultCookie = "cosmos.session"

	// DefaultExpirationDelta is the default remaining-time threshold
	// that triggers automatic session extension.
	DefaultExpirationDelta = 15 * time.Minute

	// DefaultTTL is the default total session lifetime.
	DefaultTTL = 2 * time.Hour

	// DefaultMaxLifetime is the default absolute maximum session
	// age from initial creation. After this duration, the session
	// is force-regenerated regardless of activity.
	DefaultMaxLifetime = 24 * time.Hour
)

// expectedSessionIDLength is the expected length of a valid
// session ID string. A 32-byte random value encoded with
// base64url (no padding) produces exactly 43 characters.
const expectedSessionIDLength = 43

// validSessionIDPattern matches exactly 43 base64url characters
// (A-Z, a-z, 0-9, hyphen, underscore) with no padding.
var validSessionIDPattern = regexp.MustCompile(
	`^[A-Za-z0-9_-]{43}$`,
)

// validSessionID reports whether the given ID has the expected
// format for a session identifier: exactly 43 characters using
// only the base64url alphabet. This prevents cache key injection
// and avoids unnecessary cache lookups for malformed IDs.
func validSessionID(id string) bool {
	if len(id) != expectedSessionIDLength {
		return false
	}

	return validSessionIDPattern.MatchString(id)
}

// currentSession loads an existing session from the cookie-provided
// ID or creates a fresh one when no valid session is found. The
// cookie value is validated against the expected session ID format
// before performing a cache lookup. Expired sessions and sessions
// that exceed MaxLifetime are deleted from the driver and replaced
// with a fresh session to prevent stale auth data from reaching
// the handler.
func currentSession(r *http.Request, driver contract.SessionDriver, options MiddlewareOptions) (contract.Session, error) {
	id := request.CookieValue(r, options.Name)

	if id != "" && validSessionID(id) {
		if session, err := driver.Get(r.Context(), id); err == nil {
			if session.HasExpired() {
				_ = driver.Delete(r.Context(), id)

				return NewSession(time.Now().Add(options.TTL), map[string]any{})
			}

			if options.MaxLifetime > 0 && time.Since(session.CreatedAt()) >= options.MaxLifetime {
				_ = driver.Delete(r.Context(), id)

				return NewSession(time.Now().Add(options.TTL), map[string]any{})
			}

			session.MarkAsUnchanged()

			return session, nil
		}
	}

	return NewSession(time.Now().Add(options.TTL), map[string]any{})
}

// withDefaults returns a copy of the options with secure defaults
// applied to any zero-valued fields. This ensures that callers who
// construct a partial MiddlewareOptions still get sensible cookie
// attributes (SameSite=Lax, Path="/") and sensible session
// parameters (DefaultCookie name, DefaultTTL, DefaultMaxLifetime,
// DefaultExpirationDelta, contract.SessionKey). The Secure flag
// is NOT defaulted so that callers can set it to false for local
// development; use [Middleware] for secure-by-default behaviour.
func (options MiddlewareOptions) withDefaults() MiddlewareOptions {
	if options.Name == "" {
		options.Name = DefaultCookie
	}

	if options.Path == "" {
		options.Path = "/"
	}

	if options.SameSite == 0 {
		options.SameSite = http.SameSiteLaxMode
	}

	if options.TTL == 0 {
		options.TTL = DefaultTTL
	}

	if options.ExpirationDelta == 0 {
		options.ExpirationDelta = DefaultExpirationDelta
	}

	if options.MaxLifetime == 0 {
		options.MaxLifetime = DefaultMaxLifetime
	}

	if options.Key == nil {
		options.Key = contract.SessionKey
	}

	return options
}

// reportError invokes the configured error handler if set.
// When no handler is configured, the error is silently discarded
// to preserve backward compatibility.
func reportError(options MiddlewareOptions, err error) {
	if err != nil && options.ErrorHandler != nil {
		options.ErrorHandler(err)
	}
}

// MiddlewareWith returns a session middleware configured with the
// given driver and options. It loads or creates a session per
// request, attaches it to the context, and persists changes via a
// BeforeWriteHeader hook that handles expiration, regeneration,
// and cookie updates. Zero-valued option fields are replaced with
// secure defaults via withDefaults.
//
// When MaxLifetime is set and the session's age (measured from
// its CreatedAt timestamp) exceeds that duration, the session is
// force-regenerated to prevent indefinite extension.
func MiddlewareWith(driver contract.SessionDriver, options MiddlewareOptions) framework.Middleware {
	options = options.withDefaults()

	return func(next framework.Handler) framework.Handler {
		return func(w http.ResponseWriter, r *http.Request) error {
			session, err := currentSession(r, driver, options)

			if err != nil {
				return err
			}

			hooks := request.Hooks(r)
			hooks.BeforeWriteHeader(func(w http.ResponseWriter, status int) {
				saveCtx := context.WithoutCancel(r.Context())

				if options.MaxLifetime > 0 {
					age := time.Since(session.CreatedAt())

					if age >= options.MaxLifetime {
						reportError(
							options,
							session.Regenerate(),
						)
						session.Extend(time.Now().Add(options.TTL))
					}
				}

				if session.HasExpired() {
					reportError(options, session.Regenerate())
					session.Extend(time.Now().Add(options.TTL))
				}

				if session.ExpiresSoon(options.ExpirationDelta) {
					session.Extend(time.Now().Add(options.TTL))
				}

				if session.HasRegenerated() {
					reportError(
						options,
						driver.Delete(
							saveCtx,
							session.OriginalSessionID(),
						),
					)
				}

				if session.HasChanged() {
					ttl := time.Until(session.ExpiresAt())

					if err := driver.Save(saveCtx, session, ttl); err != nil {
						reportError(options, err)

						return
					}

					http.SetCookie(w, &http.Cookie{
						Name:        options.Name,
						Value:       session.SessionID(),
						Path:        options.Path,
						Domain:      options.Domain,
						Expires:     session.ExpiresAt(),
						MaxAge:      int(ttl.Seconds()),
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
// sensible defaults: cookie name "cosmos.session", 2-hour TTL,
// 24-hour max lifetime, 15-minute expiration delta, secure +
// HttpOnly + SameSite=Lax, stored under contract.SessionKey in
// the request context.
func Middleware(driver contract.SessionDriver) framework.Middleware {
	return MiddlewareWith(driver, MiddlewareOptions{
		Name:            DefaultCookie,
		Path:            "/",
		Domain:          "",
		Secure:          true,
		SameSite:        http.SameSiteLaxMode,
		Partitioned:     false,
		TTL:             DefaultTTL,
		MaxLifetime:     DefaultMaxLifetime,
		ExpirationDelta: DefaultExpirationDelta,
		Key:             contract.SessionKey,
	})
}
