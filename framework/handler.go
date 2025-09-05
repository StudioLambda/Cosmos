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

// ServeHTTP implements the http.Handler interface, allowing Cosmos handlers
// to be used with the standard HTTP server. It bridges the gap between
// Cosmos's error-returning handlers and Go's standard http.Handler interface.
//
// This method calls the handler and provides basic error handling as a fallback.
// If the handler returns an error, it attempts to send a 500 Internal Server Error
// response to the client. However, if the response has already been partially
// written (e.g., during streaming), the error response may not be deliverable.
//
// Important: if there's an unhandled error, a basic fallback error handler is used.
// For production applications, it's strongly recommended to use dedicated error
// handling middleware like middleware.ErrorHandler, which provides more sophisticated
// error handling, logging, and response formatting capabilities.
func (handler Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := handler(w, r); err != nil {
		// Fallback error handling - this will only work if the response
		// hasn't been written to yet. For streaming or partial responses
		// that fail mid-write, this may not be effective.
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
