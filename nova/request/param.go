package request

import "net/http"

func Param(r *http.Request, k string) string {
	return r.PathValue(k)
}

func ParamOr(r *http.Request, k string, d string) string {
	if p := Param(r, k); p != "" {
		return p
	}

	return d
}
