// Package correlation provides slog integration for request correlation IDs.
//
// Use middleware.Correlation to accept a valid W3C traceparent trace ID or
// configured correlation header, generate a safe fallback, store the result in
// request context, and write it to the response header. [Handler] adds that
// context value to slog records.
package correlation
