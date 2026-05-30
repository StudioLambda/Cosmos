package contract

import (
	"encoding/base64"
	"encoding/json"
	"errors"
)

// ErrCursorEncode is returned when a cursor value fails to encode.
var ErrCursorEncode = errors.New("failed to encode cursor")

// ErrCursorDecode is returned when a cursor string fails to decode.
var ErrCursorDecode = errors.New("failed to decode cursor")

// CursorEncoder defines the encoding and decoding of cursor values
// into opaque cursor strings. Implementations control how cursor
// values are serialized for transport and deserialized on receipt.
type CursorEncoder interface {
	// Encode converts a cursor value into an opaque cursor string.
	Encode(value any) (string, error)

	// Decode converts an opaque cursor string back into a cursor value.
	Decode(cursor string) (any, error)
}

// Page represents an offset-based paginated result set.
type Page[T any] struct {
	Items       []T   `json:"items"`
	Total       int64 `json:"total"`
	PerPage     int   `json:"per_page"`
	CurrentPage int   `json:"current_page"`
	LastPage    int   `json:"last_page"`
}

// Cursor represents a cursor-based paginated result set.
type Cursor[T any] struct {
	Items      []T    `json:"items"`
	PerPage    int    `json:"per_page"`
	NextCursor string `json:"next_cursor,omitempty"`
	PrevCursor string `json:"prev_cursor,omitempty"`
}

// NewPage creates a new [Page] from the given items, total count,
// current page number, and items per page. It computes the last
// page automatically. The current page is clamped to [1, LastPage].
func NewPage[T any](items []T, total int64, page, perPage int) Page[T] {
	if perPage < 1 {
		perPage = 1
	}

	lastPage := int((total + int64(perPage) - 1) / int64(perPage))

	if lastPage < 1 {
		lastPage = 1
	}

	if page < 1 {
		page = 1
	}

	if page > lastPage {
		page = lastPage
	}

	if items == nil {
		items = []T{}
	}

	return Page[T]{
		Items:       items,
		Total:       total,
		PerPage:     perPage,
		CurrentPage: page,
		LastPage:    lastPage,
	}
}

// NewCursor creates a new [Cursor] from the given items using a
// built-in base64-JSON encoding for cursor values. The extract
// function determines which value from each item becomes the cursor.
// When hasNext is true, the last item's extracted value becomes the
// next cursor. When hasPrev is true, the first item's extracted
// value becomes the previous cursor.
func NewCursor[T any](items []T, perPage int, hasNext, hasPrev bool, extract func(T) any) (Cursor[T], error) {
	if items == nil {
		items = []T{}
	}

	result := Cursor[T]{
		Items:   items,
		PerPage: perPage,
	}

	if len(items) == 0 {
		return result, nil
	}

	if hasNext {
		encoded, err := base64JSONEncode(extract(items[len(items)-1]))

		if err != nil {
			return result, errors.Join(ErrCursorEncode, err)
		}

		result.NextCursor = encoded
	}

	if hasPrev {
		encoded, err := base64JSONEncode(extract(items[0]))

		if err != nil {
			return result, errors.Join(ErrCursorEncode, err)
		}

		result.PrevCursor = encoded
	}

	return result, nil
}

// NewCursorWith creates a new [Cursor] from the given items using
// a custom [CursorEncoder]. When hasNext is true, the last item is
// passed to the encoder to produce the next cursor. When hasPrev is
// true, the first item is passed to produce the previous cursor.
func NewCursorWith[T any](items []T, perPage int, hasNext, hasPrev bool, encoder CursorEncoder) (Cursor[T], error) {
	if items == nil {
		items = []T{}
	}

	result := Cursor[T]{
		Items:   items,
		PerPage: perPage,
	}

	if len(items) == 0 {
		return result, nil
	}

	if hasNext {
		encoded, err := encoder.Encode(items[len(items)-1])

		if err != nil {
			return result, errors.Join(ErrCursorEncode, err)
		}

		result.NextCursor = encoded
	}

	if hasPrev {
		encoded, err := encoder.Encode(items[0])

		if err != nil {
			return result, errors.Join(ErrCursorEncode, err)
		}

		result.PrevCursor = encoded
	}

	return result, nil
}

// CursorValue decodes a cursor string produced by [NewCursor] back
// into the original cursor value using the built-in base64-JSON encoding.
func CursorValue[T any](cursor string) (T, error) {
	var zero T

	raw, err := base64JSONDecode(cursor)

	if err != nil {
		return zero, errors.Join(ErrCursorDecode, err)
	}

	value, ok := raw.(T)

	if !ok {
		return zero, ErrCursorDecode
	}

	return value, nil
}

// CursorValueWith decodes a cursor string using a custom [CursorEncoder].
func CursorValueWith(cursor string, encoder CursorEncoder) (any, error) {
	value, err := encoder.Decode(cursor)

	if err != nil {
		return nil, errors.Join(ErrCursorDecode, err)
	}

	return value, nil
}

// base64JSONEncode encodes a value as JSON then base64url.
func base64JSONEncode(value any) (string, error) {
	data, err := json.Marshal(value)

	if err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(data), nil
}

// base64JSONDecode decodes a base64url string then unmarshals as JSON.
func base64JSONDecode(cursor string) (any, error) {
	data, err := base64.RawURLEncoding.DecodeString(cursor)

	if err != nil {
		return nil, err
	}

	var value any

	if err := json.Unmarshal(data, &value); err != nil {
		return nil, err
	}

	return value, nil
}
