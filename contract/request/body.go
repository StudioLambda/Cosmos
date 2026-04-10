package request

import (
	"encoding/json"
	"encoding/xml"
	"io"
	"net/http"
)

// Bytes reads the entire request body and returns it as a byte slice.
// The request body is consumed after this call and cannot be read again.
func Bytes(r *http.Request) ([]byte, error) {
	return io.ReadAll(r.Body)
}

// String reads the request body and returns it as a string.
// It uses [Bytes] internally. The request body is consumed
// after this call and cannot be read again.
func String(r *http.Request) (string, error) {
	b, err := Bytes(r)

	if err != nil {
		return "", err
	}

	return string(b), nil
}

// JSON decodes JSON data from the request body into a value of type T.
// It uses a streaming decoder for memory efficiency. The type parameter
// T should match the expected JSON structure.
func JSON[T any](r *http.Request) (value T, err error) {
	if err := json.NewDecoder(r.Body).Decode(&value); err != nil {
		return value, err
	}

	return value, nil
}

// XML decodes XML data from the request body into a value of type T.
// It uses a streaming decoder for memory efficiency. The type parameter
// T should have appropriate xml struct tags or implement [xml.Unmarshaler].
func XML[T any](r *http.Request) (value T, err error) {
	if err := xml.NewDecoder(r.Body).Decode(&value); err != nil {
		return value, err
	}

	return value, nil
}
