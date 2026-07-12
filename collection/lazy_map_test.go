package collection_test

import (
	"bytes"
	"encoding/json"
	"encoding/json/jsontext"
	"maps"
	"strings"
	"testing"

	"github.com/studiolambda/cosmos/collection"

	"github.com/stretchr/testify/require"
)

// orderedSource returns a LazyMap with a deterministic iteration order.
func orderedSource() collection.LazyMap[string, int] {
	return collection.NewLazyMap(func(yield func(string, int) bool) {
		keys := []string{"a", "b", "c"}
		vals := []int{1, 2, 3}

		for i, k := range keys {
			if !yield(k, vals[i]) {
				return
			}
		}
	})
}

func TestNewLazyMapCanBeUsedInForRange(t *testing.T) {
	t.Parallel()

	lm := collection.NewLazyMap(maps.All(map[string]int{"x": 1, "y": 2}))
	collected := make(map[string]int)

	for k, v := range lm {
		collected[k] = v
	}

	require.Equal(t, map[string]int{"x": 1, "y": 2}, collected)
}

func TestNewLazyMapYieldsAllEntriesFromSource(t *testing.T) {
	t.Parallel()

	input := map[string]int{"a": 1, "b": 2, "c": 3}
	lm := collection.NewLazyMap(maps.All(input))

	require.Equal(t, input, lm.Items())
}

func TestLazyMapEagerMaterializesToMap(t *testing.T) {
	t.Parallel()

	input := map[string]int{"a": 1, "b": 2}
	lm := collection.NewLazyMap(maps.All(input))
	eager := lm.Eager()

	require.Equal(t, input, eager.Items())
}

func TestLazyMapItemsReturnsGoMap(t *testing.T) {
	t.Parallel()

	input := map[string]int{"x": 10}
	lm := collection.NewLazyMap(maps.All(input))

	require.Equal(t, input, lm.Items())
}

func TestLazyMapIsEmptyTrueForEmpty(t *testing.T) {
	t.Parallel()

	lm := collection.NewLazyMap(maps.All(map[string]int{}))

	require.True(t, lm.IsEmpty())
}

func TestLazyMapIsEmptyFalseForNonEmpty(t *testing.T) {
	t.Parallel()

	lm := collection.NewLazyMap(maps.All(map[string]int{"a": 1}))

	require.False(t, lm.IsEmpty())
}

func TestLazyMapIsNotEmptyTrueForNonEmpty(t *testing.T) {
	t.Parallel()

	lm := collection.NewLazyMap(maps.All(map[string]int{"a": 1}))

	require.True(t, lm.IsNotEmpty())
}

func TestLazyMapIsNotEmptyFalseForEmpty(t *testing.T) {
	t.Parallel()

	lm := collection.NewLazyMap(maps.All(map[string]int{}))

	require.False(t, lm.IsNotEmpty())
}

func TestLazyMapLenReturnsCorrectCount(t *testing.T) {
	t.Parallel()

	lm := collection.NewLazyMap(maps.All(map[string]int{"a": 1, "b": 2, "c": 3}))

	require.Equal(t, 3, lm.Len())
}

func TestLazyMapLenZeroForEmpty(t *testing.T) {
	t.Parallel()

	lm := collection.NewLazyMap(maps.All(map[string]int{}))

	require.Zero(t, lm.Len())
}

func TestLazyMapEveryTrueWhenAllMatch(t *testing.T) {
	t.Parallel()

	lm := collection.NewLazyMap(maps.All(map[string]int{"a": 2, "b": 4}))

	require.True(t, lm.Every(func(_ string, v int) bool { return v%2 == 0 }))
}

func TestLazyMapEveryFalseWhenOneFails(t *testing.T) {
	t.Parallel()

	lm := collection.NewLazyMap(maps.All(map[string]int{"a": 2, "b": 3}))

	require.False(t, lm.Every(func(_ string, v int) bool { return v%2 == 0 }))
}

func TestLazyMapEveryTrueOnEmpty(t *testing.T) {
	t.Parallel()

	lm := collection.NewLazyMap(maps.All(map[string]int{}))

	require.True(t, lm.Every(func(_ string, v int) bool { return false }))
}

func TestLazyMapEachCallsFForEveryEntry(t *testing.T) {
	t.Parallel()

	input := map[string]int{"a": 1, "b": 2}
	lm := collection.NewLazyMap(maps.All(input))
	collected := make(map[string]int)

	lm.Each(func(k string, v int) {
		collected[k] = v
	})

	require.Equal(t, input, collected)
}

func TestLazyMapTapEachIsLazyBeforeConsumption(t *testing.T) {
	t.Parallel()

	lm := collection.NewLazyMap(maps.All(map[string]int{"a": 1}))
	called := false

	tapped := lm.TapEach(func(_ string, _ int) {
		called = true
	})

	// f must not have been called yet — only the lazy map was constructed.
	require.False(t, called)
	// Consume to trigger f.
	_ = tapped.Items()
	require.True(t, called)
}

func TestLazyMapTapEachCallsFUponConsumption(t *testing.T) {
	t.Parallel()

	input := map[string]int{"a": 1, "b": 2}
	lm := collection.NewLazyMap(maps.All(input))
	collected := make(map[string]int)

	tapped := lm.TapEach(func(k string, v int) {
		collected[k] = v
	})

	// Consume.
	_ = tapped.Items()

	require.Equal(t, input, collected)
}

func TestLazyMapContainsTrueWhenFound(t *testing.T) {
	t.Parallel()

	lm := collection.NewLazyMap(maps.All(map[string]int{"a": 1, "b": 99}))

	require.True(t, lm.Contains(func(_ string, v int) bool { return v == 99 }))
}

func TestLazyMapContainsFalseWhenNotFound(t *testing.T) {
	t.Parallel()

	lm := collection.NewLazyMap(maps.All(map[string]int{"a": 1, "b": 2}))

	require.False(t, lm.Contains(func(_ string, v int) bool { return v == 99 }))
}

func TestLazyMapHasAnyTrueWhenAnyKeyExists(t *testing.T) {
	t.Parallel()

	lm := orderedSource()

	require.True(t, lm.HasAny("x", "b"))
}

func TestLazyMapHasAnyFalseWhenNoKeysExist(t *testing.T) {
	t.Parallel()

	lm := orderedSource()

	require.False(t, lm.HasAny("x", "y"))
}

func TestLazyMapFilterOnlyMatchingEntriesWhenMaterialized(t *testing.T) {
	t.Parallel()

	lm := collection.NewLazyMap(maps.All(map[string]int{"a": 1, "b": 2, "c": 3, "d": 4}))
	filtered := lm.Filter(func(_ string, v int) bool { return v%2 == 0 }).Items()

	require.Equal(t, map[string]int{"b": 2, "d": 4}, filtered)
}

func TestLazyMapFilterEmptyResultWhenNothingMatches(t *testing.T) {
	t.Parallel()

	lm := collection.NewLazyMap(maps.All(map[string]int{"a": 1, "b": 3}))
	filtered := lm.Filter(func(_ string, v int) bool { return v%2 == 0 }).Items()

	require.Empty(t, filtered)
}

func TestLazyMapRejectOnlyNonMatchingEntries(t *testing.T) {
	t.Parallel()

	lm := collection.NewLazyMap(maps.All(map[string]int{"a": 1, "b": 2, "c": 3, "d": 4}))
	rejected := lm.Reject(func(_ string, v int) bool { return v%2 == 0 }).Items()

	require.Equal(t, map[string]int{"a": 1, "c": 3}, rejected)
}

func TestLazyMapOnlyKeepsRequestedKeys(t *testing.T) {
	t.Parallel()

	result := orderedSource().Only("b", "missing").Items()

	require.Equal(t, map[string]int{"b": 2}, result)
}

func TestLazyMapExceptRemovesRequestedKeys(t *testing.T) {
	t.Parallel()

	result := orderedSource().Except("b", "missing").Items()

	require.Equal(t, map[string]int{"a": 1, "c": 3}, result)
}

func TestLazyMapValuesTransformsValuesLazily(t *testing.T) {
	t.Parallel()

	lm := collection.NewLazyMap(maps.All(map[string]int{"a": 1, "b": 2}))
	result := collection.LazyMapValues(lm, func(v int) string {
		return strings.Repeat("*", v)
	}).Items()

	require.Equal(t, map[string]string{"a": "*", "b": "**"}, result)
}

func TestLazyMapKeysTransformsKeysLazily(t *testing.T) {
	t.Parallel()

	lm := collection.NewLazyMap(maps.All(map[string]int{"a": 1, "b": 2}))
	result := collection.LazyMapKeys(lm, func(k string) string {
		return strings.ToUpper(k)
	}).Items()

	require.Equal(t, map[string]int{"A": 1, "B": 2}, result)
}

func TestLazyKeysReturnsLazySliceOfKeys(t *testing.T) {
	t.Parallel()

	lm := collection.NewLazyMap(maps.All(map[string]int{"a": 1, "b": 2, "c": 3}))
	keys := collection.LazyKeys(lm).Items()

	require.ElementsMatch(t, []string{"a", "b", "c"}, keys)
}

func TestLazyValuesReturnsLazySliceOfValues(t *testing.T) {
	t.Parallel()

	lm := collection.NewLazyMap(maps.All(map[string]int{"a": 1, "b": 2, "c": 3}))
	values := collection.LazyValues(lm).Items()

	require.ElementsMatch(t, []int{1, 2, 3}, values)
}

func TestLazyMapTakeYieldsAtMostNEntries(t *testing.T) {
	t.Parallel()

	source := orderedSource()
	taken := source.Take(2).Items()

	require.Len(t, taken, 2)
}

func TestLazyMapTakeZeroLimitYieldsNothing(t *testing.T) {
	t.Parallel()

	source := orderedSource()
	taken := source.Take(0).Items()

	require.Empty(t, taken)
}

func TestLazyMapSkipSkipsFirstNEntries(t *testing.T) {
	t.Parallel()

	source := orderedSource() // a:1, b:2, c:3 in deterministic order
	skipped := source.Skip(2).Items()

	require.Equal(t, map[string]int{"c": 3}, skipped)
}

func TestLazyMapSkipZeroYieldsAll(t *testing.T) {
	t.Parallel()

	source := orderedSource()
	skipped := source.Skip(0).Items()

	require.Equal(t, map[string]int{"a": 1, "b": 2, "c": 3}, skipped)
}

func TestLazyMapUniqueDeduplicatesByKeyKeepingFirstOccurrence(t *testing.T) {
	t.Parallel()

	// Hand-written iterator with a duplicate key "a".
	source := collection.NewLazyMap(func(yield func(string, int) bool) {
		for _, e := range [][2]any{{"a", 1}, {"b", 2}, {"a", 99}} {
			if !yield(e[0].(string), e[1].(int)) {
				return
			}
		}
	})

	result := source.Unique().Items()

	require.Equal(t, map[string]int{"a": 1, "b": 2}, result)
}

func TestLazyMapConcatYieldsEntriesFromBothMaps(t *testing.T) {
	t.Parallel()

	first := collection.NewLazyMap(maps.All(map[string]int{"a": 1}))
	second := collection.NewLazyMap(maps.All(map[string]int{"b": 2}))
	combined := first.Concat(second).Items()

	require.Equal(t, map[string]int{"a": 1, "b": 2}, combined)
}

func TestLazyMapMarshalJSONProducesObject(t *testing.T) {
	t.Parallel()

	lm := collection.NewLazyMap(maps.All(map[string]int{"a": 1, "b": 2}))
	data, err := json.Marshal(lm)

	require.NoError(t, err)

	var got map[string]int
	require.NoError(t, json.Unmarshal(data, &got))
	require.Equal(t, map[string]int{"a": 1, "b": 2}, got)
}

func TestLazyMapUnmarshalJSONFromObject(t *testing.T) {
	t.Parallel()

	data := []byte(`{"x":10,"y":20}`)
	var lm collection.LazyMap[string, int]
	require.NoError(t, json.Unmarshal(data, &lm))

	require.Equal(t, map[string]int{"x": 10, "y": 20}, lm.Items())
}

func TestLazyMapUnmarshalJSONInvalidReturnsError(t *testing.T) {
	t.Parallel()

	var lm collection.LazyMap[string, int]
	err := json.Unmarshal([]byte(`not json`), &lm)

	require.Error(t, err)
}

func TestLazyMapMarshalJSONToEncodesAsObject(t *testing.T) {
	t.Parallel()

	lm := collection.NewLazyMap(maps.All(map[string]int{"a": 1}))
	var buf bytes.Buffer
	enc := jsontext.NewEncoder(&buf)

	require.NoError(t, lm.MarshalJSONTo(enc))

	var got map[string]int
	require.NoError(t, json.Unmarshal(buf.Bytes(), &got))
	require.Equal(t, map[string]int{"a": 1}, got)
}

func TestLazyMapUnmarshalJSONFromDecodesFromDecoder(t *testing.T) {
	t.Parallel()

	data := []byte(`{"p":7,"q":8}`)
	dec := jsontext.NewDecoder(bytes.NewReader(data))

	var lm collection.LazyMap[string, int]
	require.NoError(t, lm.UnmarshalJSONFrom(dec))
	require.Equal(t, map[string]int{"p": 7, "q": 8}, lm.Items())
}
