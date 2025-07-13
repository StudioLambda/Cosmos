package request

import "net/http"

func Query(r *http.Request, k string) string {
	return r.URL.Query().Get(k)
}

func HasQuery(r *http.Request, k string) bool {
	return r.URL.Query().Has(k)
}

func QueryOr(r *http.Request, k string, d string) string {
	if HasQuery(r, k) {
		return Query(r, k)
	}

	return d
}
