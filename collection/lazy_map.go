package collection

import (
	"encoding/json/jsontext"
	"encoding/json/v2"
	"iter"
	"maps"
)

// LazyMap is a lazily-evaluated map collection backed by an [iter.Seq2].
// Being a function type, it can be used directly in a for-range loop:
//
//	for k, v := range lazyMap { ... }
type LazyMap[K comparable, V any] func(yield func(K, V) bool)

// NewLazyMap creates a new [LazyMap] from the given iterator.
func NewLazyMap[K comparable, V any](seq iter.Seq2[K, V]) LazyMap[K, V] {
	return LazyMap[K, V](seq)
}

// Eager materializes the lazy map into an eager [Map].
func (lazyMapping LazyMap[K, V]) Eager() Map[K, V] {
	return NewMap(lazyMapping.Items())
}

// Items materializes all entries from the iterator into a Go map.
func (lazyMapping LazyMap[K, V]) Items() map[K]V {
	result := make(map[K]V)

	for k, v := range lazyMapping {
		result[k] = v
	}

	return result
}

// IsEmpty reports whether the iterator yields no entries.
func (lazyMapping LazyMap[K, V]) IsEmpty() bool {
	for range lazyMapping {
		return false
	}

	return true
}

// IsNotEmpty reports whether the iterator yields at least one entry.
func (lazyMapping LazyMap[K, V]) IsNotEmpty() bool {
	return !lazyMapping.IsEmpty()
}

// Len returns the total number of entries yielded by the iterator.
// It fully consumes the sequence.
func (lazyMapping LazyMap[K, V]) Len() int {
	count := 0

	for range lazyMapping {
		count++
	}

	return count
}

// Every reports whether f returns true for every entry yielded by the iterator.
// Returns true on empty.
func (lazyMapping LazyMap[K, V]) Every(f func(K, V) bool) bool {
	for k, v := range lazyMapping {
		if !f(k, v) {
			return false
		}
	}

	return true
}

// Each calls f for every entry yielded by the iterator.
func (lazyMapping LazyMap[K, V]) Each(f func(K, V)) {
	for k, v := range lazyMapping {
		f(k, v)
	}
}

// TapEach returns a [LazyMap] that calls f on each entry as it is consumed,
// then yields the same entry unchanged.
func (lazyMapping LazyMap[K, V]) TapEach(f func(K, V)) LazyMap[K, V] {
	return NewLazyMap(func(yield func(K, V) bool) {
		for k, v := range lazyMapping {
			f(k, v)

			if !yield(k, v) {
				return
			}
		}
	})
}

// Contains reports whether any entry satisfies f.
func (lazyMapping LazyMap[K, V]) Contains(f func(K, V) bool) bool {
	for k, v := range lazyMapping {
		if f(k, v) {
			return true
		}
	}

	return false
}

// Filter returns a [LazyMap] containing only entries for which f returns true.
func (lazyMapping LazyMap[K, V]) Filter(f func(K, V) bool) LazyMap[K, V] {
	return NewLazyMap(func(yield func(K, V) bool) {
		for k, v := range lazyMapping {
			if f(k, v) {
				if !yield(k, v) {
					return
				}
			}
		}
	})
}

// Reject returns a [LazyMap] containing only entries for which f returns false.
func (lazyMapping LazyMap[K, V]) Reject(f func(K, V) bool) LazyMap[K, V] {
	return lazyMapping.Filter(func(k K, v V) bool { return !f(k, v) })
}

// LazyMapValues transforms values using f and returns a new [LazyMap] with the same keys.
func LazyMapValues[K comparable, V, W any](lazyMapping LazyMap[K, V], f func(V) W) LazyMap[K, W] {
	return NewLazyMap(func(yield func(K, W) bool) {
		for k, v := range lazyMapping {
			if !yield(k, f(v)) {
				return
			}
		}
	})
}

// LazyMapKeys transforms keys using f and returns a new [LazyMap] with the same values.
// When multiple keys map to the same new key, the last one encountered wins upon materialisation.
func LazyMapKeys[K comparable, J comparable, V any](lazyMapping LazyMap[K, V], f func(K) J) LazyMap[J, V] {
	return NewLazyMap(func(yield func(J, V) bool) {
		for k, v := range lazyMapping {
			if !yield(f(k), v) {
				return
			}
		}
	})
}

// LazyKeys returns a [LazySlice] yielding all keys from the given [LazyMap].
func LazyKeys[K comparable, V any](lazyMapping LazyMap[K, V]) LazySlice[K] {
	return NewLazySlice(func(yield func(K) bool) {
		for k, _ := range lazyMapping {
			if !yield(k) {
				return
			}
		}
	})
}

// LazyValues returns a [LazySlice] yielding all values from the given [LazyMap].
func LazyValues[K comparable, V any](lazyMapping LazyMap[K, V]) LazySlice[V] {
	return NewLazySlice(func(yield func(V) bool) {
		for _, v := range lazyMapping {
			if !yield(v) {
				return
			}
		}
	})
}

// Take returns a [LazyMap] that yields at most n entries. If n<=0, yields nothing.
func (lazyMapping LazyMap[K, V]) Take(n int) LazyMap[K, V] {
	return NewLazyMap(func(yield func(K, V) bool) {
		if n <= 0 {
			return
		}

		count := 0

		for k, v := range lazyMapping {
			if !yield(k, v) {
				return
			}

			count++

			if count >= n {
				return
			}
		}
	})
}

// Skip returns a [LazyMap] that skips the first n entries.
func (lazyMapping LazyMap[K, V]) Skip(n int) LazyMap[K, V] {
	return NewLazyMap(func(yield func(K, V) bool) {
		skipped := 0

		for k, v := range lazyMapping {
			if skipped < n {
				skipped++

				continue
			}

			if !yield(k, v) {
				return
			}
		}
	})
}

// Unique returns a [LazyMap] that yields only the first entry for each key,
// deduplicating by key.
func (lazyMapping LazyMap[K, V]) Unique() LazyMap[K, V] {
	return NewLazyMap(func(yield func(K, V) bool) {
		seen := make(map[K]struct{})

		for k, v := range lazyMapping {
			if _, exists := seen[k]; exists {
				continue
			}

			seen[k] = struct{}{}

			if !yield(k, v) {
				return
			}
		}
	})
}

// Concat returns a [LazyMap] that yields all entries from this map
// followed by all entries from other.
func (lazyMapping LazyMap[K, V]) Concat(other LazyMap[K, V]) LazyMap[K, V] {
	return NewLazyMap(func(yield func(K, V) bool) {
		for k, v := range lazyMapping {
			if !yield(k, v) {
				return
			}
		}

		for k, v := range other {
			if !yield(k, v) {
				return
			}
		}
	})
}

// MarshalJSONTo materializes the iterator and encodes it as a JSON object into enc.
func (lazyMapping LazyMap[K, V]) MarshalJSONTo(enc *jsontext.Encoder) error {
	return json.MarshalEncode(enc, lazyMapping.Items())
}

// UnmarshalJSONFrom decodes a JSON object from dec into the lazy map.
func (lazyMapping *LazyMap[K, V]) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	var items map[K]V

	if err := json.UnmarshalDecode(dec, &items); err != nil {
		return err
	}

	*lazyMapping = NewLazyMap(maps.All(items))

	return nil
}

// MarshalJSON implements [encoding/json.Marshaler] for v1 compatibility.
// When used with [encoding/json/v2], [LazyMap.MarshalJSONTo] takes precedence.
func (lazyMapping LazyMap[K, V]) MarshalJSON() ([]byte, error) {
	return json.Marshal(lazyMapping.Items())
}

// UnmarshalJSON implements [encoding/json.Unmarshaler] for v1 compatibility.
// When used with [encoding/json/v2], [LazyMap.UnmarshalJSONFrom] takes precedence.
func (lazyMapping *LazyMap[K, V]) UnmarshalJSON(data []byte) error {
	var items map[K]V

	if err := json.Unmarshal(data, &items); err != nil {
		return err
	}

	*lazyMapping = NewLazyMap(maps.All(items))

	return nil
}
