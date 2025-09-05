package middleware

import (
	"log/slog"
	"net/http"

	"github.com/studiolambda/cosmos/framework"
)

// rawErrorHandlerWriter wraps an http.ResponseWriter to capture the HTTP status
// code that was written to the response. This allows middleware to inspect the
// status code after handlers have executed, which is useful for logging and
// error handling decisions.
type rawErrorHandlerWriter struct {
	http.ResponseWriter
	status int
}

// flushableErrorHandlerWriter extends rawErrorHandlerWriter to support the
// http.Flusher interface when the underlying ResponseWriter supports flushing.
// This enables streaming responses while still capturing status codes.
type flushableErrorHandlerWriter struct {
	*rawErrorHandlerWriter
	http.Flusher
}

// ErrorHandlerWriter extends http.ResponseWriter with the ability to retrieve
// the HTTP status code that was written to the response. This interface is
// used by the error handling middleware to make status-based decisions.
type ErrorHandlerWriter interface {
	http.ResponseWriter
	// Status returns the HTTP status code that was written to the response.
	// Returns http.StatusOK (200) by default if no status was explicitly set.
	Status() int
}

// ErrorHandlerDevHandler defines an interface for errors that can provide
// enhanced development-time error responses. When running in development mode,
// errors implementing this interface will use their ServeHTTPDev method
// instead of the standard error handling, allowing for richer debug information.
type ErrorHandlerDevHandler interface {
	// ServeHTTPDev handles the error response in development mode,
	// typically providing additional debugging information.
	ServeHTTPDev(w http.ResponseWriter, r *http.Request)
}

// ErrorHandlerOptions configures the behavior of the error handling middleware.
type ErrorHandlerOptions struct {
	// Logger is used to log errors with status codes >= 500. If nil,
	// a discard logger will be used to prevent panics.
	Logger *slog.Logger
	// IsDev enables development mode, which uses ErrorHandlerDevHandler
	// when available to provide enhanced error information.
	IsDev bool
}

func (w *rawErrorHandlerWriter) Status() int {
	return w.status
}

// WriteHeader overrides the [http.ResponseWriter].
func (w *rawErrorHandlerWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func wrapErrorHandlerWriter(w http.ResponseWriter) ErrorHandlerWriter {
	sw := &rawErrorHandlerWriter{
		ResponseWriter: w,
		status:         http.StatusOK, // default status for when Write() is called.
	}

	if f, ok := w.(http.Flusher); ok {
		return &flushableErrorHandlerWriter{
			rawErrorHandlerWriter: sw,
			Flusher:               f,
		}
	}

	return sw
}

// ErrorHandler creates middleware that provides centralized error handling
// for Cosmos applications. It captures HTTP status codes, logs server errors
// (5xx status codes), and provides different error handling strategies based
// on the environment and error type.
//
// The middleware supports:
//   - Status code capture and logging for server errors (5xx)
//   - Development mode with enhanced error details via ErrorHandlerDevHandler
//   - Standard http.Handler error types for custom error responses
//   - Fallback to basic error responses for unhandled error types
//
// Parameters:
//   - options: Configuration options including logger and development mode settings
//
// Returns a middleware function that handles errors from downstream handlers.
func ErrorHandler(options ErrorHandlerOptions) framework.Middleware {
	// Make sure we always have a valid logger, even if
	// this means just discarding the content itself.
	if options.Logger == nil {
		options.Logger = slog.New(slog.DiscardHandler)
	}

	return func(next framework.Handler) framework.Handler {
		return func(w http.ResponseWriter, r *http.Request) error {
			sw := wrapErrorHandlerWriter(w)
			var err error

			defer func() {
				if status := sw.Status(); status >= 500 && status < 600 {
					options.Logger.ErrorContext(
						r.Context(),
						"request failed",
						"method", r.Method,
						"url", r.URL.String(),
						"status", status,
						"err", err,
					)
				}
			}()

			if err = next(sw, r); err != nil {
				// If we are in development, we can check if
				// the handler implement a development handler.
				// Those usually add more context to the error.
				if h, ok := err.(ErrorHandlerDevHandler); options.IsDev && ok {
					h.ServeHTTPDev(sw, r)

					return nil
				}

				// We can check if the error can be directly
				// handled by using a [http.Handler], in the case
				// we'll simply handle it using ServeHTTP.
				if h, ok := err.(http.Handler); ok {
					h.ServeHTTP(sw, r)

					return nil
				}

				// In case the error does not implement the [http.Handler] it can
				// already be handled by a simple
				http.Error(sw, err.Error(), http.StatusInternalServerError)

				return nil
			}

			return nil
		}
	}
}
