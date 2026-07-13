package collection_test

import (
	"bytes"
	"cmp"
	"encoding/json"
	"encoding/json/jsontext"
	"strings"
	"testing"

	"github.com/studiolambda/cosmos/collection"

	"github.com/stretchr/testify/require"
)

// TestNewSliceCreatesSliceFromSlice verifies that NewSlice stores the given items.
func TestNewSliceCreatesSliceFromSlice(t *testing.T) {
	t.Parallel()

	c := collection.NewSlice([]int{1, 2, 3})

	require.Equal(t, []int{1, 2, 3}, c.Items())
}

// TestNewSliceAsAcceptsNamedSliceType verifies that NewSliceAs accepts named
// slice types whose underlying type is []T.
func TestNewSliceAsAcceptsNamedSliceType(t *testing.T) {
	t.Parallel()

	type numbers []int

	c := collection.NewSliceAs(numbers{1, 2, 3})

	require.Equal(t, []int{1, 2, 3}, c.Items())
}

// TestNewSliceCreatesSliceFromPlainSlice verifies that NewSlice stores the
// given plain slice.
func TestNewSliceCreatesSliceFromPlainSlice(t *testing.T) {
	t.Parallel()

	c := collection.NewSlice([]int{1, 2, 3})

	require.Equal(t, []int{1, 2, 3}, c.Items())
}

// TestNewSliceCreatesEmptySliceFromEmptySlice verifies that NewSlice with an empty
// (non-nil) slice produces an empty collection.
func TestNewSliceCreatesEmptySliceFromEmptySlice(t *testing.T) {
	t.Parallel()

	c := collection.NewSlice([]int{})

	require.True(t, c.IsEmpty())
}

// TestNewSliceCreatesEmptySliceFromNilSlice verifies that NewSlice with a nil slice
// produces an empty collection.
func TestNewSliceCreatesEmptySliceFromNilSlice(t *testing.T) {
	t.Parallel()

	c := collection.NewSlice[int](nil)

	require.True(t, c.IsEmpty())
}

// TestIterYieldsAllItemsInOrder verifies that Iter produces every item in
// insertion order.
func TestIterYieldsAllItemsInOrder(t *testing.T) {
	t.Parallel()

	c := collection.NewSlice([]int{10, 20, 30})

	var got []int
	for _, v := range c.Iter() {
		got = append(got, v)
	}

	require.Equal(t, []int{10, 20, 30}, got)
}

// TestIterYieldsNothingOnEmptyCollection verifies that Iter over an empty
// collection never calls the yield function.
func TestIterYieldsNothingOnEmptyCollection(t *testing.T) {
	t.Parallel()

	c := collection.NewSlice[int](nil)

	var got []int
	for _, v := range c.Iter() {
		got = append(got, v)
	}

	require.Empty(t, got)
}

// TestLazyContainsSameItemsAsCollection verifies that the LazySlice returned by
// Lazy() materializes the same items as the original collection.
func TestLazyContainsSameItemsAsCollection(t *testing.T) {
	t.Parallel()

	c := collection.NewSlice([]string{"a", "b", "c"})

	require.Equal(t, c.Items(), c.Lazy().Items())
}

// TestSliceReturnsUnderlyingItems verifies that Items returns the items passed
// to NewSlice.
func TestSliceReturnsUnderlyingItems(t *testing.T) {
	t.Parallel()

	items := []int{7, 8, 9}
	c := collection.NewSlice(items)

	require.Equal(t, items, c.Items())
}

// TestIsEmptyTrueForEmptyCollection verifies that IsEmpty returns true when the
// collection has no items.
func TestIsEmptyTrueForEmptyCollection(t *testing.T) {
	t.Parallel()

	c := collection.NewSlice[int](nil)

	require.True(t, c.IsEmpty())
}

// TestIsEmptyFalseForNonEmptyCollection verifies that IsEmpty returns false when
// the collection has items.
func TestIsEmptyFalseForNonEmptyCollection(t *testing.T) {
	t.Parallel()

	c := collection.NewSlice([]int{1})

	require.False(t, c.IsEmpty())
}

// TestIsNotEmptyTrueForNonEmptyCollection verifies that IsNotEmpty returns true
// when the collection has items.
func TestIsNotEmptyTrueForNonEmptyCollection(t *testing.T) {
	t.Parallel()

	c := collection.NewSlice([]int{1, 2})

	require.True(t, c.IsNotEmpty())
}

// TestIsNotEmptyFalseForEmptyCollection verifies that IsNotEmpty returns false
// when the collection has no items.
func TestIsNotEmptyFalseForEmptyCollection(t *testing.T) {
	t.Parallel()

	c := collection.NewSlice[int](nil)

	require.False(t, c.IsNotEmpty())
}

// TestLenReturnsItemCount verifies that Len returns the number of items.
func TestLenReturnsItemCount(t *testing.T) {
	t.Parallel()

	c := collection.NewSlice([]int{1, 2, 3, 4})

	require.Equal(t, 4, c.Len())
}

// TestLenReturnsZeroForEmptyCollection verifies that Len returns zero for an
// empty collection.
func TestLenReturnsZeroForEmptyCollection(t *testing.T) {
	t.Parallel()

	c := collection.NewSlice[int](nil)

	require.Equal(t, 0, c.Len())
}

// TestEveryTrueWhenAllItemsMatch verifies that Every returns true when the
// predicate holds for all items.
func TestEveryTrueWhenAllItemsMatch(t *testing.T) {
	t.Parallel()

	c := collection.NewSlice([]int{2, 4, 6})

	require.True(t, c.Every(func(i int, v int) bool { return v%2 == 0 }))
}

// TestEveryFalseWhenOneItemFails verifies that Every returns false when at
// least one item does not satisfy the predicate.
func TestEveryFalseWhenOneItemFails(t *testing.T) {
	t.Parallel()

	c := collection.NewSlice([]int{2, 3, 6})

	require.False(t, c.Every(func(i int, v int) bool { return v%2 == 0 }))
}

// TestEveryTrueOnEmptyCollection verifies the vacuous truth: Every returns true
// when there are no items.
func TestEveryTrueOnEmptyCollection(t *testing.T) {
	t.Parallel()

	c := collection.NewSlice[int](nil)

	require.True(t, c.Every(func(i int, v int) bool { return false }))
}

// TestEachCallsFForEveryItem verifies that Each visits every item in order.
func TestEachCallsFForEveryItem(t *testing.T) {
	t.Parallel()

	c := collection.NewSlice([]int{1, 2, 3})

	var got []int
	c.Each(func(i int, v int) { got = append(got, v) })

	require.Equal(t, []int{1, 2, 3}, got)
}

// TestEachDoesNotCallFOnEmpty verifies that Each never invokes f on an empty
// collection.
func TestEachDoesNotCallFOnEmpty(t *testing.T) {
	t.Parallel()

	c := collection.NewSlice[int](nil)

	called := false
	c.Each(func(int, int) { called = true })

	require.False(t, called)
}

// TestTapEachCallsFForEveryItem verifies that TapEach visits every item.
func TestTapEachCallsFForEveryItem(t *testing.T) {
	t.Parallel()

	c := collection.NewSlice([]int{4, 5, 6})

	var got []int
	c.TapEach(func(i int, v int) { got = append(got, v) })

	require.Equal(t, []int{4, 5, 6}, got)
}

// TestTapEachReturnsOriginalCollectionUnchanged verifies that TapEach returns
// the same items as the receiver.
func TestTapEachReturnsOriginalCollectionUnchanged(t *testing.T) {
	t.Parallel()

	c := collection.NewSlice([]int{4, 5, 6})

	result := c.TapEach(func(int, int) {})

	require.Equal(t, c.Items(), result.Items())
}

// TestFilterReturnsMatchingItems verifies that Filter keeps only items for
// which the predicate is true.
func TestFilterReturnsMatchingItems(t *testing.T) {
	t.Parallel()

	c := collection.NewSlice([]int{1, 2, 3, 4, 5})

	result := c.Filter(func(i int, v int) bool { return v%2 == 0 })

	require.Equal(t, []int{2, 4}, result.Items())
}

// TestFilterReturnsEmptyWhenNoneMatch verifies that Filter produces an empty
// collection when no items satisfy the predicate.
func TestFilterReturnsEmptyWhenNoneMatch(t *testing.T) {
	t.Parallel()

	c := collection.NewSlice([]int{1, 3, 5})

	result := c.Filter(func(i int, v int) bool { return v%2 == 0 })

	require.True(t, result.IsEmpty())
}

// TestFilterReturnsEmptyOnEmptyInput verifies that Filter on an empty
// collection yields an empty collection.
func TestFilterReturnsEmptyOnEmptyInput(t *testing.T) {
	t.Parallel()

	c := collection.NewSlice[int](nil)

	result := c.Filter(func(i int, v int) bool { return true })

	require.True(t, result.IsEmpty())
}

// TestFlatMapFlattensMappedResults verifies that FlatMap maps each item to many
// items and flattens the results.
func TestFlatMapFlattensMappedResults(t *testing.T) {
	t.Parallel()

	c := collection.NewSlice([]int{1, 2, 3})

	result := c.FlatMap(func(i int, v int) []int {
		return []int{v, v * 10}
	})

	require.Equal(t, []int{1, 10, 2, 20, 3, 30}, result.Items())
}

// TestFlatMapCanDropItems verifies that FlatMap supports zero-result mappings.
func TestFlatMapCanDropItems(t *testing.T) {
	t.Parallel()

	c := collection.NewSlice([]int{1, 2, 3, 4})

	result := c.FlatMap(func(i int, v int) []int {
		if v%2 == 0 {
			return nil
		}

		return []int{v}
	})

	require.Equal(t, []int{1, 3}, result.Items())
}

// TestKeyByIndexesItemsByDerivedKey verifies that KeyBy builds a lookup map from a slice.
func TestKeyByIndexesItemsByDerivedKey(t *testing.T) {
	t.Parallel()

	type user struct {
		ID   int
		Name string
	}

	c := collection.NewSlice([]user{{ID: 1, Name: "alice"}, {ID: 2, Name: "bob"}})

	result := c.KeyBy(func(i int, v user) int { return v.ID })

	require.Equal(t, map[int]user{
		1: {ID: 1, Name: "alice"},
		2: {ID: 2, Name: "bob"},
	}, result.Items())
}

// TestKeyByLastItemWinsOnCollision verifies that KeyBy keeps the last item for duplicate keys.
func TestKeyByLastItemWinsOnCollision(t *testing.T) {
	t.Parallel()

	c := collection.NewSlice([]string{"ant", "ape", "bat"})

	result := c.KeyBy(func(i int, v string) byte { return v[0] })

	require.Equal(t, map[byte]string{'a': "ape", 'b': "bat"}, result.Items())
}

// TestCountByCountsItemsPerDerivedKey verifies that CountBy aggregates counts by key.
func TestCountByCountsItemsPerDerivedKey(t *testing.T) {
	t.Parallel()

	c := collection.NewSlice([]string{"ant", "ape", "bat", "bee"})

	result := c.CountBy(func(i int, v string) byte { return v[0] })

	require.Equal(t, map[byte]int{'a': 2, 'b': 2}, result.Items())
}

// TestTakeUntilStopsBeforeMatchingItem verifies that TakeUntil excludes the first matching item.
func TestTakeUntilStopsBeforeMatchingItem(t *testing.T) {
	t.Parallel()

	c := collection.NewSlice([]int{1, 2, 3, 4, 5})

	result := c.TakeUntil(func(i int, v int) bool { return v >= 4 })

	require.Equal(t, []int{1, 2, 3}, result.Items())
}

// TestTakeUntilReturnsAllWhenNeverMatched verifies that TakeUntil returns the full slice when unmatched.
func TestTakeUntilReturnsAllWhenNeverMatched(t *testing.T) {
	t.Parallel()

	c := collection.NewSlice([]int{1, 2, 3})

	result := c.TakeUntil(func(i int, v int) bool { return v > 10 })

	require.Equal(t, []int{1, 2, 3}, result.Items())
}

// TestSkipUntilStartsAtFirstMatchingItem verifies that SkipUntil keeps items from the first match onward.
func TestSkipUntilStartsAtFirstMatchingItem(t *testing.T) {
	t.Parallel()

	c := collection.NewSlice([]int{1, 2, 3, 4, 5})

	result := c.SkipUntil(func(i int, v int) bool { return v >= 4 })

	require.Equal(t, []int{4, 5}, result.Items())
}

// TestSkipUntilReturnsEmptyWhenNeverMatched verifies that SkipUntil returns empty when no item matches.
func TestSkipUntilReturnsEmptyWhenNeverMatched(t *testing.T) {
	t.Parallel()

	c := collection.NewSlice([]int{1, 2, 3})

	result := c.SkipUntil(func(i int, v int) bool { return v > 10 })

	require.True(t, result.IsEmpty())
}

// TestRejectReturnsOnlyNonMatchingItems verifies that Reject keeps only items
// for which the predicate is false.
func TestRejectReturnsOnlyNonMatchingItems(t *testing.T) {
	t.Parallel()

	c := collection.NewSlice([]int{1, 2, 3, 4, 5})

	result := c.Reject(func(i int, v int) bool { return v%2 == 0 })

	require.Equal(t, []int{1, 3, 5}, result.Items())
}

// TestMapTransformsItemsToSameType verifies that Map applies f to every item
// and returns a collection of the same type.
func TestMapTransformsItemsToSameType(t *testing.T) {
	t.Parallel()

	c := collection.NewSlice([]int{1, 2, 3})

	result := collection.Slice[int].Map(c, func(i int, v int) int { return v * 2 })

	require.Equal(t, []int{2, 4, 6}, result.Items())
}

// TestMapTransformsItemsToDifferentType verifies that Map can change the
// element type (int → bool).
func TestMapTransformsItemsToDifferentType(t *testing.T) {
	t.Parallel()

	c := collection.NewSlice([]int{1, 2, 3, 4})

	result := collection.Slice[int].Map(c, func(i int, v int) bool { return v%2 != 0 })

	require.Equal(t, []bool{true, false, true, false}, result.Items())
}

// TestMapReturnsEmptyCollectionOnEmptyInput verifies that Map on an empty
// collection produces an empty collection.
func TestMapReturnsEmptyCollectionOnEmptyInput(t *testing.T) {
	t.Parallel()

	c := collection.NewSlice[int](nil)

	result := collection.Slice[int].Map(c, func(i int, v int) int { return v })

	require.True(t, result.IsEmpty())
}

// TestFirstWhereReturnsFirstMatchingItem verifies that FirstWhere returns the
// first item satisfying f with ok=true.
func TestFirstWhereReturnsFirstMatchingItem(t *testing.T) {
	t.Parallel()

	c := collection.NewSlice([]int{1, 2, 3, 4})

	result, ok := c.FirstWhere(func(i int, v int) bool { return v > 2 })

	require.True(t, ok)
	require.Equal(t, 3, result)
}

// TestFirstWhereReturnsOkFalseWhenNoMatch verifies that FirstWhere returns
// ok=false when no item satisfies f.
func TestFirstWhereReturnsOkFalseWhenNoMatch(t *testing.T) {
	t.Parallel()

	c := collection.NewSlice([]int{1, 2, 3})

	_, ok := c.FirstWhere(func(i int, v int) bool { return v > 100 })

	require.False(t, ok)
}

// TestFirstWhereReturnsOkFalseOnEmpty verifies that FirstWhere on an empty
// collection returns ok=false.
func TestFirstWhereReturnsOkFalseOnEmpty(t *testing.T) {
	t.Parallel()

	c := collection.NewSlice[int](nil)

	_, ok := c.FirstWhere(func(i int, v int) bool { return true })

	require.False(t, ok)
}

// TestFirstReturnsFirstItemWithOkTrue verifies that First returns the leading
// item and ok=true for a non-empty collection.
func TestFirstReturnsFirstItemWithOkTrue(t *testing.T) {
	t.Parallel()

	c := collection.NewSlice([]int{42, 99})

	result, ok := c.First()

	require.True(t, ok)
	require.Equal(t, 42, result)
}

// TestFirstReturnsOkFalseOnEmpty verifies that First returns ok=false on an
// empty collection.
func TestFirstReturnsOkFalseOnEmpty(t *testing.T) {
	t.Parallel()

	c := collection.NewSlice[int](nil)

	_, ok := c.First()

	require.False(t, ok)
}

// TestLastReturnsLastItemWithOkTrue verifies that Last returns the final item
// and ok=true for a non-empty collection.
func TestLastReturnsLastItemWithOkTrue(t *testing.T) {
	t.Parallel()

	c := collection.NewSlice([]int{1, 2, 77})

	result, ok := c.Last()

	require.True(t, ok)
	require.Equal(t, 77, result)
}

// TestLastReturnsOkFalseOnEmpty verifies that Last returns ok=false on an
// empty collection.
func TestLastReturnsOkFalseOnEmpty(t *testing.T) {
	t.Parallel()

	c := collection.NewSlice[int](nil)

	_, ok := c.Last()

	require.False(t, ok)
}

// TestContainsTrueWhenFound verifies that Contains returns true when at least
// one item satisfies the predicate.
func TestContainsTrueWhenFound(t *testing.T) {
	t.Parallel()

	c := collection.NewSlice([]int{1, 2, 3})

	require.True(t, c.Contains(func(i int, v int) bool { return v == 2 }))
}

// TestContainsFalseWhenNotFound verifies that Contains returns false when no
// item satisfies the predicate.
func TestContainsFalseWhenNotFound(t *testing.T) {
	t.Parallel()

	c := collection.NewSlice([]int{1, 2, 3})

	require.False(t, c.Contains(func(i int, v int) bool { return v == 99 }))
}

// TestTakeReturnsFirstNItems verifies that Take returns exactly the first n
// items.
func TestTakeReturnsFirstNItems(t *testing.T) {
	t.Parallel()

	c := collection.NewSlice([]int{1, 2, 3, 4, 5})

	require.Equal(t, []int{1, 2, 3}, c.Take(3).Items())
}

// TestTakeWithZeroLimitReturnsEmpty verifies that Take(0) produces an empty
// collection.
func TestTakeWithZeroLimitReturnsEmpty(t *testing.T) {
	t.Parallel()

	c := collection.NewSlice([]int{1, 2, 3})

	require.True(t, c.Take(0).IsEmpty())
}

// TestTakeWithNegativeLimitReturnsEmpty verifies that Take with a negative
// limit produces an empty collection.
func TestTakeWithNegativeLimitReturnsEmpty(t *testing.T) {
	t.Parallel()

	c := collection.NewSlice([]int{1, 2, 3})

	require.True(t, c.Take(-5).IsEmpty())
}

// TestTakeWithLimitExceedingLengthReturnsAll verifies that Take with a limit
// larger than the collection length returns all items.
func TestTakeWithLimitExceedingLengthReturnsAll(t *testing.T) {
	t.Parallel()

	c := collection.NewSlice([]int{1, 2, 3})

	require.Equal(t, []int{1, 2, 3}, c.Take(100).Items())
}

// TestSkipRemovesFirstNItems verifies that Skip(n) removes exactly the first n
// items.
func TestSkipRemovesFirstNItems(t *testing.T) {
	t.Parallel()

	c := collection.NewSlice([]int{1, 2, 3, 4, 5})

	require.Equal(t, []int{3, 4, 5}, c.Skip(2).Items())
}

// TestSkipZeroRemovesNothing verifies that Skip(0) returns all items.
func TestSkipZeroRemovesNothing(t *testing.T) {
	t.Parallel()

	c := collection.NewSlice([]int{1, 2, 3})

	require.Equal(t, []int{1, 2, 3}, c.Skip(0).Items())
}

// TestSkipNegativeRemovesNothing verifies that Skip with a negative value
// returns all items.
func TestSkipNegativeRemovesNothing(t *testing.T) {
	t.Parallel()

	c := collection.NewSlice([]int{1, 2, 3})

	require.Equal(t, []int{1, 2, 3}, c.Skip(-3).Items())
}

// TestSkipNExceedingLengthReturnsEmpty verifies that Skip with n >= length
// produces an empty collection.
func TestSkipNExceedingLengthReturnsEmpty(t *testing.T) {
	t.Parallel()

	c := collection.NewSlice([]int{1, 2, 3})

	require.True(t, c.Skip(10).IsEmpty())
}

// TestTakeWhileReturnsItemsUntilFirstFalse verifies that TakeWhile stops at
// the first item that fails the predicate.
func TestTakeWhileReturnsItemsUntilFirstFalse(t *testing.T) {
	t.Parallel()

	c := collection.NewSlice([]int{2, 4, 5, 6})

	require.Equal(t, []int{2, 4}, c.TakeWhile(func(i int, v int) bool { return v%2 == 0 }).Items())
}

// TestTakeWhileReturnsEmptyWhenFirstItemFails verifies that TakeWhile produces
// an empty collection when the first item already fails the predicate.
func TestTakeWhileReturnsEmptyWhenFirstItemFails(t *testing.T) {
	t.Parallel()

	c := collection.NewSlice([]int{1, 2, 3})

	require.True(t, c.TakeWhile(func(i int, v int) bool { return v%2 == 0 }).IsEmpty())
}

// TestTakeWhileReturnsAllWhenPredicateAlwaysTrue verifies that TakeWhile
// returns all items when the predicate is always true.
func TestTakeWhileReturnsAllWhenPredicateAlwaysTrue(t *testing.T) {
	t.Parallel()

	c := collection.NewSlice([]int{2, 4, 6})

	require.Equal(t, []int{2, 4, 6}, c.TakeWhile(func(i int, v int) bool { return v%2 == 0 }).Items())
}

// TestSkipWhileSkipsLeadingMatchingItems verifies that SkipWhile drops items
// while the predicate holds and returns the remainder.
func TestSkipWhileSkipsLeadingMatchingItems(t *testing.T) {
	t.Parallel()

	c := collection.NewSlice([]int{2, 4, 5, 6})

	require.Equal(t, []int{5, 6}, c.SkipWhile(func(i int, v int) bool { return v%2 == 0 }).Items())
}

// TestSkipWhileReturnsAllWhenFirstItemFailsPredicate verifies that SkipWhile
// returns all items when the very first item does not match.
func TestSkipWhileReturnsAllWhenFirstItemFailsPredicate(t *testing.T) {
	t.Parallel()

	c := collection.NewSlice([]int{1, 2, 3})

	require.Equal(t, []int{1, 2, 3}, c.SkipWhile(func(i int, v int) bool { return v%2 == 0 }).Items())
}

// TestSkipWhileReturnsEmptyWhenPredicateAlwaysTrue verifies that SkipWhile
// returns an empty collection when the predicate is always true.
func TestSkipWhileReturnsEmptyWhenPredicateAlwaysTrue(t *testing.T) {
	t.Parallel()

	c := collection.NewSlice([]int{2, 4, 6})

	require.True(t, c.SkipWhile(func(i int, v int) bool { return v%2 == 0 }).IsEmpty())
}

// TestChunkSplitsEvenlyIntoSubCollections verifies that Chunk produces
// evenly-sized sub-collections when items divide cleanly.
func TestChunkSplitsEvenlyIntoSubCollections(t *testing.T) {
	t.Parallel()

	c := collection.NewSlice([]int{1, 2, 3, 4, 5, 6})
	chunks := c.Chunk(2)

	require.Len(t, chunks, 3)
	require.Equal(t, []int{1, 2}, chunks[0].Items())
	require.Equal(t, []int{3, 4}, chunks[1].Items())
	require.Equal(t, []int{5, 6}, chunks[2].Items())
}

// TestChunkLastChunkIsSmallerWhenItemsDontDivideEvenly verifies that the final
// chunk contains the remaining items when items do not divide evenly.
func TestChunkLastChunkIsSmallerWhenItemsDontDivideEvenly(t *testing.T) {
	t.Parallel()

	c := collection.NewSlice([]int{1, 2, 3, 4, 5})
	chunks := c.Chunk(2)

	require.Len(t, chunks, 3)
	require.Equal(t, []int{5}, chunks[2].Items())
}

// TestChunkSizeLargerThanLengthReturnsSingleChunk verifies that Chunk returns a
// single sub-collection when the chunk size exceeds the collection length.
func TestChunkSizeLargerThanLengthReturnsSingleChunk(t *testing.T) {
	t.Parallel()

	c := collection.NewSlice([]int{1, 2, 3})
	chunks := c.Chunk(10)

	require.Len(t, chunks, 1)
	require.Equal(t, []int{1, 2, 3}, chunks[0].Items())
}

// TestChunkEmptyCollectionReturnsNil verifies that Chunk on an empty collection
// returns nil.
func TestChunkEmptyCollectionReturnsNil(t *testing.T) {
	t.Parallel()

	c := collection.NewSlice[int](nil)

	require.Nil(t, c.Chunk(3))
}

// TestChunkPanicsOnZeroSize verifies that Chunk panics when given size=0.
func TestChunkPanicsOnZeroSize(t *testing.T) {
	t.Parallel()

	c := collection.NewSlice([]int{1, 2, 3})

	require.Panics(t, func() { c.Chunk(0) })
}

// TestChunkPanicsOnNegativeSize verifies that Chunk panics when given a
// negative size.
func TestChunkPanicsOnNegativeSize(t *testing.T) {
	t.Parallel()

	c := collection.NewSlice([]int{1, 2, 3})

	require.Panics(t, func() { c.Chunk(-1) })
}

// TestSlidingReturnsWindowsWithDefaultStepOfOne verifies that Sliding with no
// explicit step advances by one item between windows.
func TestSlidingReturnsWindowsWithDefaultStepOfOne(t *testing.T) {
	t.Parallel()

	c := collection.NewSlice([]int{1, 2, 3, 4})
	windows := c.Sliding(2)

	require.Len(t, windows, 3)
	require.Equal(t, []int{1, 2}, windows[0].Items())
	require.Equal(t, []int{2, 3}, windows[1].Items())
	require.Equal(t, []int{3, 4}, windows[2].Items())
}

// TestSlidingReturnsWindowsWithExplicitStep verifies that Sliding advances by
// the given step between windows.
func TestSlidingReturnsWindowsWithExplicitStep(t *testing.T) {
	t.Parallel()

	c := collection.NewSlice([]int{1, 2, 3, 4, 5})
	windows := c.Sliding(2, 2)

	require.Len(t, windows, 2)
	require.Equal(t, []int{1, 2}, windows[0].Items())
	require.Equal(t, []int{3, 4}, windows[1].Items())
}

// TestSlidingReturnsNilWhenCollectionSmallerThanWindowSize verifies that
// Sliding returns nil when the window size exceeds the collection length.
func TestSlidingReturnsNilWhenCollectionSmallerThanWindowSize(t *testing.T) {
	t.Parallel()

	c := collection.NewSlice([]int{1, 2})

	require.Nil(t, c.Sliding(5))
}

// TestSlidingPanicsOnZeroSize verifies that Sliding panics when size=0.
func TestSlidingPanicsOnZeroSize(t *testing.T) {
	t.Parallel()

	c := collection.NewSlice([]int{1, 2, 3})

	require.Panics(t, func() { c.Sliding(0) })
}

// TestSlidingPanicsOnNegativeSize verifies that Sliding panics when given a
// negative size.
func TestSlidingPanicsOnNegativeSize(t *testing.T) {
	t.Parallel()

	c := collection.NewSlice([]int{1, 2, 3})

	require.Panics(t, func() { c.Sliding(-1) })
}

// TestSlidingPanicsOnZeroStep verifies that Sliding panics when step=0.
func TestSlidingPanicsOnZeroStep(t *testing.T) {
	t.Parallel()

	c := collection.NewSlice([]int{1, 2, 3})

	require.Panics(t, func() { c.Sliding(2, 0) })
}

// TestSlidingPanicsOnNegativeStep verifies that Sliding panics when given a
// negative step.
func TestSlidingPanicsOnNegativeStep(t *testing.T) {
	t.Parallel()

	c := collection.NewSlice([]int{1, 2, 3})

	require.Panics(t, func() { c.Sliding(2, -1) })
}

// TestReduceAccumulatesValuesWithSameType verifies that Reduce sums integers
// when K and T are the same type.
func TestReduceAccumulatesValuesWithSameType(t *testing.T) {
	t.Parallel()

	c := collection.NewSlice([]int{1, 2, 3, 4, 5})

	sum := collection.Slice[int].Reduce(c, func(acc int, i int, v int) int { return acc + v }, 0)

	require.Equal(t, 15, sum)
}

// TestReduceAccumulatesWithDifferentKType verifies that Reduce can accumulate
// int32 items into an int64 result.
func TestReduceAccumulatesWithDifferentKType(t *testing.T) {
	t.Parallel()

	c := collection.NewSlice([]int32{1, 2, 3, 4, 5})

	sum := collection.Slice[int32].Reduce(c, func(acc int64, i int, v int32) int64 { return acc + int64(v) }, 0)

	require.Equal(t, int64(15), sum)
}

// TestReduceReturnsInitialValueOnEmptyCollection verifies that Reduce returns
// the initial value unchanged when the collection is empty.
func TestReduceReturnsInitialValueOnEmptyCollection(t *testing.T) {
	t.Parallel()

	c := collection.NewSlice[int](nil)

	result := collection.Slice[int].Reduce(c, func(acc int, i int, v int) int { return acc + v }, 42)

	require.Equal(t, 42, result)
}

// TestReverseReturnsReversedOrder verifies that Reverse returns items in
// reversed order.
func TestReverseReturnsReversedOrder(t *testing.T) {
	t.Parallel()

	c := collection.NewSlice([]int{1, 2, 3})

	require.Equal(t, []int{3, 2, 1}, c.Reverse().Items())
}

// TestReverseReturnsEmptyOnEmptyInput verifies that Reverse on an empty
// collection produces an empty collection.
func TestReverseReturnsEmptyOnEmptyInput(t *testing.T) {
	t.Parallel()

	c := collection.NewSlice[int](nil)

	require.True(t, c.Reverse().IsEmpty())
}

// TestReverseDoesNotMutateOriginal verifies that Reverse does not alter the
// original collection's order.
func TestReverseDoesNotMutateOriginal(t *testing.T) {
	t.Parallel()

	c := collection.NewSlice([]int{1, 2, 3})
	_ = c.Reverse()

	require.Equal(t, []int{1, 2, 3}, c.Items())
}

// TestSortReturnsSortedItems verifies that Sort returns items in ascending
// order using cmp.Compare.
func TestSortReturnsSortedItems(t *testing.T) {
	t.Parallel()

	c := collection.NewSlice([]int{3, 1, 4, 1, 5, 9})

	require.Equal(t, []int{1, 1, 3, 4, 5, 9}, c.Sort(cmp.Compare).Items())
}

// TestSortDoesNotMutateOriginal verifies that Sort does not alter the original
// collection.
func TestSortDoesNotMutateOriginal(t *testing.T) {
	t.Parallel()

	c := collection.NewSlice([]int{3, 1, 2})
	_ = c.Sort(cmp.Compare)

	require.Equal(t, []int{3, 1, 2}, c.Items())
}

// TestUniqueRemovesDuplicatesPreservingFirstOccurrence verifies that Unique
// keeps only the first occurrence of each key.
func TestUniqueRemovesDuplicatesPreservingFirstOccurrence(t *testing.T) {
	t.Parallel()

	c := collection.NewSlice([]int{1, 2, 1, 3, 2})

	result := collection.Slice[int].Unique(c, func(i int, v int) int { return v })

	require.Equal(t, []int{1, 2, 3}, result.Items())
}

// TestUniqueWithCustomKeyFunction verifies that Unique deduplicates by an
// arbitrary key (e.g. first letter of a string).
func TestUniqueWithCustomKeyFunction(t *testing.T) {
	t.Parallel()

	c := collection.NewSlice([]string{"apple", "avocado", "banana", "blueberry", "cherry"})

	result := collection.Slice[string].Unique(c, func(i int, v string) byte { return v[0] })

	require.Equal(t, []string{"apple", "banana", "cherry"}, result.Items())
}

// TestPartitionSplitsIntoMatchingAndNonMatching verifies that Partition
// correctly separates items.
func TestPartitionSplitsIntoMatchingAndNonMatching(t *testing.T) {
	t.Parallel()

	c := collection.NewSlice([]int{1, 2, 3, 4, 5})

	evens, odds := c.Partition(func(i int, v int) bool { return v%2 == 0 })

	require.Equal(t, []int{2, 4}, evens.Items())
	require.Equal(t, []int{1, 3, 5}, odds.Items())
}

// TestPartitionAllMatchSecondResultIsEmpty verifies that the second result of
// Partition is empty when all items match.
func TestPartitionAllMatchSecondResultIsEmpty(t *testing.T) {
	t.Parallel()

	c := collection.NewSlice([]int{2, 4, 6})

	_, rest := c.Partition(func(i int, v int) bool { return v%2 == 0 })

	require.True(t, rest.IsEmpty())
}

// TestPartitionNoneMatchFirstResultIsEmpty verifies that the first result of
// Partition is empty when no items match.
func TestPartitionNoneMatchFirstResultIsEmpty(t *testing.T) {
	t.Parallel()

	c := collection.NewSlice([]int{1, 3, 5})

	matching, _ := c.Partition(func(i int, v int) bool { return v%2 == 0 })

	require.True(t, matching.IsEmpty())
}

// TestConcatAppendsOtherItemsAfterSelf verifies that Concat appends the other
// collection's items in order after the receiver's items.
func TestConcatAppendsOtherItemsAfterSelf(t *testing.T) {
	t.Parallel()

	a := collection.NewSlice([]int{1, 2, 3})
	b := collection.NewSlice([]int{4, 5, 6})

	require.Equal(t, []int{1, 2, 3, 4, 5, 6}, a.Concat(b).Items())
}

// TestConcatDoesNotMutateOriginal verifies that Concat does not alter the
// receiver collection.
func TestConcatDoesNotMutateOriginal(t *testing.T) {
	t.Parallel()

	a := collection.NewSlice([]int{1, 2})
	b := collection.NewSlice([]int{3, 4})
	_ = a.Concat(b)

	require.Equal(t, []int{1, 2}, a.Items())
}

// TestGroupByGroupsItemsByKeyFunction verifies that GroupBy partitions items
// into sub-collections by the returned key.
func TestGroupByGroupsItemsByKeyFunction(t *testing.T) {
	t.Parallel()

	c := collection.NewSlice([]int{1, 2, 3, 4, 5, 6})

	grouped := collection.Slice[int].GroupBy(c, func(i int, v int) string {
		if v%2 == 0 {
			return "even"
		}

		return "odd"
	})

	require.Equal(t, 2, grouped.Len())

	odd, ok := grouped.Get("odd")
	require.True(t, ok)
	require.Equal(t, []int{1, 3, 5}, odd.Items())

	even, ok := grouped.Get("even")
	require.True(t, ok)
	require.Equal(t, []int{2, 4, 6}, even.Items())
}

// TestNthPicksEveryNthItemFromIndexZero verifies that Nth with no offset starts
// at index 0.
func TestNthPicksEveryNthItemFromIndexZero(t *testing.T) {
	t.Parallel()

	c := collection.NewSlice([]int{1, 2, 3, 4, 5, 6})

	require.Equal(t, []int{1, 3, 5}, c.Nth(2).Items())
}

// TestNthPicksEveryNthItemStartingAtGivenOffset verifies that Nth respects an
// explicit starting offset.
func TestNthPicksEveryNthItemStartingAtGivenOffset(t *testing.T) {
	t.Parallel()

	c := collection.NewSlice([]int{1, 2, 3, 4, 5, 6})

	require.Equal(t, []int{2, 4, 6}, c.Nth(2, 1).Items())
}

// TestNthOffsetBeyondLengthReturnsEmpty verifies that Nth with an offset
// beyond the collection length returns an empty collection.
func TestNthOffsetBeyondLengthReturnsEmpty(t *testing.T) {
	t.Parallel()

	c := collection.NewSlice([]int{1, 2, 3})

	require.True(t, c.Nth(1, 10).IsEmpty())
}

// TestNthPanicsOnZeroN verifies that Nth panics when n=0.
func TestNthPanicsOnZeroN(t *testing.T) {
	t.Parallel()

	c := collection.NewSlice([]int{1, 2, 3})

	require.Panics(t, func() { c.Nth(0) })
}

// TestNthPanicsOnNegativeN verifies that Nth panics when n is negative.
func TestNthPanicsOnNegativeN(t *testing.T) {
	t.Parallel()

	c := collection.NewSlice([]int{1, 2, 3})

	require.Panics(t, func() { c.Nth(-1) })
}

// TestForPageReturnsFirstPage verifies that ForPage(1, perPage) returns the
// first page of items.
func TestForPageReturnsFirstPage(t *testing.T) {
	t.Parallel()

	c := collection.NewSlice([]int{1, 2, 3, 4, 5, 6})

	require.Equal(t, []int{1, 2}, c.ForPage(1, 2).Items())
}

// TestForPageReturnsSecondPage verifies that ForPage(2, perPage) returns the
// second page of items.
func TestForPageReturnsSecondPage(t *testing.T) {
	t.Parallel()

	c := collection.NewSlice([]int{1, 2, 3, 4, 5, 6})

	require.Equal(t, []int{3, 4}, c.ForPage(2, 2).Items())
}

// TestForPageBeyondDataReturnsEmpty verifies that ForPage returns an empty
// collection when the requested page is past the end of the data.
func TestForPageBeyondDataReturnsEmpty(t *testing.T) {
	t.Parallel()

	c := collection.NewSlice([]int{1, 2, 3})

	require.True(t, c.ForPage(10, 2).IsEmpty())
}

// TestMarshalJSONMarshalsNonEmptyCollectionToJSONArray verifies that
// MarshalJSON produces a valid JSON array for a non-empty collection.
func TestMarshalJSONMarshalsNonEmptyCollectionToJSONArray(t *testing.T) {
	t.Parallel()

	c := collection.NewSlice([]int{1, 2, 3})

	data, err := json.Marshal(c)

	require.NoError(t, err)
	require.JSONEq(t, `[1,2,3]`, string(data))
}

// TestMarshalJSONMarshalsEmptyCollectionToEmptyArray verifies that MarshalJSON
// produces `[]` for a non-nil empty collection.
func TestMarshalJSONMarshalsEmptyCollectionToEmptyArray(t *testing.T) {
	t.Parallel()

	c := collection.NewSlice([]int{})

	data, err := json.Marshal(c)

	require.NoError(t, err)
	require.Equal(t, `[]`, string(data))
}

// TestUnmarshalJSONUnmarshalsJSONArrayIntoCollection verifies that
// UnmarshalJSON populates a collection from a JSON array.
func TestUnmarshalJSONUnmarshalsJSONArrayIntoCollection(t *testing.T) {
	t.Parallel()

	var c collection.Slice[int]
	err := json.Unmarshal([]byte(`[10,20,30]`), &c)

	require.NoError(t, err)
	require.Equal(t, []int{10, 20, 30}, c.Items())
}

// TestUnmarshalJSONOfInvalidJSONReturnsError verifies that UnmarshalJSON
// returns an error when given malformed JSON.
func TestUnmarshalJSONOfInvalidJSONReturnsError(t *testing.T) {
	t.Parallel()

	var c collection.Slice[int]
	err := json.Unmarshal([]byte(`not-json`), &c)

	require.Error(t, err)
}

// TestMarshalJSONToEncodesNonEmptyCollectionToJSONArray verifies that
// MarshalJSONTo writes a valid JSON array to the encoder.
func TestMarshalJSONToEncodesNonEmptyCollectionToJSONArray(t *testing.T) {
	t.Parallel()

	c := collection.NewSlice([]int{4, 5, 6})

	var buf bytes.Buffer
	enc := jsontext.NewEncoder(&buf)
	err := c.MarshalJSONTo(enc)

	require.NoError(t, err)
	require.JSONEq(t, `[4,5,6]`, buf.String())
}

// TestUnmarshalJSONFromDecodesJSONArrayFromDecoder verifies that
// UnmarshalJSONFrom populates the collection from a JSON array.
func TestUnmarshalJSONFromDecodesJSONArrayFromDecoder(t *testing.T) {
	t.Parallel()

	dec := jsontext.NewDecoder(strings.NewReader(`[7,8,9]`))

	var c collection.Slice[int]
	err := c.UnmarshalJSONFrom(dec)

	require.NoError(t, err)
	require.Equal(t, []int{7, 8, 9}, c.Items())
}

// TestUnmarshalJSONFromOfInvalidJSONReturnsError verifies that
// UnmarshalJSONFrom returns an error when given malformed JSON.
func TestUnmarshalJSONFromOfInvalidJSONReturnsError(t *testing.T) {
	t.Parallel()

	dec := jsontext.NewDecoder(strings.NewReader(`not-json`))

	var c collection.Slice[int]
	err := c.UnmarshalJSONFrom(dec)

	require.Error(t, err)
}
