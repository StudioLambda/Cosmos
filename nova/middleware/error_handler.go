package middleware

import (
	"log/slog"
	"net/http"

	"github.com/studiolambda/cosmos/nova"
)

// errorHandlerWriter is a small type designed to access the
// response status code after it has been written.
type errorHandlerWriter struct {
	http.ResponseWriter
	status int
}

type ErrorHandlerDevHandler interface {
	ServeHTTPDev(w http.ResponseWriter, r *http.Request)
}

type ErrorHandlerOptions struct {
	Logger *slog.Logger
	IsDev  bool
}

// WriteHeader overrides the [http.ResponseWriter].
func (w *errorHandlerWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func ErrorHandler(options ErrorHandlerOptions) nova.Middleware {
	// Make sure we always have a valid logger, even if
	// this means just discarding the content itself.
	if options.Logger == nil {
		options.Logger = slog.New(slog.DiscardHandler)
	}

	return func(next nova.Handler) nova.Handler {
		return func(w http.ResponseWriter, r *http.Request) error {
			sw := &errorHandlerWriter{
				ResponseWriter: w,
				status:         http.StatusOK, // default status for when Write() is called.
			}

			defer func() {
				if sw.status >= 500 && sw.status < 600 {
					options.Logger.ErrorContext(
						r.Context(),
						"request failed",
						"method", r.Method,
						"url", r.URL.String(),
						"status", sw.status,
					)
				}
			}()

			if err := next(sw, r); err != nil {
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
