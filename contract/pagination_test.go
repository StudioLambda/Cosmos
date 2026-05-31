package contract_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/studiolambda/cosmos/contract"
)

func TestNewPageComputesLastPage(t *testing.T) {
	t.Parallel()

	page := contract.NewPage([]string{"a", "b"}, 10, 1, 5)

	require.Equal(t, 2, page.LastPage)
}

func TestNewPageComputesLastPageWithRemainder(t *testing.T) {
	t.Parallel()

	page := contract.NewPage([]string{"a", "b"}, 11, 1, 5)

	require.Equal(t, 3, page.LastPage)
}

func TestNewPageClampsPageBelowOne(t *testing.T) {
	t.Parallel()

	page := contract.NewPage([]string{"a"}, 10, 0, 5)

	require.Equal(t, 1, page.CurrentPage)
}

func TestNewPageClampsPageAboveLastPage(t *testing.T) {
	t.Parallel()

	page := contract.NewPage([]string{}, 10, 99, 5)

	require.Equal(t, 2, page.CurrentPage)
}

func TestNewPageClampsPerPageBelowOne(t *testing.T) {
	t.Parallel()

	page := contract.NewPage([]string{"a"}, 5, 1, 0)

	require.Equal(t, 1, page.PerPage)
}

func TestNewPageZeroTotalSetsLastPageOne(t *testing.T) {
	t.Parallel()

	page := contract.NewPage([]string{}, 0, 1, 10)

	require.Equal(t, 1, page.LastPage)
	require.Equal(t, 1, page.CurrentPage)
}

func TestNewPageNilItemsBecomesEmptySlice(t *testing.T) {
	t.Parallel()

	page := contract.NewPage[string](nil, 0, 1, 10)

	require.NotNil(t, page.Items)
	require.Empty(t, page.Items)
}

func TestNewPagePreservesItems(t *testing.T) {
	t.Parallel()

	items := []int{1, 2, 3}
	page := contract.NewPage(items, 100, 3, 10)

	require.Equal(t, items, page.Items)
	require.Equal(t, int64(100), page.Total)
	require.Equal(t, 3, page.CurrentPage)
	require.Equal(t, 10, page.PerPage)
}

func TestNewCursorEncodesNextCursor(t *testing.T) {
	t.Parallel()

	items := []int{1, 2, 3}
	cursor, err := contract.NewCursor(items, 3, true, false, func(item int) (string, error) {
		return contract.MarshalCursor(item)
	})

	require.NoError(t, err)
	require.NotEmpty(t, cursor.NextCursor)
	require.Empty(t, cursor.PrevCursor)
}

func TestNewCursorEncodesPrevCursor(t *testing.T) {
	t.Parallel()

	items := []int{1, 2, 3}
	cursor, err := contract.NewCursor(items, 3, false, true, func(item int) (string, error) {
		return contract.MarshalCursor(item)
	})

	require.NoError(t, err)
	require.Empty(t, cursor.NextCursor)
	require.NotEmpty(t, cursor.PrevCursor)
}

func TestNewCursorEncodesBothCursors(t *testing.T) {
	t.Parallel()

	items := []int{1, 2, 3}
	cursor, err := contract.NewCursor(items, 3, true, true, func(item int) (string, error) {
		return contract.MarshalCursor(item)
	})

	require.NoError(t, err)
	require.NotEmpty(t, cursor.NextCursor)
	require.NotEmpty(t, cursor.PrevCursor)
}

func TestNewCursorEmptyItemsNoCursors(t *testing.T) {
	t.Parallel()

	cursor, err := contract.NewCursor([]int{}, 10, true, true, func(item int) (string, error) {
		return contract.MarshalCursor(item)
	})

	require.NoError(t, err)
	require.Empty(t, cursor.NextCursor)
	require.Empty(t, cursor.PrevCursor)
}

func TestNewCursorNilItemsBecomesEmptySlice(t *testing.T) {
	t.Parallel()

	cursor, err := contract.NewCursor[int](nil, 10, false, false, func(item int) (string, error) {
		return contract.MarshalCursor(item)
	})

	require.NoError(t, err)
	require.NotNil(t, cursor.Items)
	require.Empty(t, cursor.Items)
}

func TestNewCursorPreservesPerPage(t *testing.T) {
	t.Parallel()

	cursor, err := contract.NewCursor([]int{1}, 25, false, false, func(item int) (string, error) {
		return contract.MarshalCursor(item)
	})

	require.NoError(t, err)
	require.Equal(t, 25, cursor.PerPage)
}

func TestNewCursorNextEncodeErrorReturnsErrCursorEncode(t *testing.T) {
	t.Parallel()

	_, err := contract.NewCursor([]int{1}, 10, true, false, func(item int) (string, error) {
		return "", errors.New("encode failed")
	})

	require.ErrorIs(t, err, contract.ErrCursorEncode)
}

func TestNewCursorPrevEncodeErrorReturnsErrCursorEncode(t *testing.T) {
	t.Parallel()

	_, err := contract.NewCursor([]int{1}, 10, false, true, func(item int) (string, error) {
		return "", errors.New("encode failed")
	})

	require.ErrorIs(t, err, contract.ErrCursorEncode)
}

func TestNewCursorCustomEncoder(t *testing.T) {
	t.Parallel()

	items := []int{1, 2, 3}
	cursor, err := contract.NewCursor(items, 3, true, false, func(item int) (string, error) {
		return "custom-cursor", nil
	})

	require.NoError(t, err)
	require.Equal(t, "custom-cursor", cursor.NextCursor)
}

func TestMarshalUnmarshalCursorRoundTrip(t *testing.T) {
	t.Parallel()

	encoded, err := contract.MarshalCursor(42)

	require.NoError(t, err)

	value, err := contract.UnmarshalCursor[int](encoded)

	require.NoError(t, err)
	require.Equal(t, 42, value)
}

func TestMarshalCursorStringRoundTrip(t *testing.T) {
	t.Parallel()

	encoded, err := contract.MarshalCursor("hello")

	require.NoError(t, err)

	value, err := contract.UnmarshalCursor[string](encoded)

	require.NoError(t, err)
	require.Equal(t, "hello", value)
}

func TestMarshalCursorStructRoundTrip(t *testing.T) {
	t.Parallel()

	type key struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}

	original := key{ID: 7, Name: "test"}
	encoded, err := contract.MarshalCursor(original)

	require.NoError(t, err)

	value, err := contract.UnmarshalCursor[key](encoded)

	require.NoError(t, err)
	require.Equal(t, original, value)
}

func TestMarshalCursorUnencodableReturnsError(t *testing.T) {
	t.Parallel()

	_, err := contract.MarshalCursor(make(chan int))

	require.Error(t, err)
}

func TestUnmarshalCursorInvalidBase64ReturnsErrCursorDecode(t *testing.T) {
	t.Parallel()

	_, err := contract.UnmarshalCursor[int]("not-valid-base64!!!")

	require.ErrorIs(t, err, contract.ErrCursorDecode)
}

func TestUnmarshalCursorInvalidJSONReturnsErrCursorDecode(t *testing.T) {
	t.Parallel()

	// Valid base64 but invalid JSON.
	_, err := contract.UnmarshalCursor[int]("bm90LWpzb24")

	require.ErrorIs(t, err, contract.ErrCursorDecode)
}

func TestNewCursorWithMarshalCursorEndToEnd(t *testing.T) {
	t.Parallel()

	type User struct {
		ID   int
		Name string
	}

	users := []User{{ID: 10, Name: "Alice"}, {ID: 20, Name: "Bob"}}

	cursor, err := contract.NewCursor(users, 2, true, false, func(u User) (string, error) {
		return contract.MarshalCursor(u.ID)
	})

	require.NoError(t, err)

	id, err := contract.UnmarshalCursor[int](cursor.NextCursor)

	require.NoError(t, err)
	require.Equal(t, 20, id)
}
