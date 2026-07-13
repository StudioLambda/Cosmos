package collection_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/studiolambda/cosmos/collection"
)

func TestMapTryReturnsTryLazyMapWithAllItems(t *testing.T) {
	t.Parallel()

	items, err := collection.NewMap(map[string]int{"a": 1, "b": 2}).Try().Items()

	require.NoError(t, err)
	require.Equal(t, map[string]int{"a": 1, "b": 2}, items)
}

func TestTryLazyMapMapValuesPropagatesErrors(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("boom")
	items, err := collection.NewMap(map[string]int{"a": 1, "b": 2}).Try().MapValues(func(k string, v int) (int, error) {
		if v == 2 {
			return 0, wantErr
		}

		return v * 10, nil
	}).Items()

	require.ErrorIs(t, err, wantErr)
	require.Equal(t, map[string]int{"a": 10}, items)
}

func TestTryLazyMapEachAllJoinsErrors(t *testing.T) {
	t.Parallel()

	errA := errors.New("a")
	errB := errors.New("b")
	tryMap := collection.NewTryLazyMap(func(yield func(collection.MapEntry[string, int], error) bool) {
		yield(collection.NewMapEntry("a", 1), nil)
		yield(collection.MapEntry[string, int]{}, errA)
		yield(collection.NewMapEntry("b", 2), nil)
	})

	err := tryMap.EachAll(func(k string, v int) error {
		if k == "b" {
			return errB
		}

		return nil
	})

	require.ErrorIs(t, err, errA)
	require.ErrorIs(t, err, errB)
}
