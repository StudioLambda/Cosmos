// Package problem implements RFC 9457 Problem Details for HTTP APIs.
//
// The package centers on [Details], an error-compatible value type that can be
// served directly as an HTTP response and safely derived via copy-on-write
// helpers such as [Details.With] and [Details.WithError].
//
// # Content negotiation
//
// Problem Details responses are negotiated using request Accept headers and can be
// emitted as application/problem+json, application/json, or text/plain.
//
// # Immutability
//
// Methods that add or remove metadata return modified copies rather than
// mutating shared values. This makes package-level problem templates safe to
// reuse across requests.
//
// Application code should define [Details] values as package-level variables
// (for example, var ErrUserNotFound = problem.Details{...}) and derive
// per-request instances with [Details.WithError] and [Details.With].
//
// # Request extensions
//
// Middleware can attach safe client-facing extension members with
// [WithContextValues]. [Details.ServeHTTP] includes those values in the
// response without modifying the original Details value.
//
// Example
//
//	var ErrNotFound = problem.Details{
//		Title:  "Resource Not Found",
//		Detail: "The requested resource does not exist",
//		Status: http.StatusNotFound,
//	}
//
//	ErrNotFound.With("resource_id", "123").ServeHTTP(w, r)
package problem
