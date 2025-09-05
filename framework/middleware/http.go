package middleware

import (
	"net/http"

	"github.com/studiolambda/cosmos/framework"
)

// HTTP creates an adapter that allows standard Go HTTP middleware to be used
// with Cosmos handlers. This enables integration with existing HTTP middleware
// libraries and patterns while preserving Cosmos's error-returning handler semantics.
//
// The adapter works by:
//   - Converting the Cosmos handler chain into a standard http.Handler
//   - Applying the provided HTTP middleware to the converted handler
//   - Capturing any errors returned by the Cosmos handlers
//   - Returning those errors through the Cosmos middleware chain
//
// This allows you to use third-party HTTP middleware (CORS, authentication, etc.)
// without losing the error handling benefits of Cosmos handlers.
//
// Parameters:
//   - middleware: A standard HTTP middleware function that takes and returns http.Handler
//
// Returns a Cosmos middleware that integrates the HTTP middleware into the handler chain.
//
// Example:
//
//	app.Use(HTTP(someThirdPartyMiddleware))
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
