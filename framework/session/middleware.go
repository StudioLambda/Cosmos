package session

import (
	"context"
	"net/http"
	"time"

	"github.com/studiolambda/cosmos/contract"
	"github.com/studiolambda/cosmos/contract/request"
	"github.com/studiolambda/cosmos/framework"
)

type MiddlewareOptions struct {
	Name            string
	Path            string
	Domain          string
	Secure          bool
	SameSite        http.SameSite
	Partitioned     bool
	TTL             time.Duration
	ExpirationDelta time.Duration
	Key             any
}

const DefaultCookie = "cosmos.session"
const DefaultExpirationDelta = 15 * time.Minute
const DefaultTTL = 2 * time.Hour

func currentSession(r *http.Request, driver contract.SessionDriver, options MiddlewareOptions) (contract.Session, error) {
	id := request.CookieValue(r, options.Name)

	if id != "" {
		if s, err := driver.Get(r.Context(), id); err == nil {
			s.MarkAsUnchanged()

			return s, nil
		}
	}

	return NewSession(time.Now().Add(options.TTL), map[string]any{})
}

func MiddlewareWith(driver contract.SessionDriver, options MiddlewareOptions) framework.Middleware {
	return func(next framework.Handler) framework.Handler {
		return func(w http.ResponseWriter, r *http.Request) error {
			s, err := currentSession(r, driver, options)

			if err != nil {
				return err
			}

			hooks := request.Hooks(r)
			hooks.BeforeWriteHeader(func(w http.ResponseWriter, status int) {
				if s.HasExpired() {
					_ = s.Regenerate()
					s.Extend(time.Now().Add(options.TTL))
				}

				if s.ExpiresSoon(options.ExpirationDelta) {
					s.Extend(time.Now().Add(options.TTL))
				}

				if s.HasRegenerated() {
					_ = driver.Delete(r.Context(), s.OriginalSessionID())
				}

				if s.HasChanged() {
					expiration := time.Until(s.ExpiresAt())

					_ = driver.Save(r.Context(), s, expiration)

					http.SetCookie(w, &http.Cookie{
						Name:        options.Name,
						Value:       s.SessionID(), // Will contain the new id if regenerated
						Path:        options.Path,
						Domain:      options.Domain,
						Expires:     s.ExpiresAt(),             // Will be prior date if expired.
						MaxAge:      int(expiration.Seconds()), // Will be negative if expired.
						Secure:      options.Secure,
						HttpOnly:    true,
						SameSite:    options.SameSite,
						Partitioned: options.Partitioned,
					})
				}
			})

			ctx := context.WithValue(r.Context(), options.Key, s)

			return next(w, r.WithContext(ctx))
		}
	}
}

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
