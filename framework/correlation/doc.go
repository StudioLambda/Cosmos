// Package correlation documents request correlation ID support.
//
// Use middleware.Correlation to accept a valid W3C traceparent trace ID or
// configured correlation header, generate a safe fallback, store the result in
// request context, and write it to the response header. Retrieve the ID with
// request.CorrelationID.
package correlation
