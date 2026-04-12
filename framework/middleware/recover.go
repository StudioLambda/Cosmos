package middleware

import (
	"encoding"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/studiolambda/cosmos/framework"
)

// recoverReadLimit is the maximum number of bytes that will be read
// from an io.Reader panic value. This prevents a malicious or
// excessively large reader from consuming unbounded memory during
// panic recovery.
const recoverReadLimit = 1 << 20 // 1 MB

// ErrRecoverUnexpected is the default error returned when a panic
// occurs but the recovered value cannot be converted to a meaningful
// error. This ensures that all panics result in a recoverable error
// state rather than crashing the application.
var ErrRecoverUnexpected = errors.New("an unexpected error occurred")

// ErrFailedRecovering is an error that's returned when the recover
// process failed.
var ErrFailedRecovering = errors.New(
	"failed recovering from unexpected error",
)

// defaultRecoverHandler converts recovered panic values into errors.
// It handles common panic value types and provides sensible error
// conversion:
//   - error types are joined with ErrRecoverUnexpected so that
//     only the sentinel is visible to clients
//   - string values are wrapped the same way
//   - fmt.Stringer implementations use their String() method
//   - io.Reader values are capped at recoverReadLimit bytes
//   - all other types are formatted and joined with
//     ErrRecoverUnexpected
//
// The original panic detail is always reachable via
// [errors.Unwrap] for logging or debugging purposes, but
// the outer error seen by clients is the safe sentinel.
func defaultRecoverHandler(value any) error {
	switch recovered := value.(type) {
	case error:
		return errors.Join(ErrRecoverUnexpected, recovered)
	case string:
		return errors.Join(
			ErrRecoverUnexpected,
			errors.New(recovered),
		)
	case fmt.Stringer:
		return errors.Join(
			ErrRecoverUnexpected,
			errors.New(recovered.String()),
		)
	case io.Reader:
		body, err := io.ReadAll(
			io.LimitReader(recovered, recoverReadLimit),
		)

		if err != nil {
			return errors.Join(ErrFailedRecovering, err)
		}

		return errors.Join(
			ErrRecoverUnexpected,
			errors.New(string(body)),
		)
	case encoding.TextMarshaler:
		text, err := recovered.MarshalText()

		if err != nil {
			return errors.Join(ErrFailedRecovering, err)
		}

		return errors.Join(
			ErrRecoverUnexpected,
			errors.New(string(text)),
		)
	default:
		return errors.Join(
			ErrRecoverUnexpected,
			fmt.Errorf("%+v", recovered),
		)
	}
}

// Recover creates middleware that recovers from panics in HTTP
// handlers and converts them to errors using the default recovery
// handler. This prevents panics from crashing the entire
// application and allows them to be handled through the normal
// error handling middleware chain.
//
// The default handler wraps all recovered values with
// ErrRecoverUnexpected so that internal details are never exposed
// to HTTP clients. The original value remains accessible via
// [errors.Unwrap] for server-side logging.
func Recover() framework.Middleware {
	return RecoverWith(defaultRecoverHandler)
}

// RecoverWith creates middleware that recovers from panics using
// a custom recovery handler function. This allows applications to
// implement custom panic handling logic, such as specialised
// logging, error formatting, or panic value processing.
//
// The custom handler receives the raw panic value and must return
// an error that will be passed through the normal error handling
// chain. Use [errors.As] with an interface containing a
// Stack() []byte method to access the stack trace.
func RecoverWith(handler func(value any) error) framework.Middleware {
	return func(next framework.Handler) framework.Handler {
		return func(
			w http.ResponseWriter,
			r *http.Request,
		) (err error) {
			defer func() {
				if e := recover(); e != nil {
					err = handler(e)
				}
			}()

			return next(w, r)
		}
	}
}
