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
//
// The wrapper intercepts calls to WriteHeader to record the status code while
// preserving all other ResponseWriter functionality.
type rawErrorHandlerWriter struct {
	http.ResponseWriter
	status int // The HTTP status code written to the response
}

// flushableErrorHandlerWriter extends rawErrorHandlerWriter to support the
// http.Flusher interface when the underlying ResponseWriter supports flushing.
// This enables streaming responses while still capturing status codes.
//
// This wrapper is automatically used when the original ResponseWriter implements
// http.Flusher, ensuring compatibility with streaming HTTP responses.
type flushableErrorHandlerWriter struct {
	*rawErrorHandlerWriter
	http.Flusher
}

// ErrorHandlerWriter extends http.ResponseWriter with the ability to retrieve
// the HTTP status code that was written to the response. This interface is
// used by the error handling middleware to make status-based decisions.
type ErrorHandlerWriter interface {
	http.ResponseWriter
	framework.HTTPStatus
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

// HTTPStatus returns the HTTP status code that was written to the response.
// If no status code has been explicitly set, returns http.StatusOK (200).
func (w *rawErrorHandlerWriter) HTTPStatus() int {
	return w.status
}

// WriteHeader captures the HTTP status code and forwards it to the underlying
// ResponseWriter. This override allows the middleware to track what status
// code was sent in the response for logging and error handling purposes.
//
// Parameters:
//   - status: The HTTP status code to write to the response
func (w *rawErrorHandlerWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

// wrapErrorHandlerWriter creates an ErrorHandlerWriter wrapper around the provided
// http.ResponseWriter. The wrapper captures HTTP status codes while preserving
// all original functionality of the underlying ResponseWriter.
//
// If the original ResponseWriter implements http.Flusher (for streaming responses),
// the wrapper will also implement http.Flusher to maintain compatibility.
//
// Parameters:
//   - w: The http.ResponseWriter to wrap
//
// Returns an ErrorHandlerWriter that can track HTTP status codes.
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

// Logger creates middleware that provides comprehensive request logging for Cosmos
// applications. It captures HTTP status codes, logs requests that result in errors
// or server errors (5xx status codes), and provides structured logging with
// contextual information about each request.
//
// The middleware logs the following information for failed requests:
//   - HTTP method (GET, POST, etc.)
//   - Full request URL
//   - HTTP status code returned
//   - Any error returned by the handler
//
// Logging is triggered when:
//   - A handler returns an error (regardless of status code)
//   - The response has a 5xx server error status code
//
// The middleware uses structured logging with slog for consistent log formatting
// and includes request context for distributed tracing compatibility.
//
// Parameters:
//   - logger: The slog.Logger instance to use for logging. If nil, a discard
//     logger is used to prevent panics while maintaining functionality.
//
// Returns a middleware function that wraps handlers with request logging.
//
// Example usage:
//
//	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
//	app.Use(middleware.Logger(logger))
func Logger(logger *slog.Logger) framework.Middleware {
	// Make sure we always have a valid logger, even if
	// this means just discarding the content itself.
	if logger == nil {
		logger = slog.New(slog.DiscardHandler)
	}

	return func(next framework.Handler) framework.Handler {
		return func(w http.ResponseWriter, r *http.Request) error {
			sw := wrapErrorHandlerWriter(w)
			var err error

			defer func() {
				if status := sw.HTTPStatus(); err != nil || (status >= 500 && status < 600) {
					logger.ErrorContext(
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
				return err
			}

			return nil
		}
	}
}
