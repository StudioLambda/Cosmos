package request

import (
	"encoding/json"
	"encoding/xml"
	"io"
	"net/http"
)

func Bytes(r *http.Request) ([]byte, error) {
	b, err := io.ReadAll(r.Body)

	if err != nil {
		return nil, err
	}

	return b, nil
}

func String(r *http.Request) (string, error) {
	b, err := Bytes(r)

	if err != nil {
		return "", err
	}

	return string(b), nil
}

// JSON decodes the given value from the request body.
func JSON[T any](r *http.Request) (value T, err error) {
	if err := json.NewDecoder(r.Body).Decode(&value); err != nil {
		return value, err
	}

	return value, nil
}

func XML[T any](r *http.Request) (value T, err error) {
	if err := xml.NewDecoder(r.Body).Decode(&value); err != nil {
		return value, err
	}

	return value, nil
}
