package request

import (
	"fmt"
	"net/http"
	"strconv"
)

// Query retrieves a query parameter value by name from the
// HTTP request URL. This extracts values from the URL query
// string like "?name=value&other=test" where the parameter
// name matches the provided key.
//
// Parameters:
//   - r: The HTTP request containing the URL with query parameters
//   - name: The name of the query parameter to retrieve
//
// Returns the first value associated with the key, or empty string if not found.
func Query(r *http.Request, name string) string {
	return r.URL.Query().Get(name)
}

// HasQuery checks if a query parameter exists in the HTTP
// request URL, regardless of its value. This is useful for
// distinguishing between a parameter that doesn't exist and
// one that exists but has an empty value.
//
// Parameters:
//   - r: The HTTP request containing the URL with query parameters
//   - name: The name of the query parameter to check for
//
// Returns true if the parameter exists in the query string, false otherwise.
func HasQuery(r *http.Request, name string) bool {
	return r.URL.Query().Has(name)
}

// QueryOr retrieves a query parameter value by name,
// returning a default value if the parameter doesn't exist.
// Note that if the parameter exists but has an empty value,
// the empty value is returned, not the default. This is
// useful for providing fallback values for optional
// parameters.
//
// Parameters:
//   - r: The HTTP request containing the URL with query parameters
//   - name: The name of the query parameter to retrieve
//   - fallback: The default value to return if the parameter doesn't exist
//
// Returns the parameter value if it exists, otherwise the default value.
func QueryOr(r *http.Request, name string, fallback string) string {
	values := r.URL.Query()

	if !values.Has(name) {
		return fallback
	}

	return values.Get(name)
}

// QueryInt retrieves a query parameter by name and parses
// it as an integer. This prevents injection via malformed
// numeric query parameters by validating that the value is
// a well-formed integer.
//
// Parameters:
//   - r: The HTTP request containing the URL with query
//     parameters
//   - k: The name of the query parameter to parse
//
// Returns the parsed integer value and any parsing error.
// Returns an error if the parameter is missing or is not
// a valid integer string.
func QueryInt(r *http.Request, k string) (int, error) {
	raw := Query(r, k)

	if raw == "" {
		return 0, fmt.Errorf("query parameter %q is empty", k)
	}

	value, err := strconv.Atoi(raw)

	if err != nil {
		return 0, fmt.Errorf(
			"query parameter %q is not a valid integer: %w",
			k, err,
		)
	}

	return value, nil
}

// QueryIntOr retrieves a query parameter by name and parses
// it as an integer, returning the provided fallback value
// if the parameter is missing or cannot be parsed. This is
// useful when a numeric query parameter is optional or when
// a sensible default exists (e.g., pagination page numbers).
//
// Parameters:
//   - r: The HTTP request containing the URL with query
//     parameters
//   - k: The name of the query parameter to parse
//   - d: The fallback value to return on failure
//
// Returns the parsed integer if valid, otherwise the
// fallback value.
func QueryIntOr(r *http.Request, k string, d int) int {
	value, err := QueryInt(r, k)

	if err != nil {
		return d
	}

	return value
}
