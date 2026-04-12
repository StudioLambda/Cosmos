package request

import (
	"fmt"
	"net/http"
	"strconv"
)

// Query retrieves the first value of the named query parameter
// from the request URL. It returns an empty string if the parameter
// is not present.
func Query(r *http.Request, name string) string {
	return r.URL.Query().Get(name)
}

// HasQuery reports whether the named query parameter exists in the
// request URL, regardless of its value.
func HasQuery(r *http.Request, name string) bool {
	return r.URL.Query().Has(name)
}

// QueryOr retrieves the named query parameter, falling back to the
// provided default value if the parameter does not exist. If the
// parameter exists but has an empty value, the empty value is returned.
func QueryOr(r *http.Request, name string, fallback string) string {
	values := r.URL.Query()

	if !values.Has(name) {
		return fallback
	}

	return values.Get(name)
}

// QueryInt retrieves the named query parameter and parses it as an
// integer. It returns an error if the parameter is missing or is not
// a valid integer string.
func QueryInt(r *http.Request, name string) (int, error) {
	raw := Query(r, name)

	if raw == "" {
		return 0, fmt.Errorf("query parameter %q is empty", name)
	}

	value, err := strconv.Atoi(raw)

	if err != nil {
		return 0, fmt.Errorf(
			"query parameter %q is not a valid integer: %w",
			name, err,
		)
	}

	return value, nil
}

// QueryIntOr retrieves the named query parameter and parses it as an
// integer, falling back to the provided default value if the parameter
// is missing or cannot be parsed.
func QueryIntOr(r *http.Request, name string, fallback int) int {
	value, err := QueryInt(r, name)

	if err != nil {
		return fallback
	}

	return value
}
