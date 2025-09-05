package request

import "net/http"

func Cookie(r *http.Request, k string) *http.Cookie {
	if cookie, err := r.Cookie(k); err != nil {
		return cookie
	}

	return nil
}

func CookieValue(r *http.Request, k string) string {
	if c := Cookie(r, k); c != nil {
		return c.Value
	}

	return ""
}

func CookieValueOr(r *http.Request, k string, d string) string {
	if v := CookieValue(r, k); v != "" {
		return v
	}

	return d
}
