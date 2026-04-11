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
	body, err := Bytes(r)

	if err != nil {
		return "", err
	}

	return string(body), nil
}

// LimitedString reads the request body up to maxSize bytes and
// returns it as a string. If the body exceeds maxSize, the result
// is truncated. Pass -1 to use [DefaultMaxBodySize].
func LimitedString(r *http.Request, maxSize int64) (string, error) {
	body, err := LimitedBytes(r, maxSize)

	if err != nil {
		return "", err
	}

	return string(body), nil
}

// JSON decodes JSON data from the request body into a value of
// type T. It uses a streaming decoder for memory efficiency. The
// type parameter T should match the expected JSON structure.
//
// Unknown fields in the JSON input are silently ignored. Use
// [StrictJSON] if unknown fields should cause an error.
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

// StrictJSON decodes JSON data from the request body into a value
// of type T, rejecting any fields not present in T's definition.
// This is useful for APIs that require exact schema compliance
// and want to surface typos or unsupported fields to callers.
//
// For a lenient variant that ignores unknown fields, use [JSON].
//
// WARNING: This function decodes without any body size limit.
// Prefer [StrictLimitedJSON] or apply [http.MaxBytesReader] in a
// middleware to prevent memory exhaustion from oversized requests.
func StrictJSON[T any](r *http.Request) (value T, err error) {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&value); err != nil {
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

// StrictLimitedJSON decodes JSON data from the request body into
// a value of type T, reading at most maxSize bytes and rejecting
// unknown fields. Pass -1 to use [DefaultMaxBodySize].
func StrictLimitedJSON[T any](r *http.Request, maxSize int64) (value T, err error) {
	if maxSize < 0 {
		maxSize = DefaultMaxBodySize
	}

	limited := io.LimitReader(r.Body, maxSize+1)
	decoder := json.NewDecoder(limited)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&value); err != nil {
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
