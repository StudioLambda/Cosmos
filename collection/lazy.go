package collection

import (
	"encoding/json/jsontext"
	"encoding/json/v2"
	"iter"
	"slices"
)

// Lazy is a lazily-evaluated collection backed by an [iter.Seq].
// Being a function type, it can be used directly in a for-range loop:
//
//	for v := range lazyCollection { ... }
type Lazy[T any] func(yield func(T) bool)

// NewLazy creates a new [Lazy] from the given iterator function.
func NewLazy[T any](f func(yield func(T) bool)) Lazy[T] {
	return Lazy[T](f)
}

// Collection materializes the lazy collection into an eager [Collection].
func (lazyCollection Lazy[T]) Collection() Collection[T] {
	return New(lazyCollection.Slice())
}

// Slice materializes all items from the iterator into a slice.
func (lazyCollection Lazy[T]) Slice() []T {
	return slices.Collect(iter.Seq[T](lazyCollection))
}

// IsEmpty reports whether the iterator yields no items.
func (lazyCollection Lazy[T]) IsEmpty() bool {
	for range lazyCollection {
		return false
	}

	return true
}

// IsNotEmpty reports whether the iterator yields at least one item.
func (lazyCollection Lazy[T]) IsNotEmpty() bool {
	return !lazyCollection.IsEmpty()
}

// Len returns the total number of items yielded by the iterator.
// It fully consumes the sequence.
func (lazyCollection Lazy[T]) Len() int {
	count := 0

	for range lazyCollection {
		count++
	}

	return count
}

// Every reports whether f returns true for every item yielded by the iterator.
func (lazyCollection Lazy[T]) Every(f func(T) bool) bool {
	for v := range lazyCollection {
		if !f(v) {
			return false
		}
	}

	return true
}

// Each calls f for every item yielded by the iterator.
func (lazyCollection Lazy[T]) Each(f func(T)) {
	for v := range lazyCollection {
		f(v)
	}
}

// TapEach returns a lazy collection that calls f on each item as it is consumed.
func (lazyCollection Lazy[T]) TapEach(f func(T)) Lazy[T] {
	return NewLazy(func(yield func(T) bool) {
		for v := range lazyCollection {
			f(v)

			if !yield(v) {
				return
			}
		}
	})
}

// Filter returns a lazy collection containing only items for which f returns true.
func (lazyCollection Lazy[T]) Filter(f func(T) bool) Lazy[T] {
	return NewLazy(func(yield func(T) bool) {
		for v := range lazyCollection {
			if f(v) {
				if !yield(v) {
					return
				}
			}
		}
	})
}

// Reject returns a lazy collection containing only items for which f returns false.
func (lazyCollection Lazy[T]) Reject(f func(T) bool) Lazy[T] {
	return lazyCollection.Filter(func(v T) bool { return !f(v) })
}

// Map transforms each item using f and returns a new lazy collection of type K.
func (lazyCollection Lazy[T]) Map[K any](f func(T) K) Lazy[K] {
	return NewLazy(func(yield func(K) bool) {
		for v := range lazyCollection {
			if !yield(f(v)) {
				return
			}
		}
	})
}

// FirstWhere returns the first item for which f returns true, along with a boolean
// indicating whether such an item was found.
func (lazyCollection Lazy[T]) FirstWhere(f func(T) bool) (result T, ok bool) {
	for v := range lazyCollection {
		if f(v) {
			return v, true
		}
	}

	return result, false
}

// First returns the first item yielded by the iterator, along with a boolean
// indicating whether the iterator was non-empty.
func (lazyCollection Lazy[T]) First() (result T, ok bool) {
	for v := range lazyCollection {
		return v, true
	}

	return result, false
}

// Last returns the last item yielded by the iterator, along with a boolean
// indicating whether the iterator was non-empty.
// It fully consumes the sequence.
func (lazyCollection Lazy[T]) Last() (result T, ok bool) {
	for v := range lazyCollection {
		result = v
		ok = true
	}

	return result, ok
}

// Contains reports whether any item yielded by the iterator satisfies f.
func (lazyCollection Lazy[T]) Contains(f func(T) bool) bool {
	_, found := lazyCollection.FirstWhere(f)

	return found
}

// Take returns a lazy collection that yields at most limit items.
func (lazyCollection Lazy[T]) Take(limit int) Lazy[T] {
	return NewLazy(func(yield func(T) bool) {
		if limit <= 0 {
			return
		}

		count := 0

		for v := range lazyCollection {
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

// Skip returns a lazy collection that skips the first n items.
func (lazyCollection Lazy[T]) Skip(n int) Lazy[T] {
	return NewLazy(func(yield func(T) bool) {
		skipped := 0

		for v := range lazyCollection {
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

// TakeWhile returns a lazy collection that yields items until f first returns false.
func (lazyCollection Lazy[T]) TakeWhile(f func(T) bool) Lazy[T] {
	return NewLazy(func(yield func(T) bool) {
		for v := range lazyCollection {
			if !f(v) {
				return
			}

			if !yield(v) {
				return
			}
		}
	})
}

// SkipWhile returns a lazy collection that skips items until f first returns false,
// then yields all remaining items.
func (lazyCollection Lazy[T]) SkipWhile(f func(T) bool) Lazy[T] {
	return NewLazy(func(yield func(T) bool) {
		skipping := true

		for v := range lazyCollection {
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

// Chunk returns an iterator over consecutive slices of the given size.
// It panics if size is less than or equal to zero.
func (lazyCollection Lazy[T]) Chunk(size int) iter.Seq[[]T] {
	if size <= 0 {
		panic("chunk size must be greater than zero")
	}

	return func(yield func([]T) bool) {
		var chunk []T

		for v := range lazyCollection {
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
func (lazyCollection Lazy[T]) Sliding(size int, step ...int) iter.Seq[[]T] {
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

		for v := range lazyCollection {
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

// Reduce reduces the lazy collection to a single value of type K by applying f
// to each item starting from initial.
func (lazyCollection Lazy[T]) Reduce[K any](f func(K, T) K, initial K) K {
	acc := initial

	for v := range lazyCollection {
		acc = f(acc, v)
	}

	return acc
}

// Reverse materializes the iterator and returns a new lazy collection with
// items in reversed order.
func (lazyCollection Lazy[T]) Reverse() Lazy[T] {
	return NewLazy(func(yield func(T) bool) {
		items := slices.Collect(iter.Seq[T](lazyCollection))
		slices.Reverse(items)

		for _, v := range items {
			if !yield(v) {
				return
			}
		}
	})
}

// Sort materializes the iterator, sorts items using the given comparison function,
// and returns a new lazy collection over the sorted items.
func (lazyCollection Lazy[T]) Sort(cmp func(T, T) int) Lazy[T] {
	return NewLazy(func(yield func(T) bool) {
		items := slices.Collect(iter.Seq[T](lazyCollection))
		slices.SortFunc(items, cmp)

		for _, v := range items {
			if !yield(v) {
				return
			}
		}
	})
}

// Unique returns a lazy collection containing only the first occurrence of each
// item, as determined by the key function.
func (lazyCollection Lazy[T]) Unique[K comparable](key func(T) K) Lazy[T] {
	return NewLazy(func(yield func(T) bool) {
		seen := make(map[K]struct{})

		for v := range lazyCollection {
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

// Partition splits the iterator into two lazy collections: the first contains
// items for which f returns true, the second contains the rest.
// It fully consumes the sequence to allow both results to be independently iterated.
func (lazyCollection Lazy[T]) Partition(f func(T) bool) (Lazy[T], Lazy[T]) {
	var matching, rest []T

	for v := range lazyCollection {
		if f(v) {
			matching = append(matching, v)
		} else {
			rest = append(rest, v)
		}
	}

	return NewLazy(slices.Values(matching)), NewLazy(slices.Values(rest))
}

// Concat returns a lazy collection that yields all items from this collection
// followed by all items from other.
func (lazyCollection Lazy[T]) Concat(other Lazy[T]) Lazy[T] {
	return NewLazy(func(yield func(T) bool) {
		for v := range lazyCollection {
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
// map of sub-collections. It fully consumes the sequence.
func (lazyCollection Lazy[T]) GroupBy[K comparable](key func(T) K) map[K]Collection[T] {
	grouped := make(map[K][]T)

	for v := range lazyCollection {
		k := key(v)
		grouped[k] = append(grouped[k], v)
	}

	result := make(map[K]Collection[T], len(grouped))

	for k, items := range grouped {
		result[k] = New(items)
	}

	return result
}

// Nth returns a lazy collection consisting of every nth item, starting at the
// given offset. The offset defaults to zero if not provided.
// It panics if n is less than or equal to zero.
func (lazyCollection Lazy[T]) Nth(n int, offset ...int) Lazy[T] {
	if n <= 0 {
		panic("nth step must be greater than zero")
	}

	start := 0
	if len(offset) > 0 {
		start = offset[0]
	}

	return NewLazy(func(yield func(T) bool) {
		i := 0

		for v := range lazyCollection {
			if i >= start && (i-start)%n == 0 {
				if !yield(v) {
					return
				}
			}

			i++
		}
	})
}

// ForPage returns a lazy collection containing the items for the given
// one-indexed page number and page size.
func (lazyCollection Lazy[T]) ForPage(page, perPage int) Lazy[T] {
	return lazyCollection.Skip((page - 1) * perPage).Take(perPage)
}

// MarshalJSONTo materializes the iterator and encodes it as a JSON array into enc.
func (lazyCollection Lazy[T]) MarshalJSONTo(enc *jsontext.Encoder) error {
	return json.MarshalEncode(enc, lazyCollection.Slice())
}

// UnmarshalJSONFrom decodes a JSON array from dec into the lazy collection,
// backed by the resulting slice.
func (lazyCollection *Lazy[T]) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	var items []T

	if err := json.UnmarshalDecode(dec, &items); err != nil {
		return err
	}

	*lazyCollection = NewLazy(slices.Values(items))

	return nil
}

// MarshalJSON implements [encoding/json.Marshaler] for v1 compatibility.
// When used with [encoding/json/v2], [Lazy.MarshalJSONTo] takes precedence.
func (lazyCollection Lazy[T]) MarshalJSON() ([]byte, error) {
	return json.Marshal(lazyCollection.Slice())
}

// UnmarshalJSON implements [encoding/json.Unmarshaler] for v1 compatibility.
// When used with [encoding/json/v2], [Lazy.UnmarshalJSONFrom] takes precedence.
func (lazyCollection *Lazy[T]) UnmarshalJSON(data []byte) error {
	var items []T

	if err := json.Unmarshal(data, &items); err != nil {
		return err
	}

	*lazyCollection = NewLazy(slices.Values(items))

	return nil
}
