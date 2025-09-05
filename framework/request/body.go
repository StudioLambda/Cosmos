package request

import (
	"encoding/json"
	"encoding/xml"
	"io"
	"net/http"
)

// Bytes reads the entire request body and returns it as a byte slice.
// This function consumes the request body, so it can only be called once per request.
// The request body is automatically closed after reading.
//
// Parameters:
//   - r: The HTTP request containing the body to read
//
// Returns the raw byte data from the request body, or an error if reading fails.
func Bytes(r *http.Request) ([]byte, error) {
	b, err := io.ReadAll(r.Body)

	if err != nil {
		return nil, err
	}

	return b, nil
}

// String reads the request body and returns it as a string.
// This is a convenience function that uses Bytes internally and converts
// the result to a string. The request body is consumed and cannot be read again.
//
// Parameters:
//   - r: The HTTP request containing the body to read
//
// Returns the request body as a string, or an error if reading fails.
func String(r *http.Request) (string, error) {
	b, err := Bytes(r)

	if err != nil {
		return "", err
	}

	return string(b), nil
}

// JSON decodes JSON data from the request body into a value of type T.
// It uses streaming JSON decoding for memory efficiency and automatically
// closes the request body after reading. The function uses generics to
// provide type safety for the decoded result.
//
// The target type T should be a pointer to the desired struct or type
// that matches the expected JSON structure.
//
// Parameters:
//   - r: The HTTP request containing JSON data in the body
//
// Returns the decoded value of type T, or an error if decoding fails.
func JSON[T any](r *http.Request) (value T, err error) {
	if err := json.NewDecoder(r.Body).Decode(&value); err != nil {
		return value, err
	}

	return value, nil
}

// XML decodes XML data from the request body into a value of type T.
// It uses streaming XML decoding for memory efficiency and automatically
// closes the request body after reading. The function uses generics to
// provide type safety for the decoded result.
//
// The target type T should have appropriate xml struct tags for proper
// unmarshaling, or implement xml.Unmarshaler interface for custom deserialization.
//
// Parameters:
//   - r: The HTTP request containing XML data in the body
//
// Returns the decoded value of type T, or an error if decoding fails.
func XML[T any](r *http.Request) (value T, err error) {
	if err := xml.NewDecoder(r.Body).Decode(&value); err != nil {
		return value, err
	}

	return value, nil
}
