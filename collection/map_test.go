package collection_test

import (
	"bytes"
	"encoding/json"
	"encoding/json/jsontext"
	"strings"
	"testing"

	"github.com/studiolambda/cosmos/collection"

	"github.com/stretchr/testify/require"
)

func TestNewMapItemsReturnsUnderlyingMap(t *testing.T) {
	t.Parallel()

	input := map[string]int{"a": 1, "b": 2}
	m := collection.NewMap(input)

	require.Equal(t, input, m.Items())
}

func TestNewMapAsAcceptsNamedMapType(t *testing.T) {
	t.Parallel()

	type counts map[string]int

	m := collection.NewMapAs(counts{"a": 1, "b": 2})

	require.Equal(t, map[string]int{"a": 1, "b": 2}, m.Items())
}

func TestNewMapCreatesMapFromPlainMap(t *testing.T) {
	t.Parallel()

	input := map[string]int{"a": 1, "b": 2}
	m := collection.NewMap(input)

	require.Equal(t, input, m.Items())
}

func TestNewMapNilItemsReturnsNilMap(t *testing.T) {
	t.Parallel()

	m := collection.NewMap[string, int](nil)

	require.Nil(t, m.Items())
}

func TestNewMapEmptyItems(t *testing.T) {
	t.Parallel()

	m := collection.NewMap(map[string]int{})

	require.Empty(t, m.Items())
}

func TestMapHasTrueWhenKeyExists(t *testing.T) {
	t.Parallel()

	m := collection.NewMap(map[string]int{"x": 10})

	require.True(t, m.Has("x"))
}

func TestMapHasFalseWhenKeyAbsent(t *testing.T) {
	t.Parallel()

	m := collection.NewMap(map[string]int{"x": 10})

	require.False(t, m.Has("z"))
}

func TestMapGetReturnsValueAndTrueWhenFound(t *testing.T) {
	t.Parallel()

	m := collection.NewMap(map[string]int{"k": 42})
	v, ok := m.Get("k")

	require.True(t, ok)
	require.Equal(t, 42, v)
}

func TestMapGetReturnsZeroAndFalseWhenNotFound(t *testing.T) {
	t.Parallel()

	m := collection.NewMap(map[string]int{"k": 42})
	v, ok := m.Get("missing")

	require.False(t, ok)
	require.Zero(t, v)
}

func TestMapHasAnyTrueWhenAnyKeyExists(t *testing.T) {
	t.Parallel()

	m := collection.NewMap(map[string]int{"a": 1, "b": 2})

	require.True(t, m.HasAny("x", "b"))
}

func TestMapHasAnyFalseWhenNoKeysExist(t *testing.T) {
	t.Parallel()

	m := collection.NewMap(map[string]int{"a": 1, "b": 2})

	require.False(t, m.HasAny("x", "y"))
}

func TestMapLenReturnsCorrectCount(t *testing.T) {
	t.Parallel()

	m := collection.NewMap(map[string]int{"a": 1, "b": 2, "c": 3})

	require.Equal(t, 3, m.Len())
}

func TestMapLenZeroForEmpty(t *testing.T) {
	t.Parallel()

	m := collection.NewMap(map[string]int{})

	require.Zero(t, m.Len())
}

func TestMapIsEmptyTrueForEmpty(t *testing.T) {
	t.Parallel()

	m := collection.NewMap(map[string]int{})

	require.True(t, m.IsEmpty())
}

func TestMapIsEmptyFalseForNonEmpty(t *testing.T) {
	t.Parallel()

	m := collection.NewMap(map[string]int{"a": 1})

	require.False(t, m.IsEmpty())
}

func TestMapIsNotEmptySymmetric(t *testing.T) {
	t.Parallel()

	empty := collection.NewMap(map[string]int{})
	nonEmpty := collection.NewMap(map[string]int{"a": 1})

	require.False(t, empty.IsNotEmpty())
	require.True(t, nonEmpty.IsNotEmpty())
}

func TestMapEachCallsFForEveryEntry(t *testing.T) {
	t.Parallel()

	input := map[string]int{"a": 1, "b": 2, "c": 3}
	m := collection.NewMap(input)
	collected := make(map[string]int)

	m.Each(func(k string, v int) {
		collected[k] = v
	})

	require.Equal(t, input, collected)
}

func TestMapTapEachCallsFAndReturnsOriginal(t *testing.T) {
	t.Parallel()

	input := map[string]int{"a": 1, "b": 2}
	m := collection.NewMap(input)
	collected := make(map[string]int)

	result := m.TapEach(func(k string, v int) {
		collected[k] = v
	})

	require.Equal(t, input, collected)
	require.Equal(t, input, result.Items())
}

func TestMapEveryTrueWhenAllMatch(t *testing.T) {
	t.Parallel()

	m := collection.NewMap(map[string]int{"a": 2, "b": 4, "c": 6})

	require.True(t, m.Every(func(_ string, v int) bool { return v%2 == 0 }))
}

func TestMapEveryFalseWhenOneFails(t *testing.T) {
	t.Parallel()

	m := collection.NewMap(map[string]int{"a": 2, "b": 3, "c": 6})

	require.False(t, m.Every(func(_ string, v int) bool { return v%2 == 0 }))
}

func TestMapEveryTrueOnEmpty(t *testing.T) {
	t.Parallel()

	m := collection.NewMap(map[string]int{})

	require.True(t, m.Every(func(_ string, v int) bool { return false }))
}

func TestMapContainsTrueWhenAnyEntryMatches(t *testing.T) {
	t.Parallel()

	m := collection.NewMap(map[string]int{"a": 1, "b": 99, "c": 3})

	require.True(t, m.Contains(func(_ string, v int) bool { return v == 99 }))
}

func TestMapContainsFalseWhenNoneMatch(t *testing.T) {
	t.Parallel()

	m := collection.NewMap(map[string]int{"a": 1, "b": 2})

	require.False(t, m.Contains(func(_ string, v int) bool { return v == 99 }))
}

func TestMapFilterReturnsMatchingEntries(t *testing.T) {
	t.Parallel()

	m := collection.NewMap(map[string]int{"a": 1, "b": 2, "c": 3, "d": 4})
	filtered := m.Filter(func(_ string, v int) bool { return v%2 == 0 })

	require.Equal(t, map[string]int{"b": 2, "d": 4}, filtered.Items())
}

func TestMapFilterReturnsEmptyWhenNothingMatches(t *testing.T) {
	t.Parallel()

	m := collection.NewMap(map[string]int{"a": 1, "b": 3})
	filtered := m.Filter(func(_ string, v int) bool { return v%2 == 0 })

	require.Empty(t, filtered.Items())
}

func TestMapRejectReturnsNonMatchingEntries(t *testing.T) {
	t.Parallel()

	m := collection.NewMap(map[string]int{"a": 1, "b": 2, "c": 3, "d": 4})
	rejected := m.Reject(func(_ string, v int) bool { return v%2 == 0 })

	require.Equal(t, map[string]int{"a": 1, "c": 3}, rejected.Items())
}

func TestMapOnlyKeepsRequestedKeys(t *testing.T) {
	t.Parallel()

	m := collection.NewMap(map[string]int{"a": 1, "b": 2, "c": 3})

	result := m.Only("b", "missing")

	require.Equal(t, map[string]int{"b": 2}, result.Items())
}

func TestMapExceptRemovesRequestedKeys(t *testing.T) {
	t.Parallel()

	m := collection.NewMap(map[string]int{"a": 1, "b": 2, "c": 3})

	result := m.Except("b", "missing")

	require.Equal(t, map[string]int{"a": 1, "c": 3}, result.Items())
}

func TestMapValuesTransformsValuesPreservesKeys(t *testing.T) {
	t.Parallel()

	m := collection.NewMap(map[string]int{"a": 1, "b": 2})
	result := collection.MapValues(m, func(k string, v int) string { return strings.Repeat("x", v) })

	require.Equal(t, map[string]string{"a": "x", "b": "xx"}, result.Items())
}

func TestMapKeysTransformsKeysPreservesValues(t *testing.T) {
	t.Parallel()

	m := collection.NewMap(map[string]int{"a": 1, "b": 2})
	result := collection.MapKeys(m, func(k string, v int) string { return strings.ToUpper(k) })

	require.Equal(t, map[string]int{"A": 1, "B": 2}, result.Items())
}

func TestMapKeysLastWriteWinsOnCollision(t *testing.T) {
	t.Parallel()

	// Both "a" and "A" map to "X" via ToUpper-then-force; use a fixed collision.
	m := collection.NewMap(map[string]int{"hello": 1, "world": 2})
	// Map all keys to the same key — last write wins.
	result := collection.MapKeys(m, func(_ string, _ int) string { return "same" })

	require.Equal(t, 1, result.Len())
}

func TestKeysContainsAllKeys(t *testing.T) {
	t.Parallel()

	m := collection.NewMap(map[string]int{"a": 1, "b": 2, "c": 3})
	keys := collection.Keys(m)

	require.ElementsMatch(t, []string{"a", "b", "c"}, keys.Items())
}

func TestValuesContainsAllValues(t *testing.T) {
	t.Parallel()

	m := collection.NewMap(map[string]int{"a": 1, "b": 2, "c": 3})
	values := collection.Values(m)

	require.ElementsMatch(t, []int{1, 2, 3}, values.Items())
}

func TestMapMergeOtherValuesTakePrecedenceOnCollision(t *testing.T) {
	t.Parallel()

	a := collection.NewMap(map[string]int{"x": 1, "y": 2})
	b := collection.NewMap(map[string]int{"x": 99})
	merged := a.Merge(b)

	v, ok := merged.Get("x")
	require.True(t, ok)
	require.Equal(t, 99, v)
}

func TestMapMergePreservesEntriesUniqueToSelf(t *testing.T) {
	t.Parallel()

	a := collection.NewMap(map[string]int{"x": 1, "y": 2})
	b := collection.NewMap(map[string]int{"z": 3})
	merged := a.Merge(b)

	require.True(t, merged.Has("y"))
}

func TestMapMergeAddsEntriesUniqueToOther(t *testing.T) {
	t.Parallel()

	a := collection.NewMap(map[string]int{"x": 1})
	b := collection.NewMap(map[string]int{"z": 3})
	merged := a.Merge(b)

	require.True(t, merged.Has("z"))
	require.Equal(t, 3, merged.Items()["z"])
}

func TestMapLazyYieldsAllEntries(t *testing.T) {
	t.Parallel()

	input := map[string]int{"a": 1, "b": 2, "c": 3}
	m := collection.NewMap(input)
	lazy := m.Lazy()

	require.Equal(t, input, lazy.Items())
}

func TestInvertSwapsKeysAndValues(t *testing.T) {
	t.Parallel()

	m := collection.NewMap(map[string]int{"a": 1, "b": 2})
	inverted := collection.Invert(m)

	require.Equal(t, map[int]string{1: "a", 2: "b"}, inverted.Items())
}

func TestInvertLastWriteWinsOnValueCollision(t *testing.T) {
	t.Parallel()

	// Two keys with the same value — inverted map has one entry.
	m := collection.NewMap(map[string]int{"a": 1, "b": 1})
	inverted := collection.Invert(m)

	require.Equal(t, 1, inverted.Len())
	require.True(t, inverted.Has(1))
}

func TestMapMarshalJSONProducesObject(t *testing.T) {
	t.Parallel()

	m := collection.NewMap(map[string]int{"a": 1, "b": 2})
	data, err := json.Marshal(m)

	require.NoError(t, err)

	var got map[string]int
	require.NoError(t, json.Unmarshal(data, &got))
	require.Equal(t, map[string]int{"a": 1, "b": 2}, got)
}

func TestMapUnmarshalJSONFromObject(t *testing.T) {
	t.Parallel()

	data := []byte(`{"x":10,"y":20}`)
	var m collection.Map[string, int]
	require.NoError(t, json.Unmarshal(data, &m))

	require.Equal(t, map[string]int{"x": 10, "y": 20}, m.Items())
}

func TestMapUnmarshalJSONInvalidReturnsError(t *testing.T) {
	t.Parallel()

	var m collection.Map[string, int]
	err := json.Unmarshal([]byte(`not json`), &m)

	require.Error(t, err)
}

func TestMapMarshalJSONToEncodesAsObject(t *testing.T) {
	t.Parallel()

	m := collection.NewMap(map[string]int{"a": 1})
	var buf bytes.Buffer
	enc := jsontext.NewEncoder(&buf)

	require.NoError(t, m.MarshalJSONTo(enc))

	var got map[string]int
	require.NoError(t, json.Unmarshal(buf.Bytes(), &got))
	require.Equal(t, map[string]int{"a": 1}, got)
}

func TestMapUnmarshalJSONFromDecodesFromDecoder(t *testing.T) {
	t.Parallel()

	data := []byte(`{"p":7,"q":8}`)
	dec := jsontext.NewDecoder(bytes.NewReader(data))

	var m collection.Map[string, int]
	require.NoError(t, m.UnmarshalJSONFrom(dec))
	require.Equal(t, map[string]int{"p": 7, "q": 8}, m.Items())
}
