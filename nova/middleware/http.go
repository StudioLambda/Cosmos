package middleware

import (
	"net/http"

	"github.com/studiolambda/cosmos/nova"
)

func HTTP(middleware func(http.Handler) http.Handler) nova.Middleware {
	return func(next nova.Handler) nova.Handler {
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
