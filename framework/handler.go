package framework

import "net/http"

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

// ServeHTTP implements the http.Handler interface, allowing Cosmos handlers
// to be used with the standard HTTP server. It bridges the gap between
// Cosmos's error-returning handlers and Go's standard http.Handler interface.
//
// This method calls the handler and provides basic error handling as a fallback.
// If the handler returns an error, it attempts to send a 500 Internal Server Error
// response to the client. However, if the response has already been partially
// written (e.g., during streaming), the error response may not be deliverable.
func (handler Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := handler(w, r); err != nil {
		status := http.StatusInternalServerError

		if s, ok := err.(HTTPStatus); ok {
			status = s.HTTPStatus()
		}

		http.Error(w, err.Error(), status)
	}
}
