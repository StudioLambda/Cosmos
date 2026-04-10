package request

import "net/http"

// Query retrieves a query parameter value by name from the HTTP request URL.
// This extracts values from the URL query string like "?name=value&other=test"
// where the parameter name matches the provided key.
//
// Parameters:
//   - r: The HTTP request containing the URL with query parameters
//   - name: The name of the query parameter to retrieve
//
// Returns the first value associated with the key, or empty string if not found.
func Query(r *http.Request, name string) string {
	return r.URL.Query().Get(name)
}

// HasQuery checks if a query parameter exists in the HTTP request URL,
// regardless of its value. This is useful for distinguishing between
// a parameter that doesn't exist and one that exists but has an empty value.
//
// Parameters:
//   - r: The HTTP request containing the URL with query parameters
//   - name: The name of the query parameter to check for
//
// Returns true if the parameter exists in the query string, false otherwise.
func HasQuery(r *http.Request, name string) bool {
	return r.URL.Query().Has(name)
}

// QueryOr retrieves a query parameter value by name, returning a default
// value if the parameter doesn't exist. Note that if the parameter exists
// but has an empty value, the empty value is returned, not the default.
// This is useful for providing fallback values for optional parameters.
//
// Parameters:
//   - r: The HTTP request containing the URL with query parameters
//   - name: The name of the query parameter to retrieve
//   - fallback: The default value to return if the parameter doesn't exist
//
// Returns the parameter value if it exists, otherwise the default value.
func QueryOr(r *http.Request, name string, fallback string) string {
	if HasQuery(r, name) {
		return Query(r, name)
	}

	return fallback
}
