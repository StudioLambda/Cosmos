package request

import "net/http"

// Param retrieves a path parameter value by name from the HTTP request.
// This uses Go's built-in PathValue method which extracts values from
// URL path patterns like "/users/{id}" where {id} is the parameter name.
//
// Parameters:
//   - r: The HTTP request containing the path parameters
//   - k: The name of the path parameter to retrieve
//
// Returns the parameter value as a string, or empty string if not found.
func Param(r *http.Request, k string) string {
	return r.PathValue(k)
}

// ParamOr retrieves a path parameter value by name, returning a default
// value if the parameter doesn't exist or is empty. This is useful for
// providing fallback values when path parameters are optional or when
// you want to handle missing parameters gracefully.
//
// Parameters:
//   - r: The HTTP request containing the path parameters
//   - k: The name of the path parameter to retrieve
//   - d: The default value to return if the parameter is not found or empty
//
// Returns the parameter value if found and non-empty, otherwise the default value.
func ParamOr(r *http.Request, k string, d string) string {
	if p := Param(r, k); p != "" {
		return p
	}

	return d
}
