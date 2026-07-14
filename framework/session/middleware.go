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

// MiddlewareConfig configures the session middleware behaviour
// including cookie attributes, session lifetime, and the context
// key used to store the session in the request.
type MiddlewareConfig struct {
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
func currentSession(r *http.Request, driver contract.SessionDriver, config MiddlewareConfig) (*contract.Session, error) {
	id := request.CookieValue(r, config.Name)

	if id != "" && validSessionID(id) {
		if session, err := driver.Get(r.Context(), id); err == nil {
			if session.HasExpired() {
				_ = driver.Delete(r.Context(), id)

				return contract.NewSession(time.Now().Add(config.TTL), map[string]any{})
			}

			if config.MaxLifetime > 0 && time.Since(session.CreatedAt()) >= config.MaxLifetime {
				_ = driver.Delete(r.Context(), id)

				return contract.NewSession(time.Now().Add(config.TTL), map[string]any{})
			}

			session.MarkAsUnchanged()

			return session, nil
		}
	}

	return contract.NewSession(time.Now().Add(config.TTL), map[string]any{})
}

// withDefaults returns a copy of the config with secure defaults
// applied to any zero-valued fields.
func (config MiddlewareConfig) withDefaults() MiddlewareConfig {
	if config.Name == "" {
		config.Name = DefaultCookie
	}

	if config.Path == "" {
		config.Path = "/"
	}

	if config.SameSite == 0 {
		config.SameSite = http.SameSiteLaxMode
	}

	if config.TTL == 0 {
		config.TTL = DefaultTTL
	}

	if config.ExpirationDelta == 0 {
		config.ExpirationDelta = DefaultExpirationDelta
	}

	if config.MaxLifetime == 0 {
		config.MaxLifetime = DefaultMaxLifetime
	}

	if config.Key == nil {
		config.Key = contract.SessionKey
	}

	return config
}

// reportError invokes the configured error handler if set.
func reportError(config MiddlewareConfig, err error) {
	if err != nil && config.ErrorHandler != nil {
		config.ErrorHandler(err)
	}
}

// MiddlewareWith returns a session middleware configured with the
// given driver and configuration.
func MiddlewareWith(driver contract.SessionDriver, config MiddlewareConfig) framework.Middleware {
	config = config.withDefaults()

	return func(next framework.Handler) framework.Handler {
		return func(w http.ResponseWriter, r *http.Request) error {
			session, err := currentSession(r, driver, config)

			if err != nil {
				return err
			}

			hooks := request.Hooks(r)
			hooks.BeforeWriteHeader(func(w http.ResponseWriter, status int) {
				saveCtx := context.WithoutCancel(r.Context())

				if config.MaxLifetime > 0 {
					age := time.Since(session.CreatedAt())

					if age >= config.MaxLifetime {
						reportError(
							config,
							session.Regenerate(),
						)
						session.Extend(time.Now().Add(config.TTL))
					}
				}

				if session.HasExpired() {
					reportError(config, session.Regenerate())
					session.Extend(time.Now().Add(config.TTL))
				}

				if session.ExpiresSoon(config.ExpirationDelta) {
					session.Extend(time.Now().Add(config.TTL))
				}

				if session.HasRegenerated() {
					reportError(
						config,
						driver.Delete(
							saveCtx,
							session.OriginalSessionID(),
						),
					)
				}

				if session.HasChanged() {
					ttl := time.Until(session.ExpiresAt())

					if err := driver.Save(saveCtx, session, ttl); err != nil {
						reportError(config, err)

						return
					}

					http.SetCookie(w, &http.Cookie{
						Name:        config.Name,
						Value:       session.SessionID(),
						Path:        config.Path,
						Domain:      config.Domain,
						Expires:     session.ExpiresAt(),
						MaxAge:      int(ttl.Seconds()),
						Secure:      config.Secure,
						HttpOnly:    true,
						SameSite:    config.SameSite,
						Partitioned: config.Partitioned,
					})
				}
			})

			ctx := context.WithValue(r.Context(), config.Key, session)

			return next(w, r.WithContext(ctx))
		}
	}
}

// Middleware returns a session middleware using the given driver and
// sensible defaults.
func Middleware(driver contract.SessionDriver) framework.Middleware {
	return MiddlewareWith(driver, MiddlewareConfig{
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
