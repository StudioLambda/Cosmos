package middleware

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/studiolambda/cosmos/contract"
	"github.com/studiolambda/cosmos/framework"
	"github.com/studiolambda/cosmos/framework/request"
	"github.com/studiolambda/cosmos/framework/session"
)

type SessionOptions struct {
	Name            string
	Path            string
	Domain          string
	Secure          bool
	SameSite        http.SameSite
	Partitioned     bool
	TTL             time.Duration
	ExpirationDelta time.Duration
}

const DefaultSessionCookie = "cosmos.session"
const DefaultSessionExpirationDelta = 15 * time.Minute
const DefaultSessionTTL = 2 * time.Hour

func currentSession(r *http.Request, driver contract.SessionDriver, options SessionOptions) (contract.Session, error) {
	id := request.CookieValue(r, options.Name)

	if id != "" {
		if s, err := driver.Get(r.Context(), id); err == nil {
			return s, nil
		}
	}

	return session.NewSession(time.Now().Add(options.TTL), map[string]any{})
}

func SessionWith(driver contract.SessionDriver, options SessionOptions) framework.Middleware {
	return func(next framework.Handler) framework.Handler {
		return func(w http.ResponseWriter, r *http.Request) error {
			s, err := currentSession(r, driver, options)

			if err != nil {
				return err
			}

			ctx := context.WithValue(r.Context(), session.Key, s)
			err = next(w, r.WithContext(ctx))

			if s.HasExpired() {
				if e := s.Regenerate(time.Now().Add(options.TTL)); e != nil {
					return errors.Join(e, err)
				}
			}

			if s.ExpiresSoon(options.ExpirationDelta) {
				s.Extend(time.Now().Add(options.TTL))
			}

			if s.HasRegenerated() {
				// Original session may have died during the execution of this
				// request. We don't really need to handle the error of this as
				// it's just active cleanup before ttl hits.
				_ = driver.Delete(r.Context(), s.OriginalSessionID())
			}

			if s.HasChanged() {
				expiration := time.Until(s.ExpiresAt())

				if e := driver.Save(r.Context(), s, expiration); e != nil {
					return errors.Join(e, err)
				}

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

			return err
		}
	}
}

func Session(driver contract.SessionDriver) framework.Middleware {
	return SessionWith(driver, SessionOptions{
		Name:            DefaultSessionCookie,
		Path:            "/",
		Domain:          "",
		Secure:          true,
		SameSite:        http.SameSiteLaxMode,
		Partitioned:     false,
		TTL:             DefaultSessionTTL,
		ExpirationDelta: DefaultSessionExpirationDelta,
	})
}
