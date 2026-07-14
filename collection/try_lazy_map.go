package collection

import (
	"encoding/json/jsontext"
	"encoding/json/v2"
	"errors"
	"iter"
)

// TryLazyMap is a lazily-evaluated map collection backed by an [iter.Seq2].
//
// The yielded value must be ignored whenever the yielded error is non-nil.
// Being a function type, it can be used directly in a for-range loop:
//
//	for entry, err := range tryLazyMap { ... }
type TryLazyMap[K comparable, V any] func(yield func(MapEntry[K, V], error) bool)

// NewTryLazyMap creates a new [TryLazyMap] from the given iterator function.
func NewTryLazyMap[K comparable, V any](f func(yield func(MapEntry[K, V], error) bool)) TryLazyMap[K, V] {
	return TryLazyMap[K, V](f)
}

// Eager materializes the lazy map into an eager [Map].
func (tryLazyMap TryLazyMap[K, V]) Eager() (Map[K, V], error) {
	items, err := tryLazyMap.Items()

	return NewMap(items), err
}

// EagerAll materializes the lazy map into an eager [Map], joining all errors.
func (tryLazyMap TryLazyMap[K, V]) EagerAll() (Map[K, V], error) {
	items, err := tryLazyMap.ItemsAll()

	return NewMap(items), err
}

// Items materializes all successful entries into a Go map.
func (tryLazyMap TryLazyMap[K, V]) Items() (map[K]V, error) {
	items := make(map[K]V)

	for entry, err := range tryLazyMap {
		if err != nil {
			return items, err
		}

		items[entry.Key] = entry.Value
	}

	return items, nil
}

// ItemsAll materializes all successful entries into a Go map, joining all errors.
func (tryLazyMap TryLazyMap[K, V]) ItemsAll() (map[K]V, error) {
	items := make(map[K]V)
	var errs []error

	for entry, err := range tryLazyMap {
		if err != nil {
			errs = append(errs, err)

			continue
		}

		items[entry.Key] = entry.Value
	}

	return items, errors.Join(errs...)
}

func (tryLazyMap TryLazyMap[K, V]) Iter() iter.Seq2[MapEntry[K, V], error] {
	return iter.Seq2[MapEntry[K, V], error](tryLazyMap)
}

// IsEmpty reports whether the iterator yields no successful entries.
func (tryLazyMap TryLazyMap[K, V]) IsEmpty() (bool, error) {
	for _, err := range tryLazyMap {
		if err != nil {
			return false, err
		}

		return false, nil
	}

	return true, nil
}

// IsNotEmpty reports whether the iterator yields at least one successful entry.
func (tryLazyMap TryLazyMap[K, V]) IsNotEmpty() (bool, error) {
	result, err := tryLazyMap.IsEmpty()

	return !result, err
}

// Len returns the total number of successful entries yielded by the iterator.
func (tryLazyMap TryLazyMap[K, V]) Len() (int, error) {
	count := 0

	for _, err := range tryLazyMap {
		if err != nil {
			return count, err
		}

		count++
	}

	return count, nil
}

// Every reports whether f returns true for every successful entry.
func (tryLazyMap TryLazyMap[K, V]) Every(f func(K, V) (bool, error)) (bool, error) {
	for entry, err := range tryLazyMap {
		if err != nil {
			return false, err
		}

		ok, err := f(entry.Key, entry.Value)
		if err != nil {
			return false, err
		}

		if !ok {
			return false, nil
		}
	}

	return true, nil
}

// Each calls f for every successful entry.
func (tryLazyMap TryLazyMap[K, V]) Each(f func(K, V) error) error {
	for entry, err := range tryLazyMap {
		if err != nil {
			return err
		}

		if err := f(entry.Key, entry.Value); err != nil {
			return err
		}
	}

	return nil
}

// EachAll calls f for every successful entry, joining all errors.
func (tryLazyMap TryLazyMap[K, V]) EachAll(f func(K, V) error) error {
	var errs []error

	for entry, err := range tryLazyMap {
		if err != nil {
			errs = append(errs, err)

			continue
		}

		if err := f(entry.Key, entry.Value); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

// TapEach returns a lazy map that calls f on each successful entry as it is consumed.
func (tryLazyMap TryLazyMap[K, V]) TapEach(f func(K, V) error) TryLazyMap[K, V] {
	return NewTryLazyMap(func(yield func(MapEntry[K, V], error) bool) {
		for entry, err := range tryLazyMap {
			if err != nil {
				if !tryYieldMapError(yield, err) {
					return
				}

				continue
			}

			if err := f(entry.Key, entry.Value); err != nil {
				if !tryYieldMapError(yield, err) {
					return
				}

				continue
			}

			if !yield(entry, nil) {
				return
			}
		}
	})
}

// Contains reports whether any successful entry satisfies f.
func (tryLazyMap TryLazyMap[K, V]) Contains(f func(K, V) (bool, error)) (bool, error) {
	for entry, err := range tryLazyMap {
		if err != nil {
			return false, err
		}

		ok, err := f(entry.Key, entry.Value)
		if err != nil {
			return false, err
		}

		if ok {
			return true, nil
		}
	}

	return false, nil
}

// HasAny reports whether any of the given keys exists in the lazy map.
func (tryLazyMap TryLazyMap[K, V]) HasAny(keys ...K) (bool, error) {
	if len(keys) == 0 {
		return false, nil
	}

	wanted := make(map[K]struct{}, len(keys))

	for _, key := range keys {
		wanted[key] = struct{}{}
	}

	for entry, err := range tryLazyMap {
		if err != nil {
			return false, err
		}

		if _, ok := wanted[entry.Key]; ok {
			return true, nil
		}
	}

	return false, nil
}

// Filter returns a lazy map containing only successful entries for which f returns true.
func (tryLazyMap TryLazyMap[K, V]) Filter(f func(K, V) (bool, error)) TryLazyMap[K, V] {
	return NewTryLazyMap(func(yield func(MapEntry[K, V], error) bool) {
		for entry, err := range tryLazyMap {
			if err != nil {
				if !tryYieldMapError(yield, err) {
					return
				}

				continue
			}

			ok, err := f(entry.Key, entry.Value)
			if err != nil {
				if !tryYieldMapError(yield, err) {
					return
				}

				continue
			}

			if ok {
				if !yield(entry, nil) {
					return
				}
			}
		}
	})
}

// Reject returns a lazy map containing only successful entries for which f returns false.
func (tryLazyMap TryLazyMap[K, V]) Reject(f func(K, V) (bool, error)) TryLazyMap[K, V] {
	return tryLazyMap.Filter(func(k K, v V) (bool, error) {
		ok, err := f(k, v)
		if err != nil {
			return false, err
		}

		return !ok, nil
	})
}

// Only returns a lazy map containing only the entries whose keys were given.
func (tryLazyMap TryLazyMap[K, V]) Only(keys ...K) TryLazyMap[K, V] {
	wanted := make(map[K]struct{}, len(keys))

	for _, key := range keys {
		wanted[key] = struct{}{}
	}

	return tryLazyMap.Filter(func(k K, v V) (bool, error) {
		_, ok := wanted[k]

		return ok, nil
	})
}

// Except returns a lazy map containing all entries except those whose keys were given.
func (tryLazyMap TryLazyMap[K, V]) Except(keys ...K) TryLazyMap[K, V] {
	blocked := make(map[K]struct{}, len(keys))

	for _, key := range keys {
		blocked[key] = struct{}{}
	}

	return tryLazyMap.Filter(func(k K, v V) (bool, error) {
		_, ok := blocked[k]

		return !ok, nil
	})
}

// MapValues transforms values using f and returns a new lazy map with the same keys.
func (tryLazyMap TryLazyMap[K, V]) MapValues[W any](f func(K, V) (W, error)) TryLazyMap[K, W] {
	return NewTryLazyMap(func(yield func(MapEntry[K, W], error) bool) {
		for entry, err := range tryLazyMap {
			if err != nil {
				if !tryYieldMapError(yield, err) {
					return
				}

				continue
			}

			w, err := f(entry.Key, entry.Value)
			if err != nil {
				if !tryYieldMapError(yield, err) {
					return
				}

				continue
			}

			if !yield(NewMapEntry(entry.Key, w), nil) {
				return
			}
		}
	})
}

// MapKeys transforms keys using f and returns a new lazy map with the same values.
func (tryLazyMap TryLazyMap[K, V]) MapKeys[J comparable](f func(K, V) (J, error)) TryLazyMap[J, V] {
	return NewTryLazyMap(func(yield func(MapEntry[J, V], error) bool) {
		for entry, err := range tryLazyMap {
			if err != nil {
				if !tryYieldMapError(yield, err) {
					return
				}

				continue
			}

			j, err := f(entry.Key, entry.Value)
			if err != nil {
				if !tryYieldMapError(yield, err) {
					return
				}

				continue
			}

			if !yield(NewMapEntry(j, entry.Value), nil) {
				return
			}
		}
	})
}

// LazyKeys returns a [TryLazySlice] yielding all keys.
func (tryLazyMap TryLazyMap[K, V]) LazyKeys() iter.Seq2[K, error] {
	return func(yield func(K, error) bool) {
		for entry, err := range tryLazyMap {
			if err != nil {
				if !tryYieldError(yield, err) {
					return
				}

				continue
			}

			if !yield(entry.Key, nil) {
				return
			}
		}
	}
}

// LazyValues returns a [TryLazySlice] yielding all values.
func (tryLazyMap TryLazyMap[K, V]) LazyValues() iter.Seq2[V, error] {
	return func(yield func(V, error) bool) {
		for entry, err := range tryLazyMap {
			if err != nil {
				if !tryYieldError(yield, err) {
					return
				}

				continue
			}

			if !yield(entry.Value, nil) {
				return
			}
		}
	}
}

// Take returns a lazy map that yields at most n entries.
func (tryLazyMap TryLazyMap[K, V]) Take(n int) TryLazyMap[K, V] {
	return NewTryLazyMap(func(yield func(MapEntry[K, V], error) bool) {
		if n <= 0 {
			return
		}

		count := 0

		for entry, err := range tryLazyMap {
			if err != nil {
				if !tryYieldMapError(yield, err) {
					return
				}

				continue
			}

			if !yield(entry, nil) {
				return
			}

			count++

			if count >= n {
				return
			}
		}
	})
}

// Skip returns a lazy map that skips the first n entries.
func (tryLazyMap TryLazyMap[K, V]) Skip(n int) TryLazyMap[K, V] {
	return NewTryLazyMap(func(yield func(MapEntry[K, V], error) bool) {
		skipped := 0

		for entry, err := range tryLazyMap {
			if err != nil {
				if !tryYieldMapError(yield, err) {
					return
				}

				continue
			}

			if skipped < n {
				skipped++

				continue
			}

			if !yield(entry, nil) {
				return
			}
		}
	})
}

// Unique returns a lazy map that yields only the first entry for each key.
func (tryLazyMap TryLazyMap[K, V]) Unique() TryLazyMap[K, V] {
	return NewTryLazyMap(func(yield func(MapEntry[K, V], error) bool) {
		seen := make(map[K]struct{})

		for entry, err := range tryLazyMap {
			if err != nil {
				if !tryYieldMapError(yield, err) {
					return
				}

				continue
			}

			if _, exists := seen[entry.Key]; exists {
				continue
			}

			seen[entry.Key] = struct{}{}

			if !yield(entry, nil) {
				return
			}
		}
	})
}

// Concat returns a lazy map that yields all entries from this map followed by other.
func (tryLazyMap TryLazyMap[K, V]) Concat(other TryLazyMap[K, V]) TryLazyMap[K, V] {
	return NewTryLazyMap(func(yield func(MapEntry[K, V], error) bool) {
		for entry, err := range tryLazyMap {
			if !yield(entry, err) {
				return
			}
		}

		for entry, err := range other {
			if !yield(entry, err) {
				return
			}
		}
	})
}

// MarshalJSONTo materializes the iterator and encodes it as a JSON object into enc.
func (tryLazyMap TryLazyMap[K, V]) MarshalJSONTo(enc *jsontext.Encoder) error {
	items, err := tryLazyMap.Items()
	if err != nil {
		return err
	}

	return json.MarshalEncode(enc, items)
}

// UnmarshalJSONFrom decodes a JSON object from dec into the lazy map.
func (tryLazyMap *TryLazyMap[K, V]) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	var items map[K]V

	if err := json.UnmarshalDecode(dec, &items); err != nil {
		return err
	}

	*tryLazyMap = NewMap(items).Try()

	return nil
}

// MarshalJSON implements [encoding/json.Marshaler] for v1 compatibility.
func (tryLazyMap TryLazyMap[K, V]) MarshalJSON() ([]byte, error) {
	items, err := tryLazyMap.Items()
	if err != nil {
		return nil, err
	}

	return json.Marshal(items)
}

// UnmarshalJSON implements [encoding/json.Unmarshaler] for v1 compatibility.
func (tryLazyMap *TryLazyMap[K, V]) UnmarshalJSON(data []byte) error {
	var items map[K]V

	if err := json.Unmarshal(data, &items); err != nil {
		return err
	}

	*tryLazyMap = NewMap(items).Try()

	return nil
}

// tryYieldMapError yields err with the zero value of MapEntry.
func tryYieldMapError[K comparable, V any](yield func(MapEntry[K, V], error) bool, err error) bool {
	var zero MapEntry[K, V]

	return yield(zero, err)
}
