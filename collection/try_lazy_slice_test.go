package collection_test

import (
	"bytes"
	"cmp"
	"encoding/json"
	"encoding/json/jsontext"
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/studiolambda/cosmos/collection"

	"github.com/stretchr/testify/require"
)

// TestSliceTryReturnsTryLazySliceWithAllItems verifies that Slice.Try bridges eager slices
// into TryLazySlice without introducing errors.
func TestSliceTryReturnsTryLazySliceWithAllItems(t *testing.T) {
	t.Parallel()

	items, err := collection.NewSlice([]int{1, 2, 3}).Try().Items()

	require.NoError(t, err)
	require.Equal(t, []int{1, 2, 3}, items)
}

// TestLazySliceTryReturnsTryLazySliceWithAllItems verifies that LazySlice.Try bridges lazy slices
// into TryLazySlice without introducing errors.
func TestLazySliceTryReturnsTryLazySliceWithAllItems(t *testing.T) {
	t.Parallel()

	items, err := collection.NewLazySlice(slices.All([]int{4, 5, 6})).Try().Items()

	require.NoError(t, err)
	require.Equal(t, []int{4, 5, 6}, items)
}

// TestTryLazySliceCanBeUsedDirectlyInForRangeLoop verifies that TryLazySlice[T]
// can be ranged over directly as a Seq2.
func TestTryLazySliceCanBeUsedDirectlyInForRangeLoop(t *testing.T) {
	t.Parallel()

	tryLazy := collection.NewSlice([]int{10, 20, 30}).Try()

	var got []int
	for v, err := range tryLazy {
		require.NoError(t, err)
		got = append(got, v)
	}

	require.Equal(t, []int{10, 20, 30}, got)
}

// TestTryLazySliceMapIsLazy verifies that Map on TryLazySlice defers callback execution
// until the sequence is consumed.
func TestTryLazySliceMapIsLazy(t *testing.T) {
	t.Parallel()

	called := 0
	tryLazy := collection.NewSlice([]int{1, 2, 3}).Try()
	mapped := tryLazy.Map(func(i int, v int) (int, error) {
		called++

		return v * 10, nil
	})

	require.Equal(t, 0, called)

	items, err := mapped.Items()

	require.NoError(t, err)
	require.Equal(t, []int{10, 20, 30}, items)
	require.Equal(t, 3, called)
}

// TestTryLazySliceFilterIsLazy verifies that Filter on TryLazySlice defers predicate execution
// until the sequence is consumed.
func TestTryLazySliceFilterIsLazy(t *testing.T) {
	t.Parallel()

	called := 0
	tryLazy := collection.NewSlice([]int{1, 2, 3, 4}).Try()
	filtered := tryLazy.Filter(func(i int, v int) (bool, error) {
		called++

		return v%2 == 0, nil
	})

	require.Equal(t, 0, called)

	items, err := filtered.Items()

	require.NoError(t, err)
	require.Equal(t, []int{2, 4}, items)
	require.Equal(t, 4, called)
}

// TestTryLazySliceMapPropagatesMapperErrors verifies that Map yields mapper errors and
// allows a terminal operation to fail fast.
func TestTryLazySliceMapPropagatesMapperErrors(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("boom")
	items, err := collection.NewSlice([]int{1, 2, 3}).Try().Map(func(i int, v int) (int, error) {
		if v == 2 {
			return 0, wantErr
		}

		return v * 10, nil
	}).Items()

	require.ErrorIs(t, err, wantErr)
	require.Equal(t, []int{10}, items)
}

// TestTryLazySliceItemsAllJoinsErrors verifies that ItemsAll consumes the entire sequence,
// preserving valid items and joining all errors.
func TestTryLazySliceItemsAllJoinsErrors(t *testing.T) {
	t.Parallel()

	errA := errors.New("a")
	errB := errors.New("b")
	items, err := collection.NewTryLazySlice(func(yield func(int, error) bool) {
		yield(1, nil)
		yield(0, errA)
		yield(2, nil)
		yield(0, errB)
	}).ItemsAll()

	require.Equal(t, []int{1, 2}, items)
	require.ErrorIs(t, err, errA)
	require.ErrorIs(t, err, errB)
}

// TestTryLazySliceEachStopsOnCallbackError verifies that Each aborts the underlying sequence
// when the callback returns an error.
func TestTryLazySliceEachStopsOnCallbackError(t *testing.T) {
	t.Parallel()

	closed := false
	wantErr := errors.New("stop")
	tryLazy := collection.NewTryLazySlice(func(yield func(int, error) bool) {
		defer func() { closed = true }()

		for _, v := range []int{1, 2, 3} {
			if !yield(v, nil) {
				return
			}
		}
	})

	err := tryLazy.Each(func(i int, v int) error {
		if v == 2 {
			return wantErr
		}

		return nil
	})

	require.ErrorIs(t, err, wantErr)
	require.True(t, closed)
}

// TestTryLazySliceEachAllJoinsIteratorAndCallbackErrors verifies that EachAll consumes
// the entire sequence and joins both iterator and callback errors.
func TestTryLazySliceEachAllJoinsIteratorAndCallbackErrors(t *testing.T) {
	t.Parallel()

	errA := errors.New("iterator")
	errB := errors.New("callback")
	tryLazy := collection.NewTryLazySlice(func(yield func(int, error) bool) {
		yield(1, nil)
		yield(0, errA)
		yield(2, nil)
	})

	err := tryLazy.EachAll(func(i int, v int) error {
		if v == 2 {
			return errB
		}

		return nil
	})

	require.ErrorIs(t, err, errA)
	require.ErrorIs(t, err, errB)
}

// TestTryLazySliceTakeCountsOnlySuccessfulItems verifies that Take limits successful items
// without consuming past the requested count.
func TestTryLazySliceTakeCountsOnlySuccessfulItems(t *testing.T) {
	t.Parallel()

	errA := errors.New("iterator")
	items, err := collection.NewTryLazySlice(func(yield func(int, error) bool) {
		if !yield(1, nil) {
			return
		}

		if !yield(0, errA) {
			return
		}

		if !yield(2, nil) {
			return
		}

		if !yield(3, nil) {
			return
		}

		_ = yield(4, nil)
	}).Take(2).ItemsAll()

	require.Equal(t, []int{1, 2}, items)
	require.ErrorIs(t, err, errA)
}

// TestTryLazySliceSkipSkipsOnlySuccessfulItems verifies that Skip ignores iterator errors
// when counting items to skip and still forwards them.
func TestTryLazySliceSkipSkipsOnlySuccessfulItems(t *testing.T) {
	t.Parallel()

	errA := errors.New("iterator")
	items, err := collection.NewTryLazySlice(func(yield func(int, error) bool) {
		if !yield(1, nil) {
			return
		}

		if !yield(0, errA) {
			return
		}

		if !yield(2, nil) {
			return
		}

		_ = yield(3, nil)
	}).Skip(1).ItemsAll()

	require.Equal(t, []int{2, 3}, items)
	require.ErrorIs(t, err, errA)
}

// TestTryLazySliceReduceStopsOnFirstError verifies that Reduce fail-fast behavior returns
// the accumulator state collected before the error.
func TestTryLazySliceReduceStopsOnFirstError(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("boom")
	result, err := collection.NewSlice([]int{1, 2, 3}).Try().Reduce(func(acc int, i int, v int) (int, error) {
		if v == 3 {
			return acc, wantErr
		}

		return acc + v, nil
	}, 0)

	require.Equal(t, 3, result)
	require.ErrorIs(t, err, wantErr)
}

// TestTryLazySliceFirstReturnsFirstSuccessfulItem verifies that First ignores later items
// and returns the first successful one.
func TestTryLazySliceFirstReturnsFirstSuccessfulItem(t *testing.T) {
	t.Parallel()

	result, ok, err := collection.NewSlice([]int{7, 8, 9}).Try().First()

	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, 7, result)
}

// TestTryLazySliceFirstReturnsErrorBeforeAnyValue verifies that First returns an error
// when the sequence begins with an error.
func TestTryLazySliceFirstReturnsErrorBeforeAnyValue(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("boom")
	_, ok, err := collection.NewTryLazySlice(func(yield func(int, error) bool) {
		yield(0, wantErr)
	}).First()

	require.False(t, ok)
	require.ErrorIs(t, err, wantErr)
}

// TestTryLazySliceContainsPropagatesPredicateError verifies that Contains returns predicate errors.
func TestTryLazySliceContainsPropagatesPredicateError(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("boom")
	contains, err := collection.NewSlice([]int{1, 2, 3}).Try().Contains(func(i int, v int) (bool, error) {
		if v == 2 {
			return false, wantErr
		}

		return v == 3, nil
	})

	require.False(t, contains)
	require.ErrorIs(t, err, wantErr)
}

// TestTryLazySlicePartitionReturnsSeparatedResults verifies that Partition materializes both sides
// and returns them as independently iterable TryLazySlice values.
func TestTryLazySlicePartitionReturnsSeparatedResults(t *testing.T) {
	t.Parallel()

	even, odd := collection.NewSlice([]int{1, 2, 3, 4, 5}).Try().Partition(func(i int, v int) (bool, error) {
		return v%2 == 0, nil
	})

	evenItems, err := even.Items()
	require.NoError(t, err)

	oddItems, err := odd.Items()
	require.NoError(t, err)

	require.Equal(t, []int{2, 4}, evenItems)
	require.Equal(t, []int{1, 3, 5}, oddItems)
}

// TestTryLazySliceConcatYieldsValuesAndErrorsInOrder verifies that Concat preserves order
// for both items and iterator errors.
func TestTryLazySliceConcatYieldsValuesAndErrorsInOrder(t *testing.T) {
	t.Parallel()

	errA := errors.New("a")
	errB := errors.New("b")
	first := collection.NewTryLazySlice(func(yield func(int, error) bool) {
		if !yield(1, nil) {
			return
		}

		_ = yield(0, errA)
	})
	second := collection.NewTryLazySlice(func(yield func(int, error) bool) {
		if !yield(2, nil) {
			return
		}

		_ = yield(0, errB)
	})

	items, err := first.Concat(second).ItemsAll()

	require.Equal(t, []int{1, 2}, items)
	require.ErrorIs(t, err, errA)
	require.ErrorIs(t, err, errB)
}

// TestTryLazySliceSortReturnsSortedItems verifies that Sort materializes and sorts successful items.
func TestTryLazySliceSortReturnsSortedItems(t *testing.T) {
	t.Parallel()

	items, err := collection.NewSlice([]int{3, 1, 4, 1, 5}).Try().Sort(cmp.Compare[int]).Items()

	require.NoError(t, err)
	require.Equal(t, []int{1, 1, 3, 4, 5}, items)
}

// TestTryLazySliceReverseReturnsReversedItems verifies that Reverse materializes and reverses successful items.
func TestTryLazySliceReverseReturnsReversedItems(t *testing.T) {
	t.Parallel()

	items, err := collection.NewSlice([]int{1, 2, 3}).Try().Reverse().Items()

	require.NoError(t, err)
	require.Equal(t, []int{3, 2, 1}, items)
}

// TestTryLazySliceUniquePropagatesKeyErrors verifies that Unique yields key errors and keeps
// only the first item for each key.
func TestTryLazySliceUniquePropagatesKeyErrors(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("boom")
	items, err := collection.NewSlice([]int{1, 2, 2, 3}).Try().Unique(func(i int, v int) (int, error) {
		if v == 3 {
			return 0, wantErr
		}

		return v, nil
	}).Items()

	require.Equal(t, []int{1, 2}, items)
	require.ErrorIs(t, err, wantErr)
}

// TestTryLazySliceChunkYieldsChunksAndErrors verifies that Chunk groups successful items
// while preserving iterator errors.
func TestTryLazySliceChunkYieldsChunksAndErrors(t *testing.T) {
	t.Parallel()

	errA := errors.New("iterator")
	seq := collection.NewTryLazySlice(func(yield func(int, error) bool) {
		if !yield(1, nil) {
			return
		}

		if !yield(2, nil) {
			return
		}

		if !yield(0, errA) {
			return
		}

		_ = yield(3, nil)
	}).Chunk(2)

	var chunks [][]int
	var errs []error
	for chunk, err := range seq {
		if err != nil {
			errs = append(errs, err)

			continue
		}

		chunks = append(chunks, slices.Clone(chunk))
	}

	require.Equal(t, [][]int{{1, 2}, {3}}, chunks)
	require.ErrorIs(t, errors.Join(errs...), errA)
}

// TestTryLazySliceSlidingYieldsWindowsAndErrors verifies that Sliding windows only successful items
// while preserving iterator errors.
func TestTryLazySliceSlidingYieldsWindowsAndErrors(t *testing.T) {
	t.Parallel()

	errA := errors.New("iterator")
	seq := collection.NewTryLazySlice(func(yield func(int, error) bool) {
		if !yield(1, nil) {
			return
		}

		if !yield(2, nil) {
			return
		}

		if !yield(0, errA) {
			return
		}

		if !yield(3, nil) {
			return
		}

		_ = yield(4, nil)
	}).Sliding(2)

	var windows [][]int
	var errs []error
	for window, err := range seq {
		if err != nil {
			errs = append(errs, err)

			continue
		}

		windows = append(windows, slices.Clone(window))
	}

	require.Equal(t, [][]int{{1, 2}, {2, 3}, {3, 4}}, windows)
	require.ErrorIs(t, errors.Join(errs...), errA)
}

// TestTryLazySliceGroupByReturnsGroupedItems verifies that GroupBy materializes grouped results.
func TestTryLazySliceGroupByReturnsGroupedItems(t *testing.T) {
	t.Parallel()

	grouped, err := collection.NewSlice([]int{1, 2, 3, 4}).Try().GroupBy(func(i int, v int) (string, error) {
		if v%2 == 0 {
			return "even", nil
		}

		return "odd", nil
	})

	require.NoError(t, err)

	odd, ok := grouped.Get("odd")
	require.True(t, ok)
	require.Equal(t, []int{1, 3}, odd.Items())

	even, ok := grouped.Get("even")
	require.True(t, ok)
	require.Equal(t, []int{2, 4}, even.Items())
}

// TestTryLazySliceMarshalJSONReturnsErrorOnIteratorError verifies that JSON encoding
// fails when materialization encounters an error.
func TestTryLazySliceMarshalJSONReturnsErrorOnIteratorError(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("boom")
	_, err := collection.NewTryLazySlice(func(yield func(int, error) bool) {
		yield(1, nil)
		yield(0, wantErr)
	}).MarshalJSON()

	require.ErrorIs(t, err, wantErr)
}

// TestTryLazySliceMarshalJSONMarshalsSuccessfulSequence verifies that TryLazySlice marshals
// exactly like a normal slice when no errors are yielded.
func TestTryLazySliceMarshalJSONMarshalsSuccessfulSequence(t *testing.T) {
	t.Parallel()

	data, err := collection.NewSlice([]int{1, 2, 3}).Try().MarshalJSON()

	require.NoError(t, err)
	require.JSONEq(t, `[1,2,3]`, string(data))
}

// TestTryLazySliceMarshalJSONToEncodesSuccessfulSequence verifies that MarshalJSONTo writes JSON output.
func TestTryLazySliceMarshalJSONToEncodesSuccessfulSequence(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	enc := jsontext.NewEncoder(&buf)

	err := collection.NewSlice([]int{4, 5, 6}).Try().MarshalJSONTo(enc)

	require.NoError(t, err)
	require.JSONEq(t, `[4,5,6]`, buf.String())
}

// TestTryLazySliceUnmarshalJSONDecodesJSONArrayAndYieldsItems verifies that UnmarshalJSON
// populates a TryLazySlice from JSON.
func TestTryLazySliceUnmarshalJSONDecodesJSONArrayAndYieldsItems(t *testing.T) {
	t.Parallel()

	var tryLazy collection.TryLazySlice[int]

	require.NoError(t, tryLazy.UnmarshalJSON([]byte(`[10,20,30]`)))

	items, err := tryLazy.Items()

	require.NoError(t, err)
	require.Equal(t, []int{10, 20, 30}, items)
}

// TestTryLazySliceUnmarshalJSONFromDecodesJSONArrayAndYieldsItems verifies that UnmarshalJSONFrom
// populates a TryLazySlice from a decoder.
func TestTryLazySliceUnmarshalJSONFromDecodesJSONArrayAndYieldsItems(t *testing.T) {
	t.Parallel()

	dec := jsontext.NewDecoder(strings.NewReader(`[7,8,9]`))

	var tryLazy collection.TryLazySlice[int]

	require.NoError(t, tryLazy.UnmarshalJSONFrom(dec))

	items, err := tryLazy.Items()

	require.NoError(t, err)
	require.Equal(t, []int{7, 8, 9}, items)
}

// TestTryLazySliceEagerAllReturnsSliceAndJoinedErrors verifies that EagerAll returns the eager collection form
// together with joined errors.
func TestTryLazySliceEagerAllReturnsSliceAndJoinedErrors(t *testing.T) {
	t.Parallel()

	errA := errors.New("a")
	errB := errors.New("b")
	eager, err := collection.NewTryLazySlice(func(yield func(int, error) bool) {
		if !yield(1, nil) {
			return
		}

		if !yield(0, errA) {
			return
		}

		if !yield(2, nil) {
			return
		}

		_ = yield(0, errB)
	}).EagerAll()

	require.Equal(t, []int{1, 2}, eager.Items())
	require.ErrorIs(t, err, errA)
	require.ErrorIs(t, err, errB)
}

// TestTryLazySliceSupportsMethodChaining verifies that the error-aware lazy API preserves fluent chains.
func TestTryLazySliceSupportsMethodChaining(t *testing.T) {
	t.Parallel()

	items, err := collection.NewSlice([]int{1, 2, 3, 4, 5}).
		Try().
		Filter(func(i int, v int) (bool, error) { return v%2 == 1, nil }).
		Map(func(i int, v int) (string, error) { return strings.Repeat("x", v), nil }).
		Take(2).
		Items()

	require.NoError(t, err)
	require.Equal(t, []string{"x", "xxx"}, items)
}

// TestTryLazySliceStopsGeneratorOnFirstErrorDuringItems verifies that fail-fast materialization
// stops the source and triggers its cleanup immediately.
func TestTryLazySliceStopsGeneratorOnFirstErrorDuringItems(t *testing.T) {
	t.Parallel()

	closed := false
	wantErr := errors.New("scan failed")
	items, err := collection.NewTryLazySlice(func(yield func(int, error) bool) {
		defer func() { closed = true }()

		if !yield(1, nil) {
			return
		}

		if !yield(0, wantErr) {
			return
		}

		_ = yield(2, nil)
	}).Items()

	require.Equal(t, []int{1}, items)
	require.ErrorIs(t, err, wantErr)
	require.True(t, closed)
}

// TestTryLazySliceUnmarshalJSONReturnsErrorOnInvalidJSON verifies that UnmarshalJSON rejects malformed JSON.
func TestTryLazySliceUnmarshalJSONReturnsErrorOnInvalidJSON(t *testing.T) {
	t.Parallel()

	var tryLazy collection.TryLazySlice[int]

	require.Error(t, tryLazy.UnmarshalJSON([]byte(`not-json`)))
}

// TestTryLazySliceUnmarshalJSONFromReturnsErrorOnInvalidJSON verifies that UnmarshalJSONFrom rejects malformed JSON.
func TestTryLazySliceUnmarshalJSONFromReturnsErrorOnInvalidJSON(t *testing.T) {
	t.Parallel()

	dec := jsontext.NewDecoder(strings.NewReader(`not-json`))

	var tryLazy collection.TryLazySlice[int]

	require.Error(t, tryLazy.UnmarshalJSONFrom(dec))
}

// TestTryLazySliceMarshalJSONRoundTripsWithEncodingJSON verifies encoding/json compatibility.
func TestTryLazySliceMarshalJSONRoundTripsWithEncodingJSON(t *testing.T) {
	t.Parallel()

	data, err := json.Marshal(collection.NewSlice([]int{1, 2, 3}).Try())

	require.NoError(t, err)

	var decoded []int
	require.NoError(t, json.Unmarshal(data, &decoded))
	require.Equal(t, []int{1, 2, 3}, decoded)
}
