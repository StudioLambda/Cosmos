package middleware

import (
	"context"
	"errors"
	"net/http"

	"github.com/studiolambda/cosmos/framework"
	"github.com/studiolambda/cosmos/problem"
)

// ErrorHandlerServeDev defines an interface for errors that can provide
// enhanced development-time error responses. When running in development mode,
// errors implementing this interface will use their ServeHTTPDev method
// instead of the standard error handling, allowing for richer debug information.
type ErrorHandlerServeDev interface {
	// ServeHTTPDev handles the error response in development mode,
	// typically providing additional debugging information.
	ServeHTTPDev(w http.ResponseWriter, r *http.Request)
}

// ErrorHandler creates middleware that provides centralized error handling
// for Cosmos applications. It captures HTTP status codes, logs server errors
// (5xx status codes), and provides different error handling strategies based
// on the environment and error type.
//
// The middleware supports:
//   - Development mode with enhanced error details via ErrorHandlerDevHandler
//   - Standard http.Handler error types for custom error responses
//   - Fallback to basic error responses for unhandled error types
//
// Parameters:
//   - isDev: Determines if the error handler is in development mode.
//
// Returns a middleware function that handles errors from downstream handlers.
func ErrorHandler(isDev bool) framework.Middleware {
	return func(next framework.Handler) framework.Handler {
		return func(w http.ResponseWriter, r *http.Request) error {
			if err := next(w, r); err != nil {
				// If we are in development, we can check if
				// the handler implement a development handler.
				// Those usually add more context to the error.
				if h, ok := err.(ErrorHandlerServeDev); isDev && ok {
					h.ServeHTTPDev(w, r)

					return nil
				}

				// We can check if the error can be directly
				// handled by using a [http.Handler], in the case
				// we'll simply handle it using ServeHTTP.
				if h, ok := err.(http.Handler); ok {
					h.ServeHTTP(w, r)

					return nil
				}

				status := http.StatusInternalServerError

				if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
					status = 499 // A non-standard status code: 499 Client Closed Request
				}

				if s, ok := err.(framework.HTTPStatus); ok {
					status = s.HTTPStatus()
				}

				problem.NewProblem(err, status).ServeHTTP(w, r)

				return nil
			}

			return nil
		}
	}
}
