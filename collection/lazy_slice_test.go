package collection_test

import (
	"bytes"
	"cmp"
	"encoding/json"
	"encoding/json/jsontext"
	"slices"
	"strings"
	"testing"

	"github.com/studiolambda/cosmos/collection"

	"github.com/stretchr/testify/require"
)

// TestNewLazySliceCreatesLazySliceFromFunctionAndYieldsAllItems verifies that
// NewLazySlice wraps the given iterator function and yields all items when consumed.
func TestNewLazySliceCreatesLazySliceFromFunctionAndYieldsAllItems(t *testing.T) {
	t.Parallel()

	lazy := collection.NewLazySlice(slices.All([]int{1, 2, 3}))

	require.Equal(t, []int{1, 2, 3}, lazy.Items())
}

// TestLazySliceCanBeUsedDirectlyInForRangeLoop verifies that LazySlice[T] is a
// function type that can be ranged over without any adapter method.
func TestLazySliceCanBeUsedDirectlyInForRangeLoop(t *testing.T) {
	t.Parallel()

	lazy := collection.NewLazySlice(slices.All([]int{10, 20, 30}))

	var got []int
	for _, v := range lazy {
		got = append(got, v)
	}

	require.Equal(t, []int{10, 20, 30}, got)
}

// TestLazySliceEagerMaterializesIntoEagerCollection verifies that Eager
// returns an eager Slice containing all yielded items.
func TestLazySliceEagerMaterializesIntoEagerCollection(t *testing.T) {
	t.Parallel()

	lazy := collection.NewLazySlice(slices.All([]int{5, 6, 7}))

	c := lazy.Eager()

	require.Equal(t, []int{5, 6, 7}, c.Items())
}

// TestLazySliceItemsMaterializesAllItems verifies that Items collects every item
// into a plain Go slice.
func TestLazySliceItemsMaterializesAllItems(t *testing.T) {
	t.Parallel()

	lazy := collection.NewLazySlice(slices.All([]string{"a", "b", "c"}))

	require.Equal(t, []string{"a", "b", "c"}, lazy.Items())
}

// TestLazySliceIsEmptyReturnsTrueForEmptyLazy verifies that IsEmpty is true when
// the iterator yields nothing.
func TestLazySliceIsEmptyReturnsTrueForEmptyLazy(t *testing.T) {
	t.Parallel()

	lazy := collection.NewLazySlice(slices.All([]int{}))

	require.True(t, lazy.IsEmpty())
}

// TestLazySliceIsEmptyReturnsFalseForNonEmptyLazy verifies that IsEmpty is false
// when the iterator yields at least one item.
func TestLazySliceIsEmptyReturnsFalseForNonEmptyLazy(t *testing.T) {
	t.Parallel()

	lazy := collection.NewLazySlice(slices.All([]int{1}))

	require.False(t, lazy.IsEmpty())
}

// TestLazySliceIsNotEmptyReturnsTrueForNonEmptyLazy verifies that IsNotEmpty is
// true when the iterator yields at least one item.
func TestLazySliceIsNotEmptyReturnsTrueForNonEmptyLazy(t *testing.T) {
	t.Parallel()

	lazy := collection.NewLazySlice(slices.All([]int{42}))

	require.True(t, lazy.IsNotEmpty())
}

// TestLazySliceIsNotEmptyReturnsFalseForEmptyLazy verifies that IsNotEmpty is
// false when the iterator yields nothing.
func TestLazySliceIsNotEmptyReturnsFalseForEmptyLazy(t *testing.T) {
	t.Parallel()

	lazy := collection.NewLazySlice(slices.All([]int{}))

	require.False(t, lazy.IsNotEmpty())
}

// TestLazySliceLenReturnsTotalCount verifies that Len returns the number of items
// yielded by the sequence.
func TestLazySliceLenReturnsTotalCount(t *testing.T) {
	t.Parallel()

	lazy := collection.NewLazySlice(slices.All([]int{1, 2, 3, 4, 5}))

	require.Equal(t, 5, lazy.Len())
}

// TestLazySliceLenReturnsZeroForEmpty verifies that Len is 0 for an empty lazy.
func TestLazySliceLenReturnsZeroForEmpty(t *testing.T) {
	t.Parallel()

	lazy := collection.NewLazySlice(slices.All([]int{}))

	require.Equal(t, 0, lazy.Len())
}

// TestLazySliceEveryReturnsTrueWhenAllItemsMatch verifies that Every returns true
// when the predicate holds for every item.
func TestLazySliceEveryReturnsTrueWhenAllItemsMatch(t *testing.T) {
	t.Parallel()

	lazy := collection.NewLazySlice(slices.All([]int{2, 4, 6}))

	require.True(t, lazy.Every(func(i int, v int) bool { return v%2 == 0 }))
}

// TestLazySliceEveryReturnsFalseWhenOneItemFails verifies that Every returns false
// as soon as any item fails the predicate.
func TestLazySliceEveryReturnsFalseWhenOneItemFails(t *testing.T) {
	t.Parallel()

	lazy := collection.NewLazySlice(slices.All([]int{2, 3, 6}))

	require.False(t, lazy.Every(func(i int, v int) bool { return v%2 == 0 }))
}

// TestLazySliceEveryReturnsTrueOnEmpty verifies the vacuous truth: Every on an
// empty sequence returns true.
func TestLazySliceEveryReturnsTrueOnEmpty(t *testing.T) {
	t.Parallel()

	lazy := collection.NewLazySlice(slices.All([]int{}))

	require.True(t, lazy.Every(func(i int, v int) bool { return false }))
}

// TestLazySliceEachCallsFForEveryItem verifies that Each invokes f once per item.
func TestLazySliceEachCallsFForEveryItem(t *testing.T) {
	t.Parallel()

	lazy := collection.NewLazySlice(slices.All([]int{1, 2, 3}))

	var got []int
	lazy.Each(func(i int, v int) { got = append(got, v) })

	require.Equal(t, []int{1, 2, 3}, got)
}

// TestLazySliceEachDoesNotCallFOnEmpty verifies that Each never calls f when the
// sequence is empty.
func TestLazySliceEachDoesNotCallFOnEmpty(t *testing.T) {
	t.Parallel()

	lazy := collection.NewLazySlice(slices.All([]int{}))

	called := false
	lazy.Each(func(i int, v int) { called = true })

	require.False(t, called)
}

// TestLazySliceTapEachCallsFForEveryItemWhenConsumed verifies that TapEach calls
// f for each item and still yields the same items.
func TestLazySliceTapEachCallsFForEveryItemWhenConsumed(t *testing.T) {
	t.Parallel()

	lazy := collection.NewLazySlice(slices.All([]int{7, 8, 9}))

	var tapped []int
	result := lazy.TapEach(func(i int, v int) { tapped = append(tapped, v) })

	require.Empty(t, tapped) // not yet consumed

	got := result.Items()

	require.Equal(t, []int{7, 8, 9}, got)
	require.Equal(t, []int{7, 8, 9}, tapped)
}

// TestLazySliceTapEachIsLazyAndDoesNotCallFBeforeConsumption verifies that the
// side-effect function is not invoked until the returned LazySlice is consumed.
func TestLazySliceTapEachIsLazyAndDoesNotCallFBeforeConsumption(t *testing.T) {
	t.Parallel()

	lazy := collection.NewLazySlice(slices.All([]int{1, 2, 3}))

	callCount := 0
	tapped := lazy.TapEach(func(i int, v int) { callCount++ })

	require.Equal(t, 0, callCount) // not called yet

	_ = tapped.Items()

	require.Equal(t, 3, callCount)
}

// TestLazySliceFilterReturnsMatchingItems verifies that Filter keeps only the items
// for which the predicate returns true.
func TestLazySliceFilterReturnsMatchingItems(t *testing.T) {
	t.Parallel()

	lazy := collection.NewLazySlice(slices.All([]int{1, 2, 3, 4, 5}))

	got := lazy.Filter(func(i int, v int) bool { return v%2 == 0 }).Items()

	require.Equal(t, []int{2, 4}, got)
}

// TestLazySliceFilterReturnsEmptyWhenNoneMatch verifies that Filter returns an
// empty sequence when no items satisfy the predicate.
func TestLazySliceFilterReturnsEmptyWhenNoneMatch(t *testing.T) {
	t.Parallel()

	lazy := collection.NewLazySlice(slices.All([]int{1, 3, 5}))

	got := lazy.Filter(func(i int, v int) bool { return v%2 == 0 }).Items()

	require.Empty(t, got)
}

// TestLazySliceFilterReturnsEmptyOnEmptyInput verifies that Filter on an empty
// source yields nothing.
func TestLazySliceFilterReturnsEmptyOnEmptyInput(t *testing.T) {
	t.Parallel()

	lazy := collection.NewLazySlice(slices.All([]int{}))

	got := lazy.Filter(func(i int, v int) bool { return true }).Items()

	require.Empty(t, got)
}

// TestLazySliceRejectReturnsItemsNotMatchingPredicate verifies that Reject keeps
// only the items for which the predicate returns false.
func TestLazySliceRejectReturnsItemsNotMatchingPredicate(t *testing.T) {
	t.Parallel()

	lazy := collection.NewLazySlice(slices.All([]int{1, 2, 3, 4, 5}))

	got := lazy.Reject(func(i int, v int) bool { return v%2 == 0 }).Items()

	require.Equal(t, []int{1, 3, 5}, got)
}

// TestLazySliceMapTransformsItemsToSameType verifies that Map applies the function
// to each item and collects the results.
func TestLazySliceMapTransformsItemsToSameType(t *testing.T) {
	t.Parallel()

	lazy := collection.NewLazySlice(slices.All([]int{1, 2, 3}))

	got := lazy.Map(func(i int, v int) int { return v * 10 }).Items()

	require.Equal(t, []int{10, 20, 30}, got)
}

// TestLazySliceMapTransformsItemsToDifferentType verifies that Map can produce a
// lazy of a different element type (int → bool).
func TestLazySliceMapTransformsItemsToDifferentType(t *testing.T) {
	t.Parallel()

	lazy := collection.NewLazySlice(slices.All([]int{2, 3, 4}))

	got := lazy.Map(func(i int, v int) bool { return v%2 == 0 }).Items()

	require.Equal(t, []bool{true, false, true}, got)
}

// TestLazySliceFlatMapFlattensMappedResults verifies that FlatMap yields all mapped items in order.
func TestLazySliceFlatMapFlattensMappedResults(t *testing.T) {
	t.Parallel()

	lazy := collection.NewLazySlice(slices.All([]int{1, 2, 3}))

	got := lazy.FlatMap(func(i int, v int) []int {
		return []int{v, v * 10}
	}).Items()

	require.Equal(t, []int{1, 10, 2, 20, 3, 30}, got)
}

// TestLazySliceFlatMapIsLazy verifies that FlatMap does not consume the source until needed.
func TestLazySliceFlatMapIsLazy(t *testing.T) {
	t.Parallel()

	called := 0
	lazy := collection.NewLazySlice(slices.All([]int{1, 2, 3}))
	flatMapped := lazy.FlatMap(func(i int, v int) []int {
		called++

		return []int{v}
	})

	require.Equal(t, 0, called)
	_ = flatMapped.Items()
	require.Equal(t, 3, called)
}

// TestLazySliceKeyByIndexesItemsByDerivedKey verifies that KeyBy materializes a lookup map.
func TestLazySliceKeyByIndexesItemsByDerivedKey(t *testing.T) {
	t.Parallel()

	type user struct {
		ID   int
		Name string
	}

	lazy := collection.NewLazySlice(slices.All([]user{{ID: 1, Name: "alice"}, {ID: 2, Name: "bob"}}))

	result := lazy.KeyBy(func(i int, v user) int { return v.ID })

	require.Equal(t, map[int]user{
		1: {ID: 1, Name: "alice"},
		2: {ID: 2, Name: "bob"},
	}, result.Items())
}

// TestLazySliceCountByCountsItemsPerDerivedKey verifies that CountBy groups counts by key.
func TestLazySliceCountByCountsItemsPerDerivedKey(t *testing.T) {
	t.Parallel()

	lazy := collection.NewLazySlice(slices.All([]string{"ant", "ape", "bat", "bee"}))

	result := lazy.CountBy(func(i int, v string) byte { return v[0] })

	require.Equal(t, map[byte]int{'a': 2, 'b': 2}, result.Items())
}

// TestLazySliceTakeUntilStopsBeforeMatchingItem verifies that TakeUntil excludes the matching item.
func TestLazySliceTakeUntilStopsBeforeMatchingItem(t *testing.T) {
	t.Parallel()

	lazy := collection.NewLazySlice(slices.All([]int{1, 2, 3, 4, 5}))

	got := lazy.TakeUntil(func(i int, v int) bool { return v >= 4 }).Items()

	require.Equal(t, []int{1, 2, 3}, got)
}

// TestLazySliceSkipUntilStartsAtFirstMatchingItem verifies that SkipUntil yields from the first match onward.
func TestLazySliceSkipUntilStartsAtFirstMatchingItem(t *testing.T) {
	t.Parallel()

	lazy := collection.NewLazySlice(slices.All([]int{1, 2, 3, 4, 5}))

	got := lazy.SkipUntil(func(i int, v int) bool { return v >= 4 }).Items()

	require.Equal(t, []int{4, 5}, got)
}

// TestLazySliceMapReturnsEmptyOnEmptyInput verifies that Map on an empty source
// yields an empty lazy.
func TestLazySliceMapReturnsEmptyOnEmptyInput(t *testing.T) {
	t.Parallel()

	lazy := collection.NewLazySlice(slices.All([]int{}))

	got := lazy.Map(func(i int, v int) int { return v }).Items()

	require.Empty(t, got)
}

// TestLazySliceFirstWhereReturnsFirstMatch verifies that FirstWhere returns the
// first item satisfying the predicate with ok=true.
func TestLazySliceFirstWhereReturnsFirstMatch(t *testing.T) {
	t.Parallel()

	lazy := collection.NewLazySlice(slices.All([]int{1, 2, 3, 4}))

	result, ok := lazy.FirstWhere(func(i int, v int) bool { return v > 2 })

	require.True(t, ok)
	require.Equal(t, 3, result)
}

// TestLazySliceFirstWhereReturnsNotFoundWhenNoMatch verifies that FirstWhere
// returns ok=false when no item satisfies the predicate.
func TestLazySliceFirstWhereReturnsNotFoundWhenNoMatch(t *testing.T) {
	t.Parallel()

	lazy := collection.NewLazySlice(slices.All([]int{1, 2, 3}))

	_, ok := lazy.FirstWhere(func(i int, v int) bool { return v > 10 })

	require.False(t, ok)
}

// TestLazySliceFirstWhereReturnsNotFoundOnEmpty verifies that FirstWhere returns
// ok=false when the sequence is empty.
func TestLazySliceFirstWhereReturnsNotFoundOnEmpty(t *testing.T) {
	t.Parallel()

	lazy := collection.NewLazySlice(slices.All([]int{}))

	_, ok := lazy.FirstWhere(func(i int, v int) bool { return true })

	require.False(t, ok)
}

// TestLazySliceFirstReturnsFirstItemWithOkTrue verifies that First returns the
// first item and ok=true for a non-empty sequence.
func TestLazySliceFirstReturnsFirstItemWithOkTrue(t *testing.T) {
	t.Parallel()

	lazy := collection.NewLazySlice(slices.All([]int{42, 99}))

	result, ok := lazy.First()

	require.True(t, ok)
	require.Equal(t, 42, result)
}

// TestLazySliceFirstReturnsOkFalseOnEmpty verifies that First returns ok=false
// when the sequence is empty.
func TestLazySliceFirstReturnsOkFalseOnEmpty(t *testing.T) {
	t.Parallel()

	lazy := collection.NewLazySlice(slices.All([]int{}))

	_, ok := lazy.First()

	require.False(t, ok)
}

// TestLazySliceLastReturnsLastItemWithOkTrue verifies that Last returns the final
// item and ok=true for a non-empty sequence.
func TestLazySliceLastReturnsLastItemWithOkTrue(t *testing.T) {
	t.Parallel()

	lazy := collection.NewLazySlice(slices.All([]int{1, 2, 99}))

	result, ok := lazy.Last()

	require.True(t, ok)
	require.Equal(t, 99, result)
}

// TestLazySliceLastReturnsOkFalseOnEmpty verifies that Last returns ok=false when
// the sequence is empty.
func TestLazySliceLastReturnsOkFalseOnEmpty(t *testing.T) {
	t.Parallel()

	lazy := collection.NewLazySlice(slices.All([]int{}))

	_, ok := lazy.Last()

	require.False(t, ok)
}

// TestLazySliceContainsTrueWhenFound verifies that Contains returns true when an
// item satisfying the predicate exists.
func TestLazySliceContainsTrueWhenFound(t *testing.T) {
	t.Parallel()

	lazy := collection.NewLazySlice(slices.All([]int{1, 2, 3}))

	require.True(t, lazy.Contains(func(i int, v int) bool { return v == 2 }))
}

// TestLazySliceContainsFalseWhenNotFound verifies that Contains returns false when
// no item satisfies the predicate.
func TestLazySliceContainsFalseWhenNotFound(t *testing.T) {
	t.Parallel()

	lazy := collection.NewLazySlice(slices.All([]int{1, 2, 3}))

	require.False(t, lazy.Contains(func(i int, v int) bool { return v == 99 }))
}

// TestLazySliceTakeYieldsAtMostNItems verifies that Take limits output to n items.
func TestLazySliceTakeYieldsAtMostNItems(t *testing.T) {
	t.Parallel()

	lazy := collection.NewLazySlice(slices.All([]int{1, 2, 3, 4, 5}))

	require.Equal(t, []int{1, 2, 3}, lazy.Take(3).Items())
}

// TestLazySliceTakeZeroLimitYieldsNothing verifies that Take(0) yields an empty
// sequence.
func TestLazySliceTakeZeroLimitYieldsNothing(t *testing.T) {
	t.Parallel()

	lazy := collection.NewLazySlice(slices.All([]int{1, 2, 3}))

	require.Empty(t, lazy.Take(0).Items())
}

// TestLazySliceTakeNegativeLimitYieldsNothing verifies that Take with a negative
// limit yields an empty sequence.
func TestLazySliceTakeNegativeLimitYieldsNothing(t *testing.T) {
	t.Parallel()

	lazy := collection.NewLazySlice(slices.All([]int{1, 2, 3}))

	require.Empty(t, lazy.Take(-1).Items())
}

// TestLazySliceTakeLimitExceedingLengthYieldsAll verifies that Take yields all
// items when the limit exceeds the sequence length.
func TestLazySliceTakeLimitExceedingLengthYieldsAll(t *testing.T) {
	t.Parallel()

	lazy := collection.NewLazySlice(slices.All([]int{1, 2}))

	require.Equal(t, []int{1, 2}, lazy.Take(100).Items())
}

// TestLazySliceTakeWithInfiniteSourceYieldsExactlyNItems verifies that Take stops
// early when combined with an infinite generator, demonstrating lazy evaluation.
func TestLazySliceTakeWithInfiniteSourceYieldsExactlyNItems(t *testing.T) {
	t.Parallel()

	generated := 0
	infinite := collection.NewLazySlice(func(yield func(int, int) bool) {
		for i := 0; ; i++ {
			generated++
			if !yield(i, i) {
				return
			}
		}
	})

	got := infinite.Take(3).Items()

	require.Equal(t, []int{0, 1, 2}, got)
	require.Equal(t, 3, generated)
}

// TestLazySliceSkipSkipsFirstNItems verifies that Skip omits the first n items.
func TestLazySliceSkipSkipsFirstNItems(t *testing.T) {
	t.Parallel()

	lazy := collection.NewLazySlice(slices.All([]int{1, 2, 3, 4, 5}))

	require.Equal(t, []int{4, 5}, lazy.Skip(3).Items())
}

// TestLazySliceSkipZeroSkipsNothing verifies that Skip(0) returns all items.
func TestLazySliceSkipZeroSkipsNothing(t *testing.T) {
	t.Parallel()

	lazy := collection.NewLazySlice(slices.All([]int{1, 2, 3}))

	require.Equal(t, []int{1, 2, 3}, lazy.Skip(0).Items())
}

// TestLazySliceSkipExceedingLengthYieldsEmpty verifies that skipping more than
// the total count yields an empty sequence.
func TestLazySliceSkipExceedingLengthYieldsEmpty(t *testing.T) {
	t.Parallel()

	lazy := collection.NewLazySlice(slices.All([]int{1, 2}))

	require.Empty(t, lazy.Skip(10).Items())
}

// TestLazySliceTakeWhileYieldsItemsUntilPredicateReturnsFalse verifies that
// TakeWhile stops as soon as the predicate fails.
func TestLazySliceTakeWhileYieldsItemsUntilPredicateReturnsFalse(t *testing.T) {
	t.Parallel()

	lazy := collection.NewLazySlice(slices.All([]int{1, 2, 3, 4, 5}))

	got := lazy.TakeWhile(func(i int, v int) bool { return v < 4 }).Items()

	require.Equal(t, []int{1, 2, 3}, got)
}

// TestLazySliceTakeWhileYieldsNothingWhenFirstItemFailsPredicate verifies that
// TakeWhile yields nothing when the very first item fails.
func TestLazySliceTakeWhileYieldsNothingWhenFirstItemFailsPredicate(t *testing.T) {
	t.Parallel()

	lazy := collection.NewLazySlice(slices.All([]int{5, 1, 2}))

	got := lazy.TakeWhile(func(i int, v int) bool { return v < 4 }).Items()

	require.Empty(t, got)
}

// TestLazySliceTakeWhileYieldsAllWhenPredicateAlwaysTrue verifies that TakeWhile
// yields all items when the predicate never fails.
func TestLazySliceTakeWhileYieldsAllWhenPredicateAlwaysTrue(t *testing.T) {
	t.Parallel()

	lazy := collection.NewLazySlice(slices.All([]int{1, 2, 3}))

	got := lazy.TakeWhile(func(i int, v int) bool { return true }).Items()

	require.Equal(t, []int{1, 2, 3}, got)
}

// TestLazySliceSkipWhileSkipsLeadingItemsMatchingPredicateThenYieldsRest verifies
// that SkipWhile skips the leading run and returns the rest.
func TestLazySliceSkipWhileSkipsLeadingItemsMatchingPredicateThenYieldsRest(t *testing.T) {
	t.Parallel()

	lazy := collection.NewLazySlice(slices.All([]int{1, 2, 3, 4, 5}))

	got := lazy.SkipWhile(func(i int, v int) bool { return v < 4 }).Items()

	require.Equal(t, []int{4, 5}, got)
}

// TestLazySliceSkipWhileYieldsAllWhenFirstItemFailsPredicate verifies that
// SkipWhile yields all items when the first one fails the predicate.
func TestLazySliceSkipWhileYieldsAllWhenFirstItemFailsPredicate(t *testing.T) {
	t.Parallel()

	lazy := collection.NewLazySlice(slices.All([]int{5, 1, 2}))

	got := lazy.SkipWhile(func(i int, v int) bool { return v < 4 }).Items()

	require.Equal(t, []int{5, 1, 2}, got)
}

// TestLazySliceSkipWhileYieldsNothingWhenPredicateAlwaysTrue verifies that
// SkipWhile yields nothing when the predicate holds for every item.
func TestLazySliceSkipWhileYieldsNothingWhenPredicateAlwaysTrue(t *testing.T) {
	t.Parallel()

	lazy := collection.NewLazySlice(slices.All([]int{1, 2, 3}))

	got := lazy.SkipWhile(func(i int, v int) bool { return true }).Items()

	require.Empty(t, got)
}

// TestLazySliceChunkSplitsIntoChunksOfGivenSize verifies that Chunk partitions the
// sequence into consecutive slices, with the last chunk potentially smaller.
func TestLazySliceChunkSplitsIntoChunksOfGivenSize(t *testing.T) {
	t.Parallel()

	lazy := collection.NewLazySlice(slices.All([]int{1, 2, 3, 4, 5}))

	var chunks [][]int
	for _, chunk := range lazy.Chunk(2) {
		chunks = append(chunks, slices.Clone(chunk))
	}

	require.Equal(t, [][]int{{1, 2}, {3, 4}, {5}}, chunks)
}

// TestLazySliceChunkOnEmptyLazyYieldsNoChunks verifies that Chunk over an empty
// sequence produces no chunks.
func TestLazySliceChunkOnEmptyLazyYieldsNoChunks(t *testing.T) {
	t.Parallel()

	lazy := collection.NewLazySlice(slices.All([]int{}))

	var chunks [][]int
	for _, chunk := range lazy.Chunk(3) {
		chunks = append(chunks, slices.Clone(chunk))
	}

	require.Empty(t, chunks)
}

// TestLazySliceChunkPanicsOnZeroSize verifies that Chunk panics when size is zero.
func TestLazySliceChunkPanicsOnZeroSize(t *testing.T) {
	t.Parallel()

	lazy := collection.NewLazySlice(slices.All([]int{1, 2, 3}))

	require.Panics(t, func() { lazy.Chunk(0) })
}

// TestLazySliceChunkPanicsOnNegativeSize verifies that Chunk panics when size is
// negative.
func TestLazySliceChunkPanicsOnNegativeSize(t *testing.T) {
	t.Parallel()

	lazy := collection.NewLazySlice(slices.All([]int{1, 2, 3}))

	require.Panics(t, func() { lazy.Chunk(-1) })
}

// TestLazySliceSlidingReturnsWindowsWithDefaultStepOfOne verifies that Sliding
// with no step argument advances by one between windows.
func TestLazySliceSlidingReturnsWindowsWithDefaultStepOfOne(t *testing.T) {
	t.Parallel()

	lazy := collection.NewLazySlice(slices.All([]int{1, 2, 3, 4}))

	var windows [][]int
	for _, w := range lazy.Sliding(2) {
		windows = append(windows, slices.Clone(w))
	}

	require.Equal(t, [][]int{{1, 2}, {2, 3}, {3, 4}}, windows)
}

// TestLazySliceSlidingReturnsWindowsWithExplicitStep verifies that Sliding with an
// explicit step advances by that many positions between windows.
func TestLazySliceSlidingReturnsWindowsWithExplicitStep(t *testing.T) {
	t.Parallel()

	lazy := collection.NewLazySlice(slices.All([]int{1, 2, 3, 4, 5}))

	var windows [][]int
	for _, w := range lazy.Sliding(3, 2) {
		windows = append(windows, slices.Clone(w))
	}

	require.Equal(t, [][]int{{1, 2, 3}, {3, 4, 5}}, windows)
}

// TestLazySliceSlidingReturnsNothingWhenCollectionSmallerThanWindowSize verifies
// that Sliding yields nothing when the source has fewer items than the window.
func TestLazySliceSlidingReturnsNothingWhenCollectionSmallerThanWindowSize(t *testing.T) {
	t.Parallel()

	lazy := collection.NewLazySlice(slices.All([]int{1, 2}))

	var windows [][]int
	for _, w := range lazy.Sliding(5) {
		windows = append(windows, slices.Clone(w))
	}

	require.Empty(t, windows)
}

// TestLazySliceSlidingPanicsOnZeroSize verifies that Sliding panics when size is
// zero.
func TestLazySliceSlidingPanicsOnZeroSize(t *testing.T) {
	t.Parallel()

	lazy := collection.NewLazySlice(slices.All([]int{1, 2, 3}))

	require.Panics(t, func() { lazy.Sliding(0) })
}

// TestLazySliceSlidingPanicsOnNegativeSize verifies that Sliding panics when size
// is negative.
func TestLazySliceSlidingPanicsOnNegativeSize(t *testing.T) {
	t.Parallel()

	lazy := collection.NewLazySlice(slices.All([]int{1, 2, 3}))

	require.Panics(t, func() { lazy.Sliding(-1) })
}

// TestLazySliceSlidingPanicsOnZeroStep verifies that Sliding panics when step is
// zero.
func TestLazySliceSlidingPanicsOnZeroStep(t *testing.T) {
	t.Parallel()

	lazy := collection.NewLazySlice(slices.All([]int{1, 2, 3}))

	require.Panics(t, func() { lazy.Sliding(2, 0) })
}

// TestLazySliceSlidingPanicsOnNegativeStep verifies that Sliding panics when step
// is negative.
func TestLazySliceSlidingPanicsOnNegativeStep(t *testing.T) {
	t.Parallel()

	lazy := collection.NewLazySlice(slices.All([]int{1, 2, 3}))

	require.Panics(t, func() { lazy.Sliding(2, -1) })
}

// TestLazySliceReduceAccumulatesValuesWithSameType verifies that Reduce computes
// the correct sum over a sequence of integers.
func TestLazySliceReduceAccumulatesValuesWithSameType(t *testing.T) {
	t.Parallel()

	lazy := collection.NewLazySlice(slices.All([]int{1, 2, 3, 4}))

	sum := lazy.Reduce(func(acc int, i int, v int) int { return acc + v }, 0)

	require.Equal(t, 10, sum)
}

// TestLazySliceReduceAccumulatesWithDifferentKType verifies that Reduce can use a
// different accumulator type than the element type (int32 → int64).
func TestLazySliceReduceAccumulatesWithDifferentKType(t *testing.T) {
	t.Parallel()

	lazy := collection.NewLazySlice(slices.All([]int32{1, 2, 3}))

	sum := lazy.Reduce(func(acc int64, i int, v int32) int64 { return acc + int64(v) }, int64(0))

	require.Equal(t, int64(6), sum)
}

// TestLazySliceReduceReturnsInitialOnEmpty verifies that Reduce returns the initial
// value unchanged when the sequence is empty.
func TestLazySliceReduceReturnsInitialOnEmpty(t *testing.T) {
	t.Parallel()

	lazy := collection.NewLazySlice(slices.All([]int{}))

	result := lazy.Reduce(func(acc int, i int, v int) int { return acc + v }, 42)

	require.Equal(t, 42, result)
}

// TestLazySliceReverseReturnsItemsInReversedOrder verifies that Reverse returns a
// lazy that yields items in reverse order.
func TestLazySliceReverseReturnsItemsInReversedOrder(t *testing.T) {
	t.Parallel()

	lazy := collection.NewLazySlice(slices.All([]int{1, 2, 3}))

	require.Equal(t, []int{3, 2, 1}, lazy.Reverse().Items())
}

// TestLazySliceReverseReturnsEmptyOnEmptyInput verifies that Reverse on an empty
// source yields an empty sequence.
func TestLazySliceReverseReturnsEmptyOnEmptyInput(t *testing.T) {
	t.Parallel()

	lazy := collection.NewLazySlice(slices.All([]int{}))

	require.Empty(t, lazy.Reverse().Items())
}

// TestLazySliceSortReturnsItemsInSortedOrder verifies that Sort orders items using
// the provided comparison function.
func TestLazySliceSortReturnsItemsInSortedOrder(t *testing.T) {
	t.Parallel()

	lazy := collection.NewLazySlice(slices.All([]int{3, 1, 4, 1, 5, 9}))

	got := lazy.Sort(cmp.Compare[int]).Items()

	require.Equal(t, []int{1, 1, 3, 4, 5, 9}, got)
}

// TestLazySliceUniqueRemovesDuplicatesAndPreservesFirstOccurrence verifies that
// Unique keeps only the first occurrence of each distinct key.
func TestLazySliceUniqueRemovesDuplicatesAndPreservesFirstOccurrence(t *testing.T) {
	t.Parallel()

	lazy := collection.NewLazySlice(slices.All([]int{1, 2, 2, 3, 1, 4}))

	got := lazy.Unique(func(i int, v int) int { return v }).Items()

	require.Equal(t, []int{1, 2, 3, 4}, got)
}

// TestLazySlicePartitionSplitsIntoMatchingAndNonMatching verifies that Partition
// correctly separates items into two groups.
func TestLazySlicePartitionSplitsIntoMatchingAndNonMatching(t *testing.T) {
	t.Parallel()

	lazy := collection.NewLazySlice(slices.All([]int{1, 2, 3, 4, 5}))

	even, odd := lazy.Partition(func(i int, v int) bool { return v%2 == 0 })

	require.Equal(t, []int{2, 4}, even.Items())
	require.Equal(t, []int{1, 3, 5}, odd.Items())
}

// TestLazySlicePartitionAllMatchSecondResultYieldsNothing verifies that when every
// item matches, the second (non-matching) lazy is empty.
func TestLazySlicePartitionAllMatchSecondResultYieldsNothing(t *testing.T) {
	t.Parallel()

	lazy := collection.NewLazySlice(slices.All([]int{2, 4, 6}))

	_, odd := lazy.Partition(func(i int, v int) bool { return v%2 == 0 })

	require.Empty(t, odd.Items())
}

// TestLazySlicePartitionNoneMatchFirstResultYieldsNothing verifies that when no
// item matches, the first (matching) lazy is empty.
func TestLazySlicePartitionNoneMatchFirstResultYieldsNothing(t *testing.T) {
	t.Parallel()

	lazy := collection.NewLazySlice(slices.All([]int{1, 3, 5}))

	even, _ := lazy.Partition(func(i int, v int) bool { return v%2 == 0 })

	require.Empty(t, even.Items())
}

// TestLazySliceConcatYieldsAllItemsFromBothSources verifies that Concat appends the
// second lazy's items after the first's.
func TestLazySliceConcatYieldsAllItemsFromBothSources(t *testing.T) {
	t.Parallel()

	first := collection.NewLazySlice(slices.All([]int{1, 2, 3}))
	second := collection.NewLazySlice(slices.All([]int{4, 5, 6}))

	got := first.Concat(second).Items()

	require.Equal(t, []int{1, 2, 3, 4, 5, 6}, got)
}

// TestLazySliceGroupByGroupsAllItemsByKeyIntoEagerSubCollections verifies that
// GroupBy partitions items into map entries of eager Slices.
func TestLazySliceGroupByGroupsAllItemsByKeyIntoEagerSubCollections(t *testing.T) {
	t.Parallel()

	lazy := collection.NewLazySlice(slices.All([]int{1, 2, 3, 4, 5, 6}))

	groups := lazy.GroupBy(func(i int, v int) string {
		if v%2 == 0 {
			return "even"
		}
		return "odd"
	})

	require.Equal(t, 2, groups.Len())

	odd, ok := groups.Get("odd")
	require.True(t, ok)
	require.Equal(t, []int{1, 3, 5}, odd.Items())

	even, ok := groups.Get("even")
	require.True(t, ok)
	require.Equal(t, []int{2, 4, 6}, even.Items())
}

// TestLazySliceNthPicksEveryNthItemFromIndexZero verifies that Nth selects every
// nth item starting from index 0.
func TestLazySliceNthPicksEveryNthItemFromIndexZero(t *testing.T) {
	t.Parallel()

	lazy := collection.NewLazySlice(slices.All([]int{0, 1, 2, 3, 4, 5, 6}))

	got := lazy.Nth(3).Items()

	require.Equal(t, []int{0, 3, 6}, got)
}

// TestLazySliceNthPicksEveryNthItemStartingAtGivenOffset verifies that Nth with an
// offset begins selection from that position.
func TestLazySliceNthPicksEveryNthItemStartingAtGivenOffset(t *testing.T) {
	t.Parallel()

	lazy := collection.NewLazySlice(slices.All([]int{0, 1, 2, 3, 4, 5, 6}))

	got := lazy.Nth(3, 1).Items()

	require.Equal(t, []int{1, 4}, got)
}

// TestLazySliceNthOffsetBeyondLengthReturnsEmptyLazySlice verifies that an offset
// larger than the sequence length yields nothing.
func TestLazySliceNthOffsetBeyondLengthReturnsEmptyLazySlice(t *testing.T) {
	t.Parallel()

	lazy := collection.NewLazySlice(slices.All([]int{1, 2, 3}))

	got := lazy.Nth(2, 100).Items()

	require.Empty(t, got)
}

// TestLazySliceNthPanicsOnZeroN verifies that Nth panics when n is zero.
func TestLazySliceNthPanicsOnZeroN(t *testing.T) {
	t.Parallel()

	lazy := collection.NewLazySlice(slices.All([]int{1, 2, 3}))

	require.Panics(t, func() { lazy.Nth(0) })
}

// TestLazySliceNthPanicsOnNegativeN verifies that Nth panics when n is negative.
func TestLazySliceNthPanicsOnNegativeN(t *testing.T) {
	t.Parallel()

	lazy := collection.NewLazySlice(slices.All([]int{1, 2, 3}))

	require.Panics(t, func() { lazy.Nth(-1) })
}

// TestLazySliceForPageReturnsFirstPage verifies that ForPage returns the first
// page of items.
func TestLazySliceForPageReturnsFirstPage(t *testing.T) {
	t.Parallel()

	lazy := collection.NewLazySlice(slices.All([]int{1, 2, 3, 4, 5, 6}))

	got := lazy.ForPage(1, 2).Items()

	require.Equal(t, []int{1, 2}, got)
}

// TestLazySliceForPageReturnsSecondPage verifies that ForPage returns the second
// page of items.
func TestLazySliceForPageReturnsSecondPage(t *testing.T) {
	t.Parallel()

	lazy := collection.NewLazySlice(slices.All([]int{1, 2, 3, 4, 5, 6}))

	got := lazy.ForPage(2, 2).Items()

	require.Equal(t, []int{3, 4}, got)
}

// TestLazySliceForPageBeyondDataReturnsEmptyLazySlice verifies that ForPage returns
// nothing when the requested page exceeds the available data.
func TestLazySliceForPageBeyondDataReturnsEmptyLazySlice(t *testing.T) {
	t.Parallel()

	lazy := collection.NewLazySlice(slices.All([]int{1, 2, 3}))

	got := lazy.ForPage(10, 2).Items()

	require.Empty(t, got)
}

// TestLazySliceMarshalJSONMarshalsNonEmptyLazyToJSONArray verifies that MarshalJSON
// encodes a non-empty lazy as a JSON array.
func TestLazySliceMarshalJSONMarshalsNonEmptyLazyToJSONArray(t *testing.T) {
	t.Parallel()

	lazy := collection.NewLazySlice(slices.All([]int{1, 2, 3}))

	data, err := lazy.MarshalJSON()

	require.NoError(t, err)
	require.JSONEq(t, `[1,2,3]`, string(data))
}

// TestLazySliceMarshalJSONMarshalsEmptyLazyToNull verifies that MarshalJSON on an
// empty lazy produces a null or empty-array JSON value with no error.
func TestLazySliceMarshalJSONMarshalsEmptyLazyToNull(t *testing.T) {
	t.Parallel()

	lazy := collection.NewLazySlice(slices.All([]int{}))

	data, err := lazy.MarshalJSON()

	require.NoError(t, err)

	// Round-trip: the encoded output must be valid JSON.
	var decoded []int
	require.NoError(t, json.Unmarshal(data, &decoded))
}

// TestLazySliceUnmarshalJSONDecodesJSONArrayAndYieldsItems verifies that
// UnmarshalJSON populates the lazy so that it yields the decoded items.
func TestLazySliceUnmarshalJSONDecodesJSONArrayAndYieldsItems(t *testing.T) {
	t.Parallel()

	var lazy collection.LazySlice[int]

	require.NoError(t, lazy.UnmarshalJSON([]byte(`[10,20,30]`)))
	require.Equal(t, []int{10, 20, 30}, lazy.Items())
}

// TestLazySliceUnmarshalJSONReturnsErrorOnInvalidJSON verifies that UnmarshalJSON
// returns an error when given malformed JSON.
func TestLazySliceUnmarshalJSONReturnsErrorOnInvalidJSON(t *testing.T) {
	t.Parallel()

	var lazy collection.LazySlice[int]

	require.Error(t, lazy.UnmarshalJSON([]byte(`not-json`)))
}

// TestLazySliceMarshalJSONToEncodesNonEmptyLazyToJSONArray verifies that
// MarshalJSONTo writes a valid JSON array to the encoder.
func TestLazySliceMarshalJSONToEncodesNonEmptyLazyToJSONArray(t *testing.T) {
	t.Parallel()

	lazy := collection.NewLazySlice(slices.All([]int{4, 5, 6}))

	var buf bytes.Buffer
	enc := jsontext.NewEncoder(&buf)

	require.NoError(t, lazy.MarshalJSONTo(enc))
	require.JSONEq(t, `[4,5,6]`, buf.String())
}

// TestLazySliceUnmarshalJSONFromDecodesJSONArrayAndYieldsItems verifies that
// UnmarshalJSONFrom populates the lazy from a JSON array.
func TestLazySliceUnmarshalJSONFromDecodesJSONArrayAndYieldsItems(t *testing.T) {
	t.Parallel()

	dec := jsontext.NewDecoder(strings.NewReader(`[7,8,9]`))

	var lazy collection.LazySlice[int]

	require.NoError(t, lazy.UnmarshalJSONFrom(dec))
	require.Equal(t, []int{7, 8, 9}, lazy.Items())
}

// TestLazySliceUnmarshalJSONFromReturnsErrorOnInvalidJSON verifies that
// UnmarshalJSONFrom returns an error when the decoder contains invalid JSON.
func TestLazySliceUnmarshalJSONFromReturnsErrorOnInvalidJSON(t *testing.T) {
	t.Parallel()

	dec := jsontext.NewDecoder(strings.NewReader(`not-json`))

	var lazy collection.LazySlice[int]

	require.Error(t, lazy.UnmarshalJSONFrom(dec))
}
