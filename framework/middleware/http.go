package middleware

import (
	"net/http"

	"github.com/studiolambda/cosmos/framework"
)

func HTTP(middleware func(http.Handler) http.Handler) framework.Middleware {
	return func(next framework.Handler) framework.Handler {
		return func(w http.ResponseWriter, r *http.Request) error {
			var captured error

			httpNext := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				captured = next(w, r)
			})

			middleware(httpNext).ServeHTTP(w, r)

			return captured
		}
	}
}
