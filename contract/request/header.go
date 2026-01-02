package request

import "net/http"

func HasHeader(r *http.Request, key string) bool {
	return Header(r, key) != ""
}

func Header(r *http.Request, key string) string {
	return r.Header.Get(key)
}

func HeaderOr(r *http.Request, key string, def string) string {
	if h := Header(r, key); h != "" {
		return h
	}

	return def
}

func HeaderValues(r *http.Request, key string) []string {
	return r.Header.Values(key)
}
