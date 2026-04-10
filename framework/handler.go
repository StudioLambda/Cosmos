package framework

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"

	"github.com/studiolambda/cosmos/contract"
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

	var statusErr HTTPStatus

	if errors.As(err, &statusErr) {
		status = statusErr.HTTPStatus()
	}

	// When the error itself implements http.Handler, delegate
	// rendering entirely to it. This allows error types like
	// problem.Problem to control their own HTTP response format.
	var handlerErr http.Handler

	if errors.As(err, &handlerErr) {
		handlerErr.ServeHTTP(w, r)
		return
	}

	problem.NewProblem(err, status).ServeHTTP(w, r)
}

// ServeHTTP implements the http.Handler interface, bridging Cosmos's
// error-returning handlers with Go's standard handler contract.
//
// It wraps the response writer with lifecycle hooks, calls the handler,
// and delegates error rendering to handleError when the handler fails.
// Errors implementing [HTTPStatus] get their custom status code, and
// errors implementing [http.Handler] render themselves directly.
// If no status code has been written after the handler returns, a
// 204 No Content is sent as the default. AfterResponse hooks run
// last, regardless of success or failure.
func (handler Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	hooks := NewHooks()
	wrapped := NewResponseWriter(w, hooks)
	ctx := context.WithValue(r.Context(), contract.HooksKey, hooks)
	err := handler(wrapped, r.WithContext(ctx))

	if err != nil {
		handleError(wrapped, r, err)
	}

	if !wrapped.IsWriteHeaderCalled() {
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
