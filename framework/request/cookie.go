package request

import "net/http"

func Cookie(r *http.Request, k string) *http.Cookie {
	if cookie, err := r.Cookie(k); err != nil {
		return cookie
	}

	return nil
}
