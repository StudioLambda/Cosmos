package request

import "net/http"

// HasHeader reports whether the given request contains a non-empty
// value for the specified header key.
//
// Example:
//
//	if request.HasHeader(r, "Authorization") {
//		// authenticated request path
//	}
func HasHeader(r *http.Request, key string) bool {
	return Header(r, key) != ""
}

// Header returns the first value associated with the given header
// key from the request. It returns an empty string if the header
// is not present.
//
// Example:
//
//	agent := request.Header(r, "User-Agent")
//	_ = agent
func Header(r *http.Request, key string) string {
	return r.Header.Get(key)
}

// HeaderOr returns the first value for the given header key, falling
// back to the provided default value if the header is missing or empty.
//
// Example:
//
//	contentType := request.HeaderOr(r, "Content-Type", "application/json")
//	_ = contentType
func HeaderOr(r *http.Request, key string, fallback string) string {
	if value := Header(r, key); value != "" {
		return value
	}

	return fallback
}

// HeaderValues returns all values associated with the given header
// key from the request. It returns nil if the header is not present.
//
// Example:
//
//	forwardedFor := request.HeaderValues(r, "X-Forwarded-For")
//	_ = forwardedFor
func HeaderValues(r *http.Request, key string) []string {
	return r.Header.Values(key)
}
