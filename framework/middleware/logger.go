package middleware

import (
	"log/slog"
	"net/http"

	"github.com/studiolambda/cosmos/contract/request"
	"github.com/studiolambda/cosmos/framework"
)

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
			hooks := request.Hooks(r)
			status := http.StatusInternalServerError

			hooks.BeforeWriteHeader(func(w http.ResponseWriter, s int) {
				status = s
			})

			hooks.AfterResponse(func(err error) {
				if err != nil || (status >= 500 && status < 600) {
					logger.ErrorContext(
						r.Context(),
						"request failed",
						"method", r.Method,
						"url", r.URL.String(),
						"status", status,
						"err", err,
					)
				}
			})

			return next(w, r)
		}
	}
}
