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
	cursor, err := contract.NewCursor(items, 3, true, false, func(item int) any { return item })

	require.NoError(t, err)
	require.NotEmpty(t, cursor.NextCursor)
	require.Empty(t, cursor.PrevCursor)
}

func TestNewCursorEncodesPrevCursor(t *testing.T) {
	t.Parallel()

	items := []int{1, 2, 3}
	cursor, err := contract.NewCursor(items, 3, false, true, func(item int) any { return item })

	require.NoError(t, err)
	require.Empty(t, cursor.NextCursor)
	require.NotEmpty(t, cursor.PrevCursor)
}

func TestNewCursorEncodesBothCursors(t *testing.T) {
	t.Parallel()

	items := []int{1, 2, 3}
	cursor, err := contract.NewCursor(items, 3, true, true, func(item int) any { return item })

	require.NoError(t, err)
	require.NotEmpty(t, cursor.NextCursor)
	require.NotEmpty(t, cursor.PrevCursor)
}

func TestNewCursorEmptyItemsNoCursors(t *testing.T) {
	t.Parallel()

	cursor, err := contract.NewCursor([]int{}, 10, true, true, func(item int) any { return item })

	require.NoError(t, err)
	require.Empty(t, cursor.NextCursor)
	require.Empty(t, cursor.PrevCursor)
}

func TestNewCursorNilItemsBecomesEmptySlice(t *testing.T) {
	t.Parallel()

	cursor, err := contract.NewCursor[int](nil, 10, false, false, func(item int) any { return item })

	require.NoError(t, err)
	require.NotNil(t, cursor.Items)
	require.Empty(t, cursor.Items)
}

func TestNewCursorPreservesPerPage(t *testing.T) {
	t.Parallel()

	cursor, err := contract.NewCursor([]int{1}, 25, false, false, func(item int) any { return item })

	require.NoError(t, err)
	require.Equal(t, 25, cursor.PerPage)
}

func TestNewCursorEncodeErrorReturnsErrCursorEncode(t *testing.T) {
	t.Parallel()

	items := []int{1}

	// Channels cannot be JSON-encoded.
	_, err := contract.NewCursor(items, 10, true, false, func(item int) any {
		return make(chan int)
	})

	require.ErrorIs(t, err, contract.ErrCursorEncode)
}

func TestNewCursorPrevEncodeErrorReturnsErrCursorEncode(t *testing.T) {
	t.Parallel()

	items := []int{1}

	_, err := contract.NewCursor(items, 10, false, true, func(item int) any {
		return make(chan int)
	})

	require.ErrorIs(t, err, contract.ErrCursorEncode)
}

func TestCursorValueRoundTrip(t *testing.T) {
	t.Parallel()

	items := []int{10, 20, 30}
	cursor, err := contract.NewCursor(items, 3, true, false, func(item int) any { return item })

	require.NoError(t, err)

	value, err := contract.CursorValue[float64](cursor.NextCursor)

	require.NoError(t, err)
	require.Equal(t, float64(30), value)
}

func TestCursorValueInvalidBase64ReturnsErrCursorDecode(t *testing.T) {
	t.Parallel()

	_, err := contract.CursorValue[int]("not-valid-base64!!!")

	require.ErrorIs(t, err, contract.ErrCursorDecode)
}

func TestCursorValueTypeMismatchReturnsErrCursorDecode(t *testing.T) {
	t.Parallel()

	items := []string{"hello"}
	cursor, err := contract.NewCursor(items, 1, true, false, func(item string) any { return item })

	require.NoError(t, err)

	_, err = contract.CursorValue[int](cursor.NextCursor)

	require.ErrorIs(t, err, contract.ErrCursorDecode)
}

type failEncoder struct{}

func (failEncoder) Encode(value any) (string, error) {
	return "", errors.New("encode failed")
}

func (failEncoder) Decode(cursor string) (any, error) {
	return nil, errors.New("decode failed")
}

type idEncoder struct{}

func (idEncoder) Encode(value any) (string, error) {
	return "custom-cursor", nil
}

func (idEncoder) Decode(cursor string) (any, error) {
	return cursor, nil
}

func TestNewCursorWithCustomEncoder(t *testing.T) {
	t.Parallel()

	items := []int{1, 2, 3}
	cursor, err := contract.NewCursorWith(items, 3, true, false, idEncoder{})

	require.NoError(t, err)
	require.Equal(t, "custom-cursor", cursor.NextCursor)
}

func TestNewCursorWithEncoderErrorReturnsErrCursorEncode(t *testing.T) {
	t.Parallel()

	items := []int{1}
	_, err := contract.NewCursorWith(items, 1, true, false, failEncoder{})

	require.ErrorIs(t, err, contract.ErrCursorEncode)
}

func TestNewCursorWithPrevEncoderErrorReturnsErrCursorEncode(t *testing.T) {
	t.Parallel()

	items := []int{1}
	_, err := contract.NewCursorWith(items, 1, false, true, failEncoder{})

	require.ErrorIs(t, err, contract.ErrCursorEncode)
}

func TestNewCursorWithEmptyItems(t *testing.T) {
	t.Parallel()

	cursor, err := contract.NewCursorWith([]int{}, 10, true, true, idEncoder{})

	require.NoError(t, err)
	require.Empty(t, cursor.NextCursor)
	require.Empty(t, cursor.PrevCursor)
}

func TestNewCursorWithNilItemsBecomesEmptySlice(t *testing.T) {
	t.Parallel()

	cursor, err := contract.NewCursorWith[int](nil, 10, false, false, idEncoder{})

	require.NoError(t, err)
	require.NotNil(t, cursor.Items)
}

func TestNewCursorWithEncodesBothCursors(t *testing.T) {
	t.Parallel()

	items := []int{1, 2, 3}
	cursor, err := contract.NewCursorWith(items, 3, true, true, idEncoder{})

	require.NoError(t, err)
	require.Equal(t, "custom-cursor", cursor.NextCursor)
	require.Equal(t, "custom-cursor", cursor.PrevCursor)
}

func TestCursorValueInvalidJSONReturnsErrCursorDecode(t *testing.T) {
	t.Parallel()

	// Valid base64 but invalid JSON.
	_, err := contract.CursorValue[int]("bm90LWpzb24")

	require.ErrorIs(t, err, contract.ErrCursorDecode)
}

func TestCursorValueWithCustomDecoder(t *testing.T) {
	t.Parallel()

	value, err := contract.CursorValueWith("test-cursor", idEncoder{})

	require.NoError(t, err)
	require.Equal(t, "test-cursor", value)
}

func TestCursorValueWithDecoderErrorReturnsErrCursorDecode(t *testing.T) {
	t.Parallel()

	_, err := contract.CursorValueWith("anything", failEncoder{})

	require.ErrorIs(t, err, contract.ErrCursorDecode)
}
