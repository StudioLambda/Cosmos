package request

import "net/http"

func String(r *http.Request) (string, error) {
	b, err := Bytes(r)

	if err != nil {
		return "", err
	}

	return string(b), nil
}
