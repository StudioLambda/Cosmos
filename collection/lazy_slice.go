package collection

import (
	"encoding/json/jsontext"
	"encoding/json/v2"
	"iter"
	"slices"
)

// LazySlice is a lazily-evaluated collection backed by an [iter.Seq].
// Being a function type, it can be used directly in a for-range loop:
//
//	for v := range lazySlice { ... }
type LazySlice[T any] func(yield func(T) bool)

// NewLazySlice creates a new [LazySlice] from the given iterator function.
func NewLazySlice[T any](f func(yield func(T) bool)) LazySlice[T] {
	return LazySlice[T](f)
}

// Eager materializes the lazy slice into an eager [Slice].
func (lazySlice LazySlice[T]) Eager() Slice[T] {
	return NewSlice(lazySlice.Items())
}

// Items materializes all items from the iterator into a Go slice.
func (lazySlice LazySlice[T]) Items() []T {
	return slices.Collect(lazySlice.Iter())
}

func (lazySlice LazySlice[T]) Iter() iter.Seq[T] {
	return iter.Seq[T](lazySlice)
}

// IsEmpty reports whether the iterator yields no items.
func (lazySlice LazySlice[T]) IsEmpty() bool {
	for range lazySlice {
		return false
	}

	return true
}

// IsNotEmpty reports whether the iterator yields at least one item.
func (lazySlice LazySlice[T]) IsNotEmpty() bool {
	return !lazySlice.IsEmpty()
}

// Len returns the total number of items yielded by the iterator.
// It fully consumes the sequence.
func (lazySlice LazySlice[T]) Len() int {
	count := 0

	for range lazySlice {
		count++
	}

	return count
}

// Every reports whether f returns true for every item yielded by the iterator.
func (lazySlice LazySlice[T]) Every(f func(T) bool) bool {
	for v := range lazySlice {
		if !f(v) {
			return false
		}
	}

	return true
}

// Each calls f for every item yielded by the iterator.
func (lazySlice LazySlice[T]) Each(f func(T)) {
	for v := range lazySlice {
		f(v)
	}
}

// TapEach returns a lazy slice that calls f on each item as it is consumed.
func (lazySlice LazySlice[T]) TapEach(f func(T)) LazySlice[T] {
	return NewLazySlice(func(yield func(T) bool) {
		for v := range lazySlice {
			f(v)

			if !yield(v) {
				return
			}
		}
	})
}

// Filter returns a lazy slice containing only items for which f returns true.
func (lazySlice LazySlice[T]) Filter(f func(T) bool) LazySlice[T] {
	return NewLazySlice(func(yield func(T) bool) {
		for v := range lazySlice {
			if f(v) {
				if !yield(v) {
					return
				}
			}
		}
	})
}

// Reject returns a lazy slice containing only items for which f returns false.
func (lazySlice LazySlice[T]) Reject(f func(T) bool) LazySlice[T] {
	return lazySlice.Filter(func(v T) bool { return !f(v) })
}

// Map transforms each item using f and returns a new lazy slice of type K.
func (lazySlice LazySlice[T]) Map[K any](f func(T) K) LazySlice[K] {
	return NewLazySlice(func(yield func(K) bool) {
		for v := range lazySlice {
			if !yield(f(v)) {
				return
			}
		}
	})
}

// FlatMap transforms each item into zero or more items and yields the flattened
// results lazily.
func (lazySlice LazySlice[T]) FlatMap[K any](f func(T) []K) LazySlice[K] {
	return NewLazySlice(func(yield func(K) bool) {
		for v := range lazySlice {
			for _, mapped := range f(v) {
				if !yield(mapped) {
					return
				}
			}
		}
	})
}

// KeyBy indexes yielded items by the key returned from f. When multiple items
// map to the same key, the last item wins.
func (lazySlice LazySlice[T]) KeyBy[K comparable](f func(T) K) Map[K, T] {
	indexed := make(map[K]T)

	for v := range lazySlice {
		indexed[f(v)] = v
	}

	return NewMap(indexed)
}

// CountBy counts yielded items by the key returned from f.
func (lazySlice LazySlice[T]) CountBy[K comparable](f func(T) K) Map[K, int] {
	counts := make(map[K]int)

	for v := range lazySlice {
		counts[f(v)]++
	}

	return NewMap(counts)
}

// FirstWhere returns the first item for which f returns true, along with a boolean
// indicating whether such an item was found.
func (lazySlice LazySlice[T]) FirstWhere(f func(T) bool) (result T, ok bool) {
	for v := range lazySlice {
		if f(v) {
			return v, true
		}
	}

	return result, false
}

// First returns the first item yielded by the iterator, along with a boolean
// indicating whether the iterator was non-empty.
func (lazySlice LazySlice[T]) First() (result T, ok bool) {
	for v := range lazySlice {
		return v, true
	}

	return result, false
}

// Last returns the last item yielded by the iterator, along with a boolean
// indicating whether the iterator was non-empty.
// It fully consumes the sequence.
func (lazySlice LazySlice[T]) Last() (result T, ok bool) {
	for v := range lazySlice {
		result = v
		ok = true
	}

	return result, ok
}

// Contains reports whether any item yielded by the iterator satisfies f.
func (lazySlice LazySlice[T]) Contains(f func(T) bool) bool {
	_, found := lazySlice.FirstWhere(f)

	return found
}

// Take returns a lazy slice that yields at most limit items.
func (lazySlice LazySlice[T]) Take(limit int) LazySlice[T] {
	return NewLazySlice(func(yield func(T) bool) {
		if limit <= 0 {
			return
		}

		count := 0

		for v := range lazySlice {
			if !yield(v) {
				return
			}

			count++

			if count >= limit {
				return
			}
		}
	})
}

// Skip returns a lazy slice that skips the first n items.
func (lazySlice LazySlice[T]) Skip(n int) LazySlice[T] {
	return NewLazySlice(func(yield func(T) bool) {
		skipped := 0

		for v := range lazySlice {
			if skipped < n {
				skipped++

				continue
			}

			if !yield(v) {
				return
			}
		}
	})
}

// TakeWhile returns a lazy slice that yields items until f first returns false.
func (lazySlice LazySlice[T]) TakeWhile(f func(T) bool) LazySlice[T] {
	return NewLazySlice(func(yield func(T) bool) {
		for v := range lazySlice {
			if !f(v) {
				return
			}

			if !yield(v) {
				return
			}
		}
	})
}

// SkipWhile returns a lazy slice that skips items until f first returns false,
// then yields all remaining items.
func (lazySlice LazySlice[T]) SkipWhile(f func(T) bool) LazySlice[T] {
	return NewLazySlice(func(yield func(T) bool) {
		skipping := true

		for v := range lazySlice {
			if skipping && f(v) {
				continue
			}

			skipping = false

			if !yield(v) {
				return
			}
		}
	})
}

// TakeUntil returns a lazy slice that yields items until f first returns true.
func (lazySlice LazySlice[T]) TakeUntil(f func(T) bool) LazySlice[T] {
	return lazySlice.TakeWhile(func(v T) bool { return !f(v) })
}

// SkipUntil returns a lazy slice that skips items until f first returns true,
// then yields all remaining items.
func (lazySlice LazySlice[T]) SkipUntil(f func(T) bool) LazySlice[T] {
	return lazySlice.SkipWhile(func(v T) bool { return !f(v) })
}

// Chunk returns an iterator over consecutive slices of the given size.
// It panics if size is less than or equal to zero.
func (lazySlice LazySlice[T]) Chunk(size int) iter.Seq[[]T] {
	if size <= 0 {
		panic("chunk size must be greater than zero")
	}

	return func(yield func([]T) bool) {
		var chunk []T

		for v := range lazySlice {
			chunk = append(chunk, v)

			if len(chunk) == size {
				if !yield(chunk) {
					return
				}

				chunk = nil
			}
		}

		if len(chunk) > 0 {
			yield(chunk)
		}
	}
}

// Sliding returns an iterator over overlapping windows of the given size,
// advancing by step items between each window. The step defaults to one if
// not provided. Each yielded slice is a fresh copy.
// It panics if size or step is less than or equal to zero.
func (lazySlice LazySlice[T]) Sliding(size int, step ...int) iter.Seq[[]T] {
	if size <= 0 {
		panic("sliding window size must be greater than zero")
	}

	s := 1
	if len(step) > 0 {
		s = step[0]
	}

	if s <= 0 {
		panic("sliding window step must be greater than zero")
	}

	return func(yield func([]T) bool) {
		window := make([]T, 0, size)

		for v := range lazySlice {
			window = append(window, v)

			if len(window) == size {
				if !yield(slices.Clone(window)) {
					return
				}

				window = window[s:]
			}
		}
	}
}

// Reduce reduces the lazy slice to a single value of type K by applying f
// to each item starting from initial.
func (lazySlice LazySlice[T]) Reduce[K any](f func(K, T) K, initial K) K {
	acc := initial

	for v := range lazySlice {
		acc = f(acc, v)
	}

	return acc
}

// Reverse materializes the iterator and returns a new lazy slice with
// items in reversed order.
func (lazySlice LazySlice[T]) Reverse() LazySlice[T] {
	return NewLazySlice(func(yield func(T) bool) {
		items := slices.Collect(lazySlice.Iter())
		slices.Reverse(items)

		for _, v := range items {
			if !yield(v) {
				return
			}
		}
	})
}

// Sort materializes the iterator, sorts items using the given comparison function,
// and returns a new lazy slice over the sorted items.
func (lazySlice LazySlice[T]) Sort(cmp func(T, T) int) LazySlice[T] {
	return NewLazySlice(func(yield func(T) bool) {
		items := slices.Collect(lazySlice.Iter())
		slices.SortFunc(items, cmp)

		for _, v := range items {
			if !yield(v) {
				return
			}
		}
	})
}

// Unique returns a lazy slice containing only the first occurrence of each
// item, as determined by the key function.
func (lazySlice LazySlice[T]) Unique[K comparable](key func(T) K) LazySlice[T] {
	return NewLazySlice(func(yield func(T) bool) {
		seen := make(map[K]struct{})

		for v := range lazySlice {
			k := key(v)

			if _, exists := seen[k]; !exists {
				seen[k] = struct{}{}

				if !yield(v) {
					return
				}
			}
		}
	})
}

// Partition splits the iterator into two lazy slices: the first contains
// items for which f returns true, the second contains the rest.
// It fully consumes the sequence to allow both results to be independently iterated.
func (lazySlice LazySlice[T]) Partition(f func(T) bool) (LazySlice[T], LazySlice[T]) {
	var matching, rest []T

	for v := range lazySlice {
		if f(v) {
			matching = append(matching, v)
		} else {
			rest = append(rest, v)
		}
	}

	return NewLazySlice(slices.Values(matching)), NewLazySlice(slices.Values(rest))
}

// Concat returns a lazy slice that yields all items from this slice
// followed by all items from other.
func (lazySlice LazySlice[T]) Concat(other LazySlice[T]) LazySlice[T] {
	return NewLazySlice(func(yield func(T) bool) {
		for v := range lazySlice {
			if !yield(v) {
				return
			}
		}

		for v := range other {
			if !yield(v) {
				return
			}
		}
	})
}

// GroupBy groups all items by the key returned by the key function, returning a
// [Map] of sub-slices. It fully consumes the sequence.
func (lazySlice LazySlice[T]) GroupBy[K comparable](key func(T) K) Map[K, Slice[T]] {
	grouped := make(map[K][]T)

	for v := range lazySlice {
		k := key(v)
		grouped[k] = append(grouped[k], v)
	}

	result := make(map[K]Slice[T], len(grouped))

	for k, items := range grouped {
		result[k] = NewSlice(items)
	}

	return NewMap(result)
}

// Nth returns a lazy slice consisting of every nth item, starting at the
// given offset. The offset defaults to zero if not provided.
// It panics if n is less than or equal to zero.
func (lazySlice LazySlice[T]) Nth(n int, offset ...int) LazySlice[T] {
	if n <= 0 {
		panic("nth step must be greater than zero")
	}

	start := 0
	if len(offset) > 0 {
		start = offset[0]
	}

	return NewLazySlice(func(yield func(T) bool) {
		i := 0

		for v := range lazySlice {
			if i >= start && (i-start)%n == 0 {
				if !yield(v) {
					return
				}
			}

			i++
		}
	})
}

// ForPage returns a lazy slice containing the items for the given
// one-indexed page number and page size.
func (lazySlice LazySlice[T]) ForPage(page, perPage int) LazySlice[T] {
	return lazySlice.Skip((page - 1) * perPage).Take(perPage)
}

// MarshalJSONTo materializes the iterator and encodes it as a JSON array into enc.
func (lazySlice LazySlice[T]) MarshalJSONTo(enc *jsontext.Encoder) error {
	return json.MarshalEncode(enc, lazySlice.Items())
}

// UnmarshalJSONFrom decodes a JSON array from dec into the lazy slice,
// backed by the resulting slice.
func (lazySlice *LazySlice[T]) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	var items []T

	if err := json.UnmarshalDecode(dec, &items); err != nil {
		return err
	}

	*lazySlice = NewLazySlice(slices.Values(items))

	return nil
}

// MarshalJSON implements [encoding/json.Marshaler] for v1 compatibility.
// When used with [encoding/json/v2], [LazySlice.MarshalJSONTo] takes precedence.
func (lazySlice LazySlice[T]) MarshalJSON() ([]byte, error) {
	return json.Marshal(lazySlice.Items())
}

// UnmarshalJSON implements [encoding/json.Unmarshaler] for v1 compatibility.
// When used with [encoding/json/v2], [LazySlice.UnmarshalJSONFrom] takes precedence.
func (lazySlice *LazySlice[T]) UnmarshalJSON(data []byte) error {
	var items []T

	if err := json.Unmarshal(data, &items); err != nil {
		return err
	}

	*lazySlice = NewLazySlice(slices.Values(items))

	return nil
}
