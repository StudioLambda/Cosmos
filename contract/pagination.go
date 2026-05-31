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

// Paginate creates a new [Page] from the given items, total count,
// current page number, and items per page. It computes the last
// page automatically. The current page is clamped to [1, LastPage].
//
//	page, perPage := request.Pagination(r)
//
//	var users []User
//	db.Select(ctx, "SELECT * FROM users LIMIT $1 OFFSET $2", &users, perPage, (page-1)*perPage)
//
//	var total int64
//	db.Find(ctx, "SELECT COUNT(*) FROM users", &total)
//
//	result := contract.Paginate(users, total, page, perPage)
func Paginate[T any](items []T, total int64, page, perPage int) Page[T] {
	perPage = max(perPage, 1)
	lastPage := max(int((total+int64(perPage)-1)/int64(perPage)), 1)
	page = min(max(page, 1), lastPage)

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

// CursorPaginate creates a new [Cursor] from the given items. The encode
// function determines how each item is transformed into an opaque
// cursor string. When hasNext is true, the last item is encoded to
// produce the next cursor. When hasPrev is true, the first item is
// encoded to produce the previous cursor.
//
//	cursor, perPage := request.CursorPagination(r)
//
//	var startID int64
//	if cursor != "" {
//		startID, _ = contract.CursorDecode[int64](cursor)
//	}
//
//	var items []FeedItem
//	db.Select(ctx, "SELECT * FROM feed WHERE id > $1 ORDER BY id LIMIT $2", &items, startID, perPage+1)
//
//	hasNext := len(items) > perPage
//	if hasNext {
//		items = items[:perPage]
//	}
//
//	result, err := contract.CursorPaginate(items, perPage, hasNext, cursor != "", func(item FeedItem) (string, error) {
//		return contract.CursorEncode(item.ID)
//	})
func CursorPaginate[T any](items []T, perPage int, hasNext, hasPrev bool, encode func(T) (string, error)) (Cursor[T], error) {
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
		encoded, err := encode(items[len(items)-1])

		if err != nil {
			return result, errors.Join(ErrCursorEncode, err)
		}

		result.NextCursor = encoded
	}

	if hasPrev {
		encoded, err := encode(items[0])

		if err != nil {
			return result, errors.Join(ErrCursorEncode, err)
		}

		result.PrevCursor = encoded
	}

	return result, nil
}

// CursorEncode encodes a value into an opaque cursor string using
// JSON serialization and base64url encoding. Use this as the encoding
// helper inside the encode function passed to [CursorPaginate].
//
//	encoded, err := contract.CursorEncode(user.ID)
func CursorEncode[V any](value V) (string, error) {
	data, err := json.Marshal(value)

	if err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(data), nil
}

// CursorDecode decodes an opaque cursor string back into a typed
// value. It reverses the encoding performed by [CursorEncode].
//
//	id, err := contract.CursorDecode[int64](cursorString)
func CursorDecode[V any](cursor string) (V, error) {
	var value V

	data, err := base64.RawURLEncoding.DecodeString(cursor)

	if err != nil {
		var zero V
		return zero, errors.Join(ErrCursorDecode, err)
	}

	if err := json.Unmarshal(data, &value); err != nil {
		var zero V
		return zero, errors.Join(ErrCursorDecode, err)
	}

	return value, nil
}
