package contract_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/studiolambda/cosmos/contract"
)

func TestPaginateComputesLastPage(t *testing.T) {
	t.Parallel()

	page := contract.Paginate([]string{"a", "b"}, 10, 1, 5)

	require.Equal(t, 2, page.LastPage)
}

func TestPaginateComputesLastPageWithRemainder(t *testing.T) {
	t.Parallel()

	page := contract.Paginate([]string{"a", "b"}, 11, 1, 5)

	require.Equal(t, 3, page.LastPage)
}

func TestPaginateClampsPageBelowOne(t *testing.T) {
	t.Parallel()

	page := contract.Paginate([]string{"a"}, 10, 0, 5)

	require.Equal(t, 1, page.CurrentPage)
}

func TestPaginateClampsPageAboveLastPage(t *testing.T) {
	t.Parallel()

	page := contract.Paginate([]string{}, 10, 99, 5)

	require.Equal(t, 2, page.CurrentPage)
}

func TestPaginateClampsPerPageBelowOne(t *testing.T) {
	t.Parallel()

	page := contract.Paginate([]string{"a"}, 5, 1, 0)

	require.Equal(t, 1, page.PerPage)
}

func TestPaginateZeroTotalSetsLastPageOne(t *testing.T) {
	t.Parallel()

	page := contract.Paginate([]string{}, 0, 1, 10)

	require.Equal(t, 1, page.LastPage)
	require.Equal(t, 1, page.CurrentPage)
}

func TestPaginateNilItemsBecomesEmptySlice(t *testing.T) {
	t.Parallel()

	page := contract.Paginate[string](nil, 0, 1, 10)

	require.NotNil(t, page.Items)
	require.Empty(t, page.Items)
}

func TestPaginatePreservesItems(t *testing.T) {
	t.Parallel()

	items := []int{1, 2, 3}
	page := contract.Paginate(items, 100, 3, 10)

	require.Equal(t, items, page.Items)
	require.Equal(t, int64(100), page.Total)
	require.Equal(t, 3, page.CurrentPage)
	require.Equal(t, 10, page.PerPage)
}

func TestCursorPaginateEncodesNextCursor(t *testing.T) {
	t.Parallel()

	items := []int{1, 2, 3}
	cursor, err := contract.CursorPaginate(items, 3, true, false, func(item int) (string, error) {
		return contract.CursorEncode(item)
	})

	require.NoError(t, err)
	require.NotEmpty(t, cursor.NextCursor)
	require.Empty(t, cursor.PrevCursor)
}

func TestCursorPaginateEncodesPrevCursor(t *testing.T) {
	t.Parallel()

	items := []int{1, 2, 3}
	cursor, err := contract.CursorPaginate(items, 3, false, true, func(item int) (string, error) {
		return contract.CursorEncode(item)
	})

	require.NoError(t, err)
	require.Empty(t, cursor.NextCursor)
	require.NotEmpty(t, cursor.PrevCursor)
}

func TestCursorPaginateEncodesBothCursors(t *testing.T) {
	t.Parallel()

	items := []int{1, 2, 3}
	cursor, err := contract.CursorPaginate(items, 3, true, true, func(item int) (string, error) {
		return contract.CursorEncode(item)
	})

	require.NoError(t, err)
	require.NotEmpty(t, cursor.NextCursor)
	require.NotEmpty(t, cursor.PrevCursor)
}

func TestCursorPaginateEmptyItemsNoCursors(t *testing.T) {
	t.Parallel()

	cursor, err := contract.CursorPaginate([]int{}, 10, true, true, func(item int) (string, error) {
		return contract.CursorEncode(item)
	})

	require.NoError(t, err)
	require.Empty(t, cursor.NextCursor)
	require.Empty(t, cursor.PrevCursor)
}

func TestCursorPaginateNilItemsBecomesEmptySlice(t *testing.T) {
	t.Parallel()

	cursor, err := contract.CursorPaginate[int](nil, 10, false, false, func(item int) (string, error) {
		return contract.CursorEncode(item)
	})

	require.NoError(t, err)
	require.NotNil(t, cursor.Items)
	require.Empty(t, cursor.Items)
}

func TestCursorPaginatePreservesPerPage(t *testing.T) {
	t.Parallel()

	cursor, err := contract.CursorPaginate([]int{1}, 25, false, false, func(item int) (string, error) {
		return contract.CursorEncode(item)
	})

	require.NoError(t, err)
	require.Equal(t, 25, cursor.PerPage)
}

func TestCursorPaginateNextEncodeErrorReturnsErrCursorEncode(t *testing.T) {
	t.Parallel()

	_, err := contract.CursorPaginate([]int{1}, 10, true, false, func(item int) (string, error) {
		return "", errors.New("encode failed")
	})

	require.ErrorIs(t, err, contract.ErrCursorEncode)
}

func TestCursorPaginatePrevEncodeErrorReturnsErrCursorEncode(t *testing.T) {
	t.Parallel()

	_, err := contract.CursorPaginate([]int{1}, 10, false, true, func(item int) (string, error) {
		return "", errors.New("encode failed")
	})

	require.ErrorIs(t, err, contract.ErrCursorEncode)
}

func TestCursorPaginateCustomEncoder(t *testing.T) {
	t.Parallel()

	items := []int{1, 2, 3}
	cursor, err := contract.CursorPaginate(items, 3, true, false, func(item int) (string, error) {
		return "custom-cursor", nil
	})

	require.NoError(t, err)
	require.Equal(t, "custom-cursor", cursor.NextCursor)
}

func TestCursorEncodeDecoderRoundTrip(t *testing.T) {
	t.Parallel()

	encoded, err := contract.CursorEncode(42)

	require.NoError(t, err)

	value, err := contract.CursorDecode[int](encoded)

	require.NoError(t, err)
	require.Equal(t, 42, value)
}

func TestCursorEncodeStringRoundTrip(t *testing.T) {
	t.Parallel()

	encoded, err := contract.CursorEncode("hello")

	require.NoError(t, err)

	value, err := contract.CursorDecode[string](encoded)

	require.NoError(t, err)
	require.Equal(t, "hello", value)
}

func TestCursorEncodeStructRoundTrip(t *testing.T) {
	t.Parallel()

	type key struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}

	original := key{ID: 7, Name: "test"}
	encoded, err := contract.CursorEncode(original)

	require.NoError(t, err)

	value, err := contract.CursorDecode[key](encoded)

	require.NoError(t, err)
	require.Equal(t, original, value)
}

func TestCursorEncodeUnencodableReturnsError(t *testing.T) {
	t.Parallel()

	_, err := contract.CursorEncode(make(chan int))

	require.Error(t, err)
}

func TestCursorDecodeInvalidBase64ReturnsErrCursorDecode(t *testing.T) {
	t.Parallel()

	_, err := contract.CursorDecode[int]("not-valid-base64!!!")

	require.ErrorIs(t, err, contract.ErrCursorDecode)
}

func TestCursorDecodeInvalidJSONReturnsErrCursorDecode(t *testing.T) {
	t.Parallel()

	// Valid base64 but invalid JSON.
	_, err := contract.CursorDecode[int]("bm90LWpzb24")

	require.ErrorIs(t, err, contract.ErrCursorDecode)
}

func TestCursorPaginateWithMarshalCursorEndToEnd(t *testing.T) {
	t.Parallel()

	type User struct {
		ID   int
		Name string
	}

	users := []User{{ID: 10, Name: "Alice"}, {ID: 20, Name: "Bob"}}

	cursor, err := contract.CursorPaginate(users, 2, true, false, func(u User) (string, error) {
		return contract.CursorEncode(u.ID)
	})

	require.NoError(t, err)

	id, err := contract.CursorDecode[int](cursor.NextCursor)

	require.NoError(t, err)
	require.Equal(t, 20, id)
}
