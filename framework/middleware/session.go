package middleware

import (
	"context"
	"net/http"
	"regexp"
	"time"

	"github.com/studiolambda/cosmos/contract"
	"github.com/studiolambda/cosmos/contract/request"
	"github.com/studiolambda/cosmos/framework"
)

// SessionConfig configures the session middleware behaviour
// including cookie attributes, session lifetime, and the context
// key used to store the session in the request.
type SessionConfig struct {
	// Name is the cookie name sent to the client.
	Name string

	// Path restricts the cookie to the given URL path prefix.
	Path string

	// Domain restricts the cookie to the given domain.
	Domain string

	// Secure marks the cookie for HTTPS-only transmission. It is derived from
	// AllowInsecure during defaulting; cookies are secure unless explicitly
	// opted out of below.
	Secure bool

	// AllowInsecure explicitly permits cookies over HTTP. This is intended only
	// for local development; production deployments should leave it false.
	AllowInsecure bool

	// SameSite controls cross-site cookie behaviour. Supported values
	// are "lax", "strict", "none", and "default".
	SameSite string

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
}

// SessionRuntime holds runtime-only session middleware settings.
type SessionRuntime struct {
	// Key is the context key under which the session is stored.
	Key any

	// ErrorHandler is an optional callback invoked when internal
	// session operations fail. When nil, errors are silently discarded.
	ErrorHandler func(error)

	// PersistenceTimeout bounds session storage operations performed before
	// response headers are written. Zero defaults to five seconds.
	PersistenceTimeout time.Duration
}

// DefaultSessionConfig holds the default session middleware
// configuration.
var DefaultSessionConfig = SessionConfig{
	Name:            DefaultSessionCookie,
	Path:            "/",
	Domain:          "",
	Secure:          true,
	SameSite:        "lax",
	Partitioned:     false,
	TTL:             DefaultSessionTTL,
	MaxLifetime:     DefaultSessionMaxLifetime,
	ExpirationDelta: DefaultSessionExpirationDelta,
}

// DefaultSessionRuntime holds the default runtime settings for the
// session middleware.
var DefaultSessionRuntime = SessionRuntime{
	Key:                contract.SessionKey,
	PersistenceTimeout: 5 * time.Second,
}

const (
	// DefaultSessionCookie is the default cookie name for sessions.
	DefaultSessionCookie = "cosmos.session"

	// DefaultSessionExpirationDelta is the default remaining-time threshold
	// that triggers automatic session extension.
	DefaultSessionExpirationDelta = 15 * time.Minute

	// DefaultSessionTTL is the default total session lifetime.
	DefaultSessionTTL = 2 * time.Hour

	// DefaultSessionMaxLifetime is the default absolute maximum session age.
	DefaultSessionMaxLifetime = 24 * time.Hour
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
func currentSession(r *http.Request, driver contract.SessionDriver, config SessionConfig) (*contract.Session, error) {
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
func (config SessionConfig) withDefaults() SessionConfig {
	if config.Name == "" {
		config.Name = DefaultSessionCookie
	}

	if config.Path == "" {
		config.Path = "/"
	}

	if config.SameSite == "" {
		config.SameSite = DefaultSessionConfig.SameSite
	}

	if config.TTL == 0 {
		config.TTL = DefaultSessionTTL
	}

	if config.ExpirationDelta == 0 {
		config.ExpirationDelta = DefaultSessionExpirationDelta
	}

	if config.MaxLifetime == 0 {
		config.MaxLifetime = DefaultSessionMaxLifetime
	}

	config.Secure = !config.AllowInsecure

	return config
}

func (runtime SessionRuntime) withDefaults() SessionRuntime {
	if runtime.Key == nil {
		runtime.Key = DefaultSessionRuntime.Key
	}

	if runtime.PersistenceTimeout == 0 {
		runtime.PersistenceTimeout = DefaultSessionRuntime.PersistenceTimeout
	}

	return runtime
}

func sameSiteMode(value string) http.SameSite {
	switch value {
	case "default":
		return http.SameSiteDefaultMode
	case "none":
		return http.SameSiteNoneMode
	case "strict":
		return http.SameSiteStrictMode
	case "lax", "":
		return http.SameSiteLaxMode
	default:
		panic("session middleware: invalid SameSite value")
	}
}

// reportError invokes the configured error handler if set.
func reportError(runtime SessionRuntime, err error) {
	if err != nil && runtime.ErrorHandler != nil {
		runtime.ErrorHandler(err)
	}
}

// Session returns session middleware configured with the given
// driver and configuration.
func Session(driver contract.SessionDriver, config SessionConfig) framework.Middleware {
	return SessionWith(driver, config, SessionRuntime{})
}

// SessionWith returns session middleware configured with the
// given driver, configuration, and runtime options.
func SessionWith(driver contract.SessionDriver, config SessionConfig, runtime SessionRuntime) framework.Middleware {
	config = config.withDefaults()
	runtime = runtime.withDefaults()
	sameSite := sameSiteMode(config.SameSite)

	return func(next framework.Handler) framework.Handler {
		return func(w http.ResponseWriter, r *http.Request) error {
			session, err := currentSession(r, driver, config)

			if err != nil {
				return err
			}

			hooks := request.Hooks(r)
			hooks.BeforeWriteHeader(func(w http.ResponseWriter, status int) {
				saveCtx, cancel := context.WithTimeout(
					context.WithoutCancel(r.Context()),
					runtime.PersistenceTimeout,
				)
				defer cancel()

				if config.MaxLifetime > 0 {
					age := time.Since(session.CreatedAt())

					if age >= config.MaxLifetime {
						session.Regenerate()
						session.Extend(time.Now().Add(config.TTL))
					}
				}

				if session.HasExpired() {
					session.Regenerate()
					session.Extend(time.Now().Add(config.TTL))
				}

				if session.ExpiresSoon(config.ExpirationDelta) {
					session.Extend(time.Now().Add(config.TTL))
				}

				if session.HasRegenerated() {
					reportError(
						runtime,
						driver.Delete(
							saveCtx,
							session.OriginalSessionID(),
						),
					)
				}

				if session.HasChanged() {
					ttl := time.Until(session.ExpiresAt())

					if err := driver.Save(saveCtx, session, ttl); err != nil {
						reportError(runtime, err)

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
						SameSite:    sameSite,
						Partitioned: config.Partitioned,
					})
				}
			})

			ctx := context.WithValue(r.Context(), runtime.Key, session)

			return next(w, r.WithContext(ctx))
		}
	}
}
