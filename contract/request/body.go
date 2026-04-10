package request

import (
	"encoding/json"
	"encoding/xml"
	"io"
	"net/http"
)

// DefaultMaxBodySize is the default maximum request body size
// (10 MB) used by the size-limited body reading functions.
// This prevents denial-of-service attacks via excessively
// large request bodies that could exhaust server memory.
const DefaultMaxBodySize int64 = 10 << 20 // 10 MB

// Bytes reads the entire request body and returns it as a byte slice.
// The request body is consumed after this call and cannot be read again.
//
// WARNING: This function reads the body without any size limit.
// Prefer [LimitedBytes] or apply [http.MaxBytesReader] in a
// middleware to prevent memory exhaustion from oversized requests.
func Bytes(r *http.Request) ([]byte, error) {
	return io.ReadAll(r.Body)
}

// LimitedBytes reads the request body up to maxSize bytes and
// returns it as a byte slice. If the body exceeds maxSize, an
// error is returned. This prevents denial-of-service attacks via
// excessively large request bodies. Pass -1 to use
// [DefaultMaxBodySize].
func LimitedBytes(r *http.Request, maxSize int64) ([]byte, error) {
	if maxSize < 0 {
		maxSize = DefaultMaxBodySize
	}

	return io.ReadAll(io.LimitReader(r.Body, maxSize+1))
}

// String reads the request body and returns it as a string.
// It uses [Bytes] internally. The request body is consumed
// after this call and cannot be read again.
//
// WARNING: This function reads the body without any size limit.
// Prefer [LimitedString] or apply [http.MaxBytesReader] in a
// middleware to prevent memory exhaustion from oversized requests.
func String(r *http.Request) (string, error) {
	b, err := Bytes(r)

	if err != nil {
		return "", err
	}

	return string(b), nil
}

// LimitedString reads the request body up to maxSize bytes and
// returns it as a string. If the body exceeds maxSize, the result
// is truncated. Pass -1 to use [DefaultMaxBodySize].
func LimitedString(r *http.Request, maxSize int64) (string, error) {
	b, err := LimitedBytes(r, maxSize)

	if err != nil {
		return "", err
	}

	return string(b), nil
}

// JSON decodes JSON data from the request body into a value of type T.
// It uses a streaming decoder for memory efficiency. The type parameter
// T should match the expected JSON structure.
//
// WARNING: This function decodes without any body size limit.
// Prefer [LimitedJSON] or apply [http.MaxBytesReader] in a
// middleware to prevent memory exhaustion from oversized requests.
func JSON[T any](r *http.Request) (value T, err error) {
	if err := json.NewDecoder(r.Body).Decode(&value); err != nil {
		return value, err
	}

	return value, nil
}

// LimitedJSON decodes JSON data from the request body into a value
// of type T, reading at most maxSize bytes. This prevents
// denial-of-service attacks via oversized JSON payloads. Pass -1
// to use [DefaultMaxBodySize].
func LimitedJSON[T any](r *http.Request, maxSize int64) (value T, err error) {
	if maxSize < 0 {
		maxSize = DefaultMaxBodySize
	}

	limited := io.LimitReader(r.Body, maxSize+1)

	if err := json.NewDecoder(limited).Decode(&value); err != nil {
		return value, err
	}

	return value, nil
}

// XML decodes XML data from the request body into a value of type T.
// It uses a streaming decoder for memory efficiency. The type parameter
// T should have appropriate xml struct tags or implement [xml.Unmarshaler].
//
// WARNING: This function decodes without any body size limit.
// Prefer [LimitedXML] or apply [http.MaxBytesReader] in a
// middleware to prevent memory exhaustion from oversized requests.
func XML[T any](r *http.Request) (value T, err error) {
	if err := xml.NewDecoder(r.Body).Decode(&value); err != nil {
		return value, err
	}

	return value, nil
}

// LimitedXML decodes XML data from the request body into a value
// of type T, reading at most maxSize bytes. This prevents
// denial-of-service attacks via oversized XML payloads. Pass -1
// to use [DefaultMaxBodySize].
func LimitedXML[T any](r *http.Request, maxSize int64) (value T, err error) {
	if maxSize < 0 {
		maxSize = DefaultMaxBodySize
	}

	limited := io.LimitReader(r.Body, maxSize+1)

	if err := xml.NewDecoder(limited).Decode(&value); err != nil {
		return value, err
	}

	return value, nil
}
