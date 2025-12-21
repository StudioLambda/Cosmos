package framework

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"

	"github.com/studiolambda/cosmos/framework/hook"
	"github.com/studiolambda/cosmos/problem"
)

// Handler defines the function signature for HTTP request handlers in Cosmos.
// Unlike the standard http.HandlerFunc, Cosmos handlers return an error to
// enable centralized error handling and cleaner error propagation throughout
// the application middleware chain.
//
// The error return value allows handlers to:
//   - Return structured errors that can be handled by error middleware
//   - Avoid having to manually write error responses in every handler
//   - Enable consistent error formatting across the application
//
// Parameters:
//   - w: The HTTP response writer for sending the response
//   - r: The HTTP request containing client data and context
//
// Returns an error if the request handling fails, nil on success.
type Handler func(w http.ResponseWriter, r *http.Request) error

// HTTPStatus is an interface that errors can implement to specify
// a custom HTTP status code when they are returned from a handler.
// This allows for more precise error handling and appropriate HTTP
// response codes based on the type of error that occurred.
//
// When a handler returns an error that implements HTTPStatus, the
// ServeHTTP method will use the returned status code instead of
// the default 500 Internal Server Error.
//
// Example usage:
//
//	type NotFoundError struct {
//	    Resource string
//	}
//
//	func (e NotFoundError) Error() string {
//	    return fmt.Sprintf("resource not found: %s", e.Resource)
//	}
//
//	func (e NotFoundError) HTTPStatus() int {
//	    return http.StatusNotFound
//	}
type HTTPStatus interface {
	HTTPStatus() int
}

func handleError(w http.ResponseWriter, r *http.Request, err error) {
	status := http.StatusInternalServerError

	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		status = 499 // A non-standard status code: 499 Client Closed Request
	}

	if s, ok := err.(HTTPStatus); ok {
		status = s.HTTPStatus()
	}

	// We can check if the error can be directly
	// handled by using a [http.Handler], in the case
	// we'll simply handle it using ServeHTTP.
	if h, ok := err.(http.Handler); ok {
		h.ServeHTTP(w, r)
		return
	}

	problem.NewProblem(err, status).ServeHTTP(w, r)
}

// ServeHTTP implements the http.Handler interface, allowing Cosmos handlers
// to be used with the standard HTTP server. It bridges the gap between
// Cosmos's error-returning handlers and Go's standard http.Handler interface.
//
// This method calls the handler and provides basic error handling as a fallback.
// If the handler returns an error, it attempts to send a 500 Internal Server Error
// response to the client. However, if the response has already been partially
// written (e.g., during streaming), the error response may not be deliverable.
func (handler Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	hooks := hook.NewManager()
	wrapped := hook.NewResponseWriter(w, hooks)
	ctx := context.WithValue(r.Context(), hook.Key, hooks)
	err := handler(wrapped, r.WithContext(ctx))

	if err != nil {
		handleError(w, r, err)
	}

	if !wrapped.WriteHeaderCalled() {
		wrapped.WriteHeader(http.StatusNoContent)
	}

	for _, callback := range hooks.AfterResponseFuncs() {
		callback(err)
	}
}

// Record executes the handler with the given request and returns the resulting HTTP response.
// It uses httptest.NewRecorder() to capture the response that would be written to a client,
// making it useful for testing HTTP handlers without starting a server.
func (handler Handler) Record(r *http.Request) *http.Response {
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, r)

	return rec.Result()
}
