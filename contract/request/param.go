package request

import (
	"fmt"
	"net/http"
	"strconv"
)

// Param retrieves the value of the named path parameter from the
// given HTTP request using [http.Request.PathValue].
func Param(r *http.Request, name string) string {
	return r.PathValue(name)
}

// ParamOr retrieves the named path parameter, falling back to
// the provided default value if the parameter is missing or empty.
func ParamOr(r *http.Request, name string, fallback string) string {
	if value := Param(r, name); value != "" {
		return value
	}

	return fallback
}

// ParamInt retrieves the named path parameter and parses it as an
// integer. It returns an error if the parameter is empty or is not
// a valid integer string.
func ParamInt(r *http.Request, name string) (int, error) {
	raw := Param(r, name)

	if raw == "" {
		return 0, fmt.Errorf("path parameter %q is empty", name)
	}

	value, err := strconv.Atoi(raw)

	if err != nil {
		return 0, fmt.Errorf(
			"path parameter %q is not a valid integer: %w",
			name, err,
		)
	}

	return value, nil
}

// ParamIntOr retrieves the named path parameter and parses it as an
// integer, falling back to the provided default value if the parameter
// is empty or cannot be parsed.
func ParamIntOr(r *http.Request, name string, fallback int) int {
	value, err := ParamInt(r, name)

	if err != nil {
		return fallback
	}

	return value
}
