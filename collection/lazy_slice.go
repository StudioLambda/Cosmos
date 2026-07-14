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
//	for i, v := range lazySlice { ... }
type LazySlice[T any] func(yield func(int, T) bool)

// NewLazySlice creates a new [LazySlice] from the given iterator function.
func NewLazySlice[T any](f func(yield func(int, T) bool)) LazySlice[T] {
	return LazySlice[T](f)
}

// Eager materializes the lazy slice into an eager [Slice].
func (lazySlice LazySlice[T]) Eager() Slice[T] {
	return NewSlice(lazySlice.Items())
}

// Items materializes all items from the iterator into a Go slice.
func (lazySlice LazySlice[T]) Items() []T {
	var items []T
	for _, v := range lazySlice {
		items = append(items, v)
	}
	return items
}

func (lazySlice LazySlice[T]) Iter() iter.Seq2[int, T] {
	return iter.Seq2[int, T](lazySlice)
}

// Try returns a [TryLazySlice] backed by the items in the lazy slice.
func (lazySlice LazySlice[T]) Try() TryLazySlice[T] {
	return NewTryLazySlice(func(yield func(T, error) bool) {
		for _, v := range lazySlice {
			if !yield(v, nil) {
				return
			}
		}
	})
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
func (lazySlice LazySlice[T]) Every(f func(int, T) bool) bool {
	for i, v := range lazySlice {
		if !f(i, v) {
			return false
		}
	}

	return true
}

// Each calls f for every item yielded by the iterator.
func (lazySlice LazySlice[T]) Each(f func(int, T)) {
	for i, v := range lazySlice {
		f(i, v)
	}
}

// TapEach returns a lazy slice that calls f on each item as it is consumed.
func (lazySlice LazySlice[T]) TapEach(f func(int, T)) LazySlice[T] {
	return NewLazySlice(func(yield func(int, T) bool) {
		index := 0

		for i, v := range lazySlice {
			f(i, v)

			if !yield(index, v) {
				return
			}

			index++
		}
	})
}

// Filter returns a lazy slice containing only items for which f returns true.
func (lazySlice LazySlice[T]) Filter(f func(int, T) bool) LazySlice[T] {
	return NewLazySlice(func(yield func(int, T) bool) {
		index := 0

		for i, v := range lazySlice {
			if f(i, v) {
				if !yield(index, v) {
					return
				}
			}

			index++
		}
	})
}

// Reject returns a lazy slice containing only items for which f returns false.
func (lazySlice LazySlice[T]) Reject(f func(int, T) bool) LazySlice[T] {
	return lazySlice.Filter(func(i int, v T) bool { return !f(i, v) })
}

// Map transforms each item using f and returns a new lazy slice of type K.
func (lazySlice LazySlice[T]) Map[K any](f func(int, T) K) LazySlice[K] {
	return NewLazySlice(func(yield func(int, K) bool) {
		index := 0

		for i, v := range lazySlice {
			if !yield(index, f(i, v)) {
				return
			}

			index++
		}
	})
}

// FlatMap transforms each item into zero or more items and yields the flattened
// results lazily.
func (lazySlice LazySlice[T]) FlatMap[K any](f func(int, T) []K) LazySlice[K] {
	return NewLazySlice(func(yield func(int, K) bool) {
		index := 0

		for i, v := range lazySlice {
			for _, mapped := range f(i, v) {
				if !yield(index, mapped) {
					return
				}
			}

			index++
		}
	})
}

// KeyBy indexes yielded items by the key returned from f. When multiple items
// map to the same key, the last item wins.
func (lazySlice LazySlice[T]) KeyBy[K comparable](f func(int, T) K) Map[K, T] {
	indexed := make(map[K]T)

	for i, v := range lazySlice {
		indexed[f(i, v)] = v
	}

	return NewMap(indexed)
}

// CountBy counts yielded items by the key returned from f.
func (lazySlice LazySlice[T]) CountBy[K comparable](f func(int, T) K) Map[K, int] {
	counts := make(map[K]int)

	for i, v := range lazySlice {
		counts[f(i, v)]++
	}

	return NewMap(counts)
}

// FirstWhere returns the first item for which f returns true, along with a boolean
// indicating whether such an item was found.
func (lazySlice LazySlice[T]) FirstWhere(f func(int, T) bool) (result T, ok bool) {
	for i, v := range lazySlice {
		if f(i, v) {
			return v, true
		}
	}

	return result, false
}

// First returns the first item yielded by the iterator, along with a boolean
// indicating whether the iterator was non-empty.
func (lazySlice LazySlice[T]) First() (result T, ok bool) {
	for _, v := range lazySlice {
		return v, true
	}

	return result, false
}

// Last returns the last item yielded by the iterator, along with a boolean
// indicating whether the iterator was non-empty.
// It fully consumes the sequence.
func (lazySlice LazySlice[T]) Last() (result T, ok bool) {
	for _, v := range lazySlice {
		result = v
		ok = true
	}

	return result, ok
}

// Contains reports whether any item yielded by the iterator satisfies f.
func (lazySlice LazySlice[T]) Contains(f func(int, T) bool) bool {
	_, found := lazySlice.FirstWhere(f)

	return found
}

// Take returns a lazy slice that yields at most limit items.
func (lazySlice LazySlice[T]) Take(limit int) LazySlice[T] {
	return NewLazySlice(func(yield func(int, T) bool) {
		if limit <= 0 {
			return
		}

		count := 0

		index := 0

		for _, v := range lazySlice {
			if !yield(index, v) {
				return
			}

			count++

			if count >= limit {
				return
			}

			index++
		}
	})
}

// Skip returns a lazy slice that skips the first n items.
func (lazySlice LazySlice[T]) Skip(n int) LazySlice[T] {
	return NewLazySlice(func(yield func(int, T) bool) {
		skipped := 0

		index := 0

		for _, v := range lazySlice {
			if skipped < n {
				skipped++

				continue
			}

			if !yield(index, v) {
				return
			}

			index++
		}
	})
}

// TakeWhile returns a lazy slice that yields items until f first returns false.
func (lazySlice LazySlice[T]) TakeWhile(f func(int, T) bool) LazySlice[T] {
	return NewLazySlice(func(yield func(int, T) bool) {
		index := 0

		for i, v := range lazySlice {
			if !f(i, v) {
				return
			}

			if !yield(index, v) {
				return
			}

			index++
		}
	})
}

// SkipWhile returns a lazy slice that skips items until f first returns false,
// then yields all remaining items.
func (lazySlice LazySlice[T]) SkipWhile(f func(int, T) bool) LazySlice[T] {
	return NewLazySlice(func(yield func(int, T) bool) {
		skipping := true

		index := 0

		for i, v := range lazySlice {
			if skipping && f(i, v) {
				continue
			}

			skipping = false

			if !yield(index, v) {
				return
			}

			index++
		}
	})
}

// TakeUntil returns a lazy slice that yields items until f first returns true.
func (lazySlice LazySlice[T]) TakeUntil(f func(int, T) bool) LazySlice[T] {
	return lazySlice.TakeWhile(func(i int, v T) bool { return !f(i, v) })
}

// SkipUntil returns a lazy slice that skips items until f first returns true,
// then yields all remaining items.
func (lazySlice LazySlice[T]) SkipUntil(f func(int, T) bool) LazySlice[T] {
	return lazySlice.SkipWhile(func(i int, v T) bool { return !f(i, v) })
}

// Chunk returns an iterator over consecutive slices of the given size.
// It panics if size is less than or equal to zero.
func (lazySlice LazySlice[T]) Chunk(size int) iter.Seq2[int, []T] {
	if size <= 0 {
		panic("chunk size must be greater than zero")
	}

	return func(yield func(int, []T) bool) {
		var chunk []T

		index := 0

		for _, v := range lazySlice {
			chunk = append(chunk, v)

			if len(chunk) == size {
				if !yield(index, chunk) {
					return
				}

				chunk = nil
			}
		}

		if len(chunk) > 0 {
			yield(index, chunk)
		}
	}
}

// Sliding returns an iterator over overlapping windows of the given size,
// advancing by step items between each window. The step defaults to one if
// not provided. Each yielded slice is a fresh copy.
// It panics if size or step is less than or equal to zero.
func (lazySlice LazySlice[T]) Sliding(size int, step ...int) iter.Seq2[int, []T] {
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

	return func(yield func(int, []T) bool) {
		window := make([]T, 0, size)

		index := 0

		for _, v := range lazySlice {
			window = append(window, v)

			if len(window) == size {
				if !yield(index, slices.Clone(window)) {
					return
				}

				window = window[s:]
			}
		}
	}
}

// Reduce reduces the lazy slice to a single value of type K by applying f
// to each item starting from initial.
func (lazySlice LazySlice[T]) Reduce[K any](f func(K, int, T) K, initial K) K {
	acc := initial

	for i, v := range lazySlice {
		acc = f(acc, i, v)
	}

	return acc
}

// Reverse materializes the iterator and returns a new lazy slice with
// items in reversed order.
func (lazySlice LazySlice[T]) Reverse() LazySlice[T] {
	return NewLazySlice(func(yield func(int, T) bool) {
		index := 0
		items := lazySlice.Items()
		slices.Reverse(items)

		for _, v := range items {
			if !yield(index, v) {
				return
			}

			index++
		}
	})
}

// Sort materializes the iterator, sorts items using the given comparison function,
// and returns a new lazy slice over the sorted items.
func (lazySlice LazySlice[T]) Sort(cmp func(T, T) int) LazySlice[T] {
	return NewLazySlice(func(yield func(int, T) bool) {
		index := 0
		items := lazySlice.Items()
		slices.SortFunc(items, cmp)

		for _, v := range items {
			if !yield(index, v) {
				return
			}

			index++
		}
	})
}

// Unique returns a lazy slice containing only the first occurrence of each
// item, as determined by the key function.
func (lazySlice LazySlice[T]) Unique[K comparable](key func(int, T) K) LazySlice[T] {
	return NewLazySlice(func(yield func(int, T) bool) {
		seen := make(map[K]struct{})

		index := 0

		for i, v := range lazySlice {
			k := key(i, v)

			if _, exists := seen[k]; !exists {
				seen[k] = struct{}{}

				if !yield(index, v) {
					return
				}
			}

			index++
		}
	})
}

// Partition splits the iterator into two lazy slices: the first contains
// items for which f returns true, the second contains the rest.
// It fully consumes the sequence to allow both results to be independently iterated.
func (lazySlice LazySlice[T]) Partition(f func(int, T) bool) (LazySlice[T], LazySlice[T]) {
	var matching, rest []T

	for i, v := range lazySlice {
		if f(i, v) {
			matching = append(matching, v)
		} else {
			rest = append(rest, v)
		}
	}

	return NewLazySlice(slices.All(matching)), NewLazySlice(slices.All(rest))
}

// Concat returns a lazy slice that yields all items from this slice
// followed by all items from other.
func (lazySlice LazySlice[T]) Concat(other LazySlice[T]) LazySlice[T] {
	return NewLazySlice(func(yield func(int, T) bool) {
		index := 0

		for _, v := range lazySlice {
			if !yield(index, v) {
				return
			}
		}

		for _, v := range other {
			if !yield(index, v) {
				return
			}

			index++
		}
	})
}

// GroupBy groups all items by the key returned by the key function, returning a
// [Map] of sub-slices. It fully consumes the sequence.
func (lazySlice LazySlice[T]) GroupBy[K comparable](key func(int, T) K) Map[K, Slice[T]] {
	grouped := make(map[K][]T)

	for i, v := range lazySlice {
		k := key(i, v)
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

	return NewLazySlice(func(yield func(int, T) bool) {
		stepIdx := 0

		index := 0

		for _, v := range lazySlice {
			if stepIdx >= start && (stepIdx-start)%n == 0 {
				if !yield(index, v) {
					return
				}
			}

			stepIdx++
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

	*lazySlice = NewLazySlice(slices.All(items))

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

	*lazySlice = NewLazySlice(slices.All(items))

	return nil
}
