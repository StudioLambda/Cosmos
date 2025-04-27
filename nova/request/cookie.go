package request

import "net/http"

func Cookie(r *http.Request, key string) *http.Cookie {
	if cookie, err := r.Cookie(key); err != nil {
		return cookie
	}

	return nil
}
