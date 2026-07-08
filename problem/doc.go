// Package problem implements RFC 9457 Problem Details for HTTP APIs.
//
// The package centers on [Problem], an error-compatible value type that can be
// served directly as an HTTP response and safely derived via copy-on-write
// helpers such as [Problem.With] and [Problem.WithError].
//
// # Content negotiation
//
// Problem responses are negotiated using request Accept headers and can be
// emitted as application/problem+json, application/json, or text/plain.
//
// # Immutability
//
// Methods that add or remove metadata return modified copies rather than
// mutating shared values. This makes package-level problem templates safe to
// reuse across requests.
//
// Application code should define [Problem] values as package-level variables
// (for example, var ErrUserNotFound = problem.Problem{...}) and derive
// per-request instances with [Problem.WithError] and [Problem.With].
//
// Example
//
//	var ErrNotFound = problem.Problem{
//		Title:  "Resource Not Found",
//		Detail: "The requested resource does not exist",
//		Status: http.StatusNotFound,
//	}
//
//	ErrNotFound.With("resource_id", "123").ServeHTTP(w, r)
package problem
