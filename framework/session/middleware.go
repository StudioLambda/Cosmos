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

	// TTL is the total lifetime of a session from creation or renewal.
	TTL time.Duration

	// MaxLifetime is the absolute maximum duration a session may
	// exist from its initial creation, regardless of activity.
	// A zero value disables the absolute lifetime check.
	// Default: 24 hours.
	MaxLifetime time.Duration

	// ExpirationDelta is the remaining time threshold at which
	// an active session is automatically extended by a full TTL.
	ExpirationDelta time.Duration

	// Key is the context key under which the session is stored.
	Key any

	// ErrorHandler is an optional callback invoked when internal
	// session operations fail. When nil, errors are silently discarded.
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

	// DefaultMaxLifetime is the default absolute maximum session age.
	DefaultMaxLifetime = 24 * time.Hour
)

// expectedSessionIDLength is the expected length of a valid session ID.
const expectedSessionIDLength = 43

// validSessionIDPattern matches exactly 43 base64url characters.
var validSessionIDPattern = regexp.MustCompile(
	`^[A-Za-z0-9_-]{43}$`,
)

// validSessionID reports whether the given ID has the expected format.
func validSessionID(id string) bool {
	if len(id) != expectedSessionIDLength {
		return false
	}

	return validSessionIDPattern.MatchString(id)
}

// currentSession loads an existing session from the cookie-provided
// ID or creates a fresh one when no valid session is found.
func currentSession(r *http.Request, driver contract.SessionDriver, options MiddlewareOptions) (*contract.Session, error) {
	id := request.CookieValue(r, options.Name)

	if id != "" && validSessionID(id) {
		if session, err := driver.Get(r.Context(), id); err == nil {
			if session.HasExpired() {
				_ = driver.Delete(r.Context(), id)

				return contract.NewSession(time.Now().Add(options.TTL), map[string]any{})
			}

			if options.MaxLifetime > 0 && time.Since(session.CreatedAt()) >= options.MaxLifetime {
				_ = driver.Delete(r.Context(), id)

				return contract.NewSession(time.Now().Add(options.TTL), map[string]any{})
			}

			session.MarkAsUnchanged()

			return session, nil
		}
	}

	return contract.NewSession(time.Now().Add(options.TTL), map[string]any{})
}

// withDefaults returns a copy of the options with secure defaults
// applied to any zero-valued fields.
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
func reportError(options MiddlewareOptions, err error) {
	if err != nil && options.ErrorHandler != nil {
		options.ErrorHandler(err)
	}
}

// MiddlewareWith returns a session middleware configured with the
// given driver and options.
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
// sensible defaults.
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
