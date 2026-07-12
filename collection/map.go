package collection

import (
	"encoding/json/jsontext"
	"encoding/json/v2"
	"iter"
	"maps"
	"slices"
)

// Map is an eager, in-memory map collection of key-value pairs of type K and V.
type Map[K comparable, V any] struct {
	items map[K]V
}

// NewMap creates a new [Map] from the given map.
func NewMap[K comparable, V any](items map[K]V) Map[K, V] {
	return Map[K, V]{items: items}
}

// Items returns the underlying map.
func (mapping Map[K, V]) Items() map[K]V {
	return mapping.items
}

func (mapping Map[K, V]) Iter() iter.Seq2[K, V] {
	return maps.All(mapping.Items())
}

func (mapping Map[K, V]) Lazy() LazyMap[K, V] {
	return NewLazyMap(mapping.Iter())
}

// Has reports whether the key exists in the map.
func (mapping Map[K, V]) Has(key K) bool {
	_, ok := mapping.items[key]

	return ok
}

// Get returns the value for the key and whether it was present.
func (mapping Map[K, V]) Get(key K) (V, bool) {
	v, ok := mapping.items[key]

	return v, ok
}

// HasAny reports whether any of the given keys exists in the map.
func (mapping Map[K, V]) HasAny(keys ...K) bool {
	for _, key := range keys {
		if mapping.Has(key) {
			return true
		}
	}

	return false
}

// Len returns the number of entries in the map.
func (mapping Map[K, V]) Len() int {
	return len(mapping.items)
}

// IsEmpty reports whether the map contains no entries.
func (mapping Map[K, V]) IsEmpty() bool {
	return len(mapping.items) == 0
}

// IsNotEmpty reports whether the map contains at least one entry.
func (mapping Map[K, V]) IsNotEmpty() bool {
	return len(mapping.items) > 0
}

// Each calls f for every entry in the map. Order is not guaranteed.
func (mapping Map[K, V]) Each(f func(K, V)) {
	for k, v := range mapping.items {
		f(k, v)
	}
}

// TapEach calls f for every entry and returns the map unchanged. Order is not guaranteed.
func (mapping Map[K, V]) TapEach(f func(K, V)) Map[K, V] {
	for k, v := range mapping.items {
		f(k, v)
	}

	return mapping
}

// Every reports whether f returns true for all entries. Returns true on empty.
func (mapping Map[K, V]) Every(f func(K, V) bool) bool {
	for k, v := range mapping.items {
		if !f(k, v) {
			return false
		}
	}

	return true
}

// Contains reports whether any entry satisfies f.
func (mapping Map[K, V]) Contains(f func(K, V) bool) bool {
	for k, v := range mapping.items {
		if f(k, v) {
			return true
		}
	}

	return false
}

// Filter returns a new [Map] containing only entries for which f returns true.
func (mapping Map[K, V]) Filter(f func(K, V) bool) Map[K, V] {
	result := make(map[K]V, len(mapping.items))

	for k, v := range mapping.items {
		if f(k, v) {
			result[k] = v
		}
	}

	return NewMap(result)
}

// Reject returns a new [Map] containing only entries for which f returns false.
func (mapping Map[K, V]) Reject(f func(K, V) bool) Map[K, V] {
	return mapping.Filter(func(k K, v V) bool { return !f(k, v) })
}

// Only returns a new [Map] containing only the entries whose keys were given.
func (mapping Map[K, V]) Only(keys ...K) Map[K, V] {
	result := make(map[K]V, len(keys))

	for _, key := range keys {
		if value, ok := mapping.items[key]; ok {
			result[key] = value
		}
	}

	return NewMap(result)
}

// Except returns a new [Map] containing all entries except those whose keys were given.
func (mapping Map[K, V]) Except(keys ...K) Map[K, V] {
	result := maps.Clone(mapping.items)

	for _, key := range keys {
		delete(result, key)
	}

	return NewMap(result)
}

// MapValues transforms values using f and returns a new [Map] with the same keys.
func MapValues[K comparable, V, W any](mapping Map[K, V], f func(V) W) Map[K, W] {
	result := make(map[K]W, len(mapping.items))

	for k, v := range mapping.items {
		result[k] = f(v)
	}

	return NewMap(result)
}

// MapKeys transforms keys using f and returns a new [Map] with the same values.
// When multiple keys map to the same new key, the last one encountered wins.
func MapKeys[K comparable, J comparable, V any](mapping Map[K, V], f func(K) J) Map[J, V] {
	result := make(map[J]V, len(mapping.items))

	for k, v := range mapping.items {
		result[f(k)] = v
	}

	return NewMap(result)
}

// rawKeys returns all keys as a plain Go slice.
func (mapping Map[K, V]) rawKeys() []K {
	return slices.Collect(maps.Keys(mapping.items))
}

// rawValues returns all values as a plain Go slice.
func (mapping Map[K, V]) rawValues() []V {
	return slices.Collect(maps.Values(mapping.items))
}

// Keys returns a [Slice] of all keys in the given [Map]. Order is not guaranteed.
func Keys[K comparable, V any](mapping Map[K, V]) Slice[K] {
	return NewSlice(mapping.rawKeys())
}

// Values returns a [Slice] of all values in the given [Map]. Order is not guaranteed.
func Values[K comparable, V any](mapping Map[K, V]) Slice[V] {
	return NewSlice(mapping.rawValues())
}

// Merge returns a new [Map] with all entries from both maps.
// When keys collide, other's values take precedence.
func (mapping Map[K, V]) Merge(other Map[K, V]) Map[K, V] {
	result := maps.Clone(mapping.items)
	maps.Copy(result, other.items)

	return NewMap(result)
}

// MarshalJSONTo encodes the map as a JSON object into enc.
func (mapping Map[K, V]) MarshalJSONTo(enc *jsontext.Encoder) error {
	return json.MarshalEncode(enc, mapping.items)
}

// UnmarshalJSONFrom decodes a JSON object from dec into the map.
func (mapping *Map[K, V]) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	return json.UnmarshalDecode(dec, &mapping.items)
}

// MarshalJSON implements [encoding/json.Marshaler] for v1 compatibility.
// When used with [encoding/json/v2], [Map.MarshalJSONTo] takes precedence.
func (mapping Map[K, V]) MarshalJSON() ([]byte, error) {
	return json.Marshal(mapping.items)
}

// UnmarshalJSON implements [encoding/json.Unmarshaler] for v1 compatibility.
// When used with [encoding/json/v2], [Map.UnmarshalJSONFrom] takes precedence.
func (mapping *Map[K, V]) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &mapping.items)
}

// Invert returns a new [Map] with keys and values swapped.
// When multiple keys share the same value, the last one encountered wins.
func Invert[K, V comparable](m Map[K, V]) Map[V, K] {
	result := make(map[V]K, len(m.items))

	for k, v := range m.items {
		result[v] = k
	}

	return NewMap(result)
}
