package middleware

import (
	"context"
	"net/http"

	"github.com/studiolambda/cosmos/framework"
)

type ProvideFunc = func(w http.ResponseWriter, r *http.Request) (context.Context, error)

func Provide(key any, val any) framework.Middleware {
	return ProvideWith(func(w http.ResponseWriter, r *http.Request) (context.Context, error) {
		return context.WithValue(r.Context(), key, val), nil
	})
}

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
