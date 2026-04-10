package request

import (
	"fmt"
	"net/http"
	"strconv"
)

// Param retrieves a path parameter value by name from the
// HTTP request. This uses Go's built-in PathValue method
// which extracts values from URL path patterns like
// "/users/{id}" where {id} is the parameter name.
//
// Parameters:
//   - r: The HTTP request containing the path parameters
//   - k: The name of the path parameter to retrieve
//
// Returns the parameter value as a string, or empty string
// if not found.
func Param(r *http.Request, k string) string {
	return r.PathValue(k)
}

// ParamOr retrieves a path parameter value by name,
// returning a default value if the parameter doesn't exist
// or is empty. This is useful for providing fallback values
// when path parameters are optional or when you want to
// handle missing parameters gracefully.
//
// Parameters:
//   - r: The HTTP request containing the path parameters
//   - k: The name of the path parameter to retrieve
//   - d: The default value to return if the parameter is
//     not found or empty
//
// Returns the parameter value if found and non-empty,
// otherwise the default value.
func ParamOr(r *http.Request, k string, d string) string {
	if p := Param(r, k); p != "" {
		return p
	}

	return d
}

// ParamInt retrieves a path parameter by name and parses
// it as an integer. This prevents injection via malformed
// numeric path parameters by validating that the value is
// a well-formed integer.
//
// Parameters:
//   - r: The HTTP request containing the path parameters
//   - k: The name of the path parameter to parse
//
// Returns the parsed integer value and any parsing error.
// Returns an error if the parameter is empty or is not
// a valid integer string.
func ParamInt(r *http.Request, k string) (int, error) {
	raw := Param(r, k)

	if raw == "" {
		return 0, fmt.Errorf("path parameter %q is empty", k)
	}

	value, err := strconv.Atoi(raw)

	if err != nil {
		return 0, fmt.Errorf(
			"path parameter %q is not a valid integer: %w",
			k, err,
		)
	}

	return value, nil
}

// ParamIntOr retrieves a path parameter by name and parses
// it as an integer, returning the provided fallback value
// if the parameter is empty or cannot be parsed. This is
// useful when a numeric path parameter is optional or when
// a sensible default exists.
//
// Parameters:
//   - r: The HTTP request containing the path parameters
//   - k: The name of the path parameter to parse
//   - d: The fallback value to return on failure
//
// Returns the parsed integer if valid, otherwise the
// fallback value.
func ParamIntOr(r *http.Request, k string, d int) int {
	value, err := ParamInt(r, k)

	if err != nil {
		return d
	}

	return value
}
