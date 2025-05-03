package middleware

import (
	"log/slog"
	"net/http"

	"github.com/studiolambda/cosmos/nova"
)

// rawErrorHandlerWriter is a small type designed to access the
// response status code after it has been written.
type rawErrorHandlerWriter struct {
	http.ResponseWriter
	status int
}

// flushableErrorHandlerWriter is a small type that wraps the
// [errorHandlerWriter] struct so that it supports the flush
// operation.
type flushableErrorHandlerWriter struct {
	*rawErrorHandlerWriter
	http.Flusher
}

type ErrorHandlerWriter interface {
	http.ResponseWriter
	Status() int
}

type ErrorHandlerDevHandler interface {
	ServeHTTPDev(w http.ResponseWriter, r *http.Request)
}

type ErrorHandlerOptions struct {
	Logger *slog.Logger
	IsDev  bool
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

func ErrorHandler(options ErrorHandlerOptions) nova.Middleware {
	// Make sure we always have a valid logger, even if
	// this means just discarding the content itself.
	if options.Logger == nil {
		options.Logger = slog.New(slog.DiscardHandler)
	}

	return func(next nova.Handler) nova.Handler {
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
