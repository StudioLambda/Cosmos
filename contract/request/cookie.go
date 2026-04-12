package request

import "net/http"

// Cookie retrieves the named cookie from the HTTP request using
// [http.Request.Cookie]. It returns nil if the cookie does not exist.
func Cookie(r *http.Request, name string) *http.Cookie {
	if cookie, err := r.Cookie(name); err == nil {
		return cookie
	}

	return nil
}

// CookieValue retrieves the value of the named cookie from the HTTP
// request. It returns an empty string if the cookie does not exist.
func CookieValue(r *http.Request, name string) string {
	if cookie := Cookie(r, name); cookie != nil {
		return cookie.Value
	}

	return ""
}

// CookieValueOr retrieves the value of the named cookie, falling back
// to the provided default value if the cookie is missing or empty.
func CookieValueOr(r *http.Request, name string, fallback string) string {
	if value := CookieValue(r, name); value != "" {
		return value
	}

	return fallback
}
