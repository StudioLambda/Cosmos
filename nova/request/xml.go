package request

import (
	"encoding/xml"
	"net/http"
)

func XML[T any](r *http.Request) (value T, err error) {
	if err := xml.NewDecoder(r.Body).Decode(value); err != nil {
		return value, err
	}

	return value, nil
}
