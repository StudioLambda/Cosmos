package request

import "net/http"

// Cookie retrieves a cookie by name from the HTTP request.
// Returns the cookie if found, or nil if the cookie doesn't exist
// or if there's an error retrieving it.
//
// Parameters:
//   - r: The HTTP request to search for the cookie
//   - name: The name of the cookie to retrieve
//
// Returns the cookie object or nil if not found.
func Cookie(r *http.Request, name string) *http.Cookie {
	if cookie, err := r.Cookie(name); err == nil {
		return cookie
	}

	return nil
}

// CookieValue retrieves the value of a cookie by name from the HTTP request.
// This is a convenience function that extracts just the value from the cookie.
// Returns an empty string if the cookie doesn't exist.
//
// Parameters:
//   - r: The HTTP request to search for the cookie
//   - name: The name of the cookie whose value to retrieve
//
// Returns the cookie value as a string, or empty string if not found.
func CookieValue(r *http.Request, name string) string {
	if cookie := Cookie(r, name); cookie != nil {
		return cookie.Value
	}

	return ""
}

// CookieValueOr retrieves the value of a cookie by name, returning a default
// value if the cookie doesn't exist or has an empty value. This is useful
// for providing fallback values when cookies are optional.
//
// Parameters:
//   - r: The HTTP request to search for the cookie
//   - name: The name of the cookie whose value to retrieve
//   - fallback: The default value to return if the cookie is not found or empty
//
// Returns the cookie value if found and non-empty, otherwise the default value.
func CookieValueOr(r *http.Request, name string, fallback string) string {
	if value := CookieValue(r, name); value != "" {
		return value
	}

	return fallback
}
