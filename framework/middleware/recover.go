package middleware

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/studiolambda/cosmos/framework"
)

// ErrRecoverUnexpectedError is the default error returned when a panic
// occurs but the recovered value cannot be converted to a meaningful error.
// This ensures that all panics result in a recoverable error state rather
// than crashing the application.
var ErrRecoverUnexpectedError = errors.New("an unexpected error occurred")

// defaultRecoverHandler converts recovered panic values into errors.
// It handles common panic value types and provides sensible error conversion:
//   - error types are returned as-is
//   - string values are wrapped in errors.New()
//   - fmt.Stringer implementations use their String() method
//   - all other types are formatted using fmt.Sprintf and joined with ErrRecoverUnexpectedError
//
// Parameters:
//   - value: The value recovered from a panic
//
// Returns an error representation of the panic value.
func defaultRecoverHandler(value any) error {
	switch r := value.(type) {
	case error:
		return r
	case string:
		return errors.New(r)
	case fmt.Stringer:
		return errors.New(r.String())
	default:
		return errors.Join(ErrRecoverUnexpectedError, fmt.Errorf("%+v", r))
	}
}

// Recover creates middleware that recovers from panics in HTTP handlers and
// converts them to errors using the default recovery handler. This prevents
// panics from crashing the entire application and allows them to be handled
// through the normal error handling middleware chain.
//
// The default handler converts common panic types (errors, strings, fmt.Stringers)
// into appropriate error values. For unknown types, it returns ErrRecoverUnexpectedError
// joined with a formatted representation of the panic value.
//
// Returns a middleware function that catches panics and converts them to errors.
func Recover() framework.Middleware {
	return RecoverWith(defaultRecoverHandler)
}

// RecoverWith creates middleware that recovers from panics using a custom
// recovery handler function. This allows applications to implement custom
// panic handling logic, such as specialized logging, error formatting, or
// panic value processing.
//
// The custom handler receives the raw panic value and must return an error
// that will be passed through the normal error handling chain.
//
// Parameters:
//   - handler: A function that converts panic values to errors
//
// Returns a middleware function that catches panics and processes them with the custom handler.
func RecoverWith(handler func(value any) error) framework.Middleware {
	return func(next framework.Handler) framework.Handler {
		return func(w http.ResponseWriter, r *http.Request) (err error) {
			defer func() {
				if e := recover(); e != nil {
					err = handler(e)
				}
			}()

			return next(w, r)
		}
	}
}
