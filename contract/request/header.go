package request

import "net/http"

// HasHeader reports whether the given request contains a non-empty
// value for the specified header key.
func HasHeader(r *http.Request, key string) bool {
	return Header(r, key) != ""
}

// Header returns the first value associated with the given header
// key from the request. It returns an empty string if the header
// is not present.
func Header(r *http.Request, key string) string {
	return r.Header.Get(key)
}

// HeaderOr returns the first value for the given header key, falling
// back to the provided default value if the header is missing or empty.
func HeaderOr(r *http.Request, key string, def string) string {
	if h := Header(r, key); h != "" {
		return h
	}

	return def
}

// HeaderValues returns all values associated with the given header
// key from the request. It returns nil if the header is not present.
func HeaderValues(r *http.Request, key string) []string {
	return r.Header.Values(key)
}
