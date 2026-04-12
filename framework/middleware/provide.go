package middleware

import (
	"context"
	"net/http"

	"github.com/studiolambda/cosmos/framework"
)

// ProvideFunc is the callback signature for ProvideWith. It receives
// the current request and returns a new context carrying the injected
// value, or an error if the value cannot be resolved.
type ProvideFunc = func(w http.ResponseWriter, r *http.Request) (context.Context, error)

// Provide returns a middleware that injects a static key-value pair
// into the request context. The key should be an unexported type to
// avoid collisions. Downstream handlers retrieve the value with
// r.Context().Value(key).
func Provide(key any, val any) framework.Middleware {
	return ProvideWith(func(w http.ResponseWriter, r *http.Request) (context.Context, error) {
		return context.WithValue(r.Context(), key, val), nil
	})
}

// ProvideWith returns a middleware that injects a dynamically resolved
// value into the request context. The ProvideFunc is called on each
// request and may return an error to short-circuit the middleware chain.
func ProvideWith(f ProvideFunc) framework.Middleware {
	return func(next framework.Handler) framework.Handler {
		return func(w http.ResponseWriter, r *http.Request) error {
			ctx, err := f(w, r)

			if err != nil {
				return err
			}

			return next(w, r.WithContext(ctx))
		}
	}
}
