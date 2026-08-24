package collection

import (
	"encoding/json/jsontext"
	"encoding/json/v2"
	"iter"
	"slices"
)

// Slice is an eager, in-memory collection of items of type T.
type Slice[T any] []T

// NewSlice creates a new [Slice] from a plain Go slice of items of type T.
func NewSlice[T any](items []T) Slice[T] {
	return Slice[T](items)
}

// NewSliceAs creates a new [Slice] from the given slice type, preserving
// compatibility with named slice types whose underlying type is []T.
func NewSliceAs[S ~[]T, T any](items S) Slice[T] {
	return NewSlice([]T(items))
}

// Iter returns an [iter.Seq2] over the items in the slice, suitable for
// use in a for-range loop.
//
//	for i, v := range slice.Iter() { ... }
func (slice Slice[T]) Iter() iter.Seq2[int, T] {
	return slices.All([]T(slice))
}

// Lazy returns a [LazySlice] backed by the items in the slice.
func (slice Slice[T]) Lazy() LazySlice[T] {
	return NewLazySlice(slice.Iter())
}

// Try returns a [TryLazySlice] backed by the items in the slice.
func (slice Slice[T]) Try() TryLazySlice[T] {
	return NewTryLazySlice(func(yield func(T, error) bool) {
		for _, v := range slice {
			if !yield(v, nil) {
				return
			}
		}
	})
}

// Items returns the underlying slice of items.
func (slice Slice[T]) Items() []T {
	return []T(slice)
}

// ItemsAs returns the underlying items as the given slice type.
func (slice Slice[T]) ItemsAs[S ~[]T]() S {
	return S(slice)
}

// IsEmpty reports whether the slice contains no items.
func (slice Slice[T]) IsEmpty() bool {
	return len(slice) == 0
}

// IsNotEmpty reports whether the slice contains at least one item.
func (slice Slice[T]) IsNotEmpty() bool {
	return len(slice) > 0
}

// Len returns the number of items in the slice.
func (slice Slice[T]) Len() int {
	return len(slice)
}

// Every reports whether f returns true for every item in the slice.
func (slice Slice[T]) Every(f func(int, T) bool) bool {
	for i, v := range slice {
		if !f(i, v) {
			return false
		}
	}

	return true
}

// Each calls f for every item in the slice.
func (slice Slice[T]) Each(f func(int, T)) {
	for i, v := range slice {
		f(i, v)
	}
}

// TapEach calls f for every item and returns the slice unchanged.
func (slice Slice[T]) TapEach(f func(int, T)) Slice[T] {
	for i, v := range slice {
		f(i, v)
	}

	return slice
}

// Filter returns a new slice containing only items for which f returns true.
func (slice Slice[T]) Filter(f func(int, T) bool) Slice[T] {
	filtered := make([]T, 0, len(slice))

	for i, v := range slice {
		if f(i, v) {
			filtered = append(filtered, v)
		}
	}

	return NewSlice(filtered)
}

// Reject returns a new slice containing only items for which f returns false.
func (slice Slice[T]) Reject(f func(int, T) bool) Slice[T] {
	return slice.Filter(func(i int, v T) bool { return !f(i, v) })
}

// Map transforms each item using f and returns a new slice of type K.
func (slice Slice[T]) Map[K any](f func(int, T) K) Slice[K] {
	mapped := make([]K, len(slice))

	for i, v := range slice {
		mapped[i] = f(i, v)
	}

	return NewSlice(mapped)
}

// FlatMap transforms each item into zero or more items and flattens the
// results into a single slice of type K.
func (slice Slice[T]) FlatMap[K any](f func(int, T) []K) Slice[K] {
	flattened := make([]K, 0, len(slice))

	for i, v := range slice {
		flattened = append(flattened, f(i, v)...)
	}

	return NewSlice(flattened)
}

// KeyBy indexes items by the key returned from f. When multiple items map to
// the same key, the last item wins.
func (slice Slice[T]) KeyBy[K comparable](f func(int, T) K) Map[K, T] {
	indexed := make(map[K]T, len(slice))

	for i, v := range slice {
		indexed[f(i, v)] = v
	}

	return NewMap(indexed)
}

// CountBy counts items by the key returned from f.
func (slice Slice[T]) CountBy[K comparable](f func(int, T) K) Map[K, int] {
	counts := make(map[K]int, len(slice))

	for i, v := range slice {
		counts[f(i, v)]++
	}

	return NewMap(counts)
}

// FirstWhere returns the first item for which f returns true, along with a boolean
// indicating whether such an item was found.
func (slice Slice[T]) FirstWhere(f func(int, T) bool) (result T, ok bool) {
	for i, v := range slice {
		if f(i, v) {
			return v, true
		}
	}

	return result, false
}

// First returns the first item in the slice, along with a boolean
// indicating whether the slice is non-empty.
func (slice Slice[T]) First() (result T, ok bool) {
	if len(slice) == 0 {
		return result, false
	}

	return slice[0], true
}

// Last returns the last item in the slice, along with a boolean
// indicating whether the slice is non-empty.
func (slice Slice[T]) Last() (result T, ok bool) {
	if len(slice) == 0 {
		return result, false
	}

	return slice[len(slice)-1], true
}

// Contains reports whether any item in the slice satisfies f.
func (slice Slice[T]) Contains(f func(int, T) bool) bool {
	_, found := slice.FirstWhere(f)

	return found
}

// Take returns a new slice with at most limit items from the start.
func (slice Slice[T]) Take(limit int) Slice[T] {
	limit = max(0, min(limit, len(slice)))

	return NewSlice(slice[:limit])
}

// Skip returns a new slice with the first n items removed.
func (slice Slice[T]) Skip(n int) Slice[T] {
	n = max(0, min(n, len(slice)))

	return NewSlice(slice[n:])
}

// TakeWhile returns a new slice with items from the start until f first
// returns false.
func (slice Slice[T]) TakeWhile(f func(int, T) bool) Slice[T] {
	taken := make([]T, 0, len(slice))

	for i, v := range slice {
		if !f(i, v) {
			break
		}

		taken = append(taken, v)
	}

	return NewSlice(taken)
}

// SkipWhile returns a new slice with leading items removed until f first
// returns false.
func (slice Slice[T]) SkipWhile(f func(int, T) bool) Slice[T] {
	i := 0

	for i < len(slice) && f(i, slice[i]) {
		i++
	}

	return NewSlice(slice[i:])
}

// TakeUntil returns a new slice with items from the start until f first
// returns true.
func (slice Slice[T]) TakeUntil(f func(int, T) bool) Slice[T] {
	return slice.TakeWhile(func(i int, v T) bool { return !f(i, v) })
}

// SkipUntil returns a new slice with leading items removed until f first
// returns true.
func (slice Slice[T]) SkipUntil(f func(int, T) bool) Slice[T] {
	return slice.SkipWhile(func(i int, v T) bool { return !f(i, v) })
}

// Chunk splits the slice into consecutive sub-slices of the given size.
// It panics if size is less than or equal to zero.
func (slice Slice[T]) Chunk(size int) []Slice[T] {
	if size <= 0 {
		panic("chunk size must be greater than zero")
	}

	var chunks []Slice[T]

	for i := 0; i < len(slice); i += size {
		end := min(i+size, len(slice))
		chunks = append(chunks, NewSlice(slice[i:end]))
	}

	return chunks
}

// Sliding returns a slice of overlapping windows of the given size, advancing
// by step items between each window. The step defaults to one if not provided.
// It panics if size or step is less than or equal to zero.
func (slice Slice[T]) Sliding(size int, step ...int) []Slice[T] {
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

	var windows []Slice[T]

	for i := 0; i+size <= len(slice); i += s {
		windows = append(windows, NewSlice(slice[i:i+size]))
	}

	return windows
}

// Reduce reduces the slice to a single value of type K by applying f
// to each item starting from initial.
func (slice Slice[T]) Reduce[K any](f func(K, int, T) K, initial K) K {
	acc := initial

	for i, v := range slice {
		acc = f(acc, i, v)
	}

	return acc
}

// Reverse returns a new slice with items in reversed order.
func (slice Slice[T]) Reverse() Slice[T] {
	reversed := slices.Clone([]T(slice))
	slices.Reverse(reversed)

	return NewSlice(reversed)
}

// Sort returns a new slice with items sorted using the given comparison function.
func (slice Slice[T]) Sort(cmp func(T, T) int) Slice[T] {
	sorted := slices.Clone([]T(slice))
	slices.SortFunc(sorted, cmp)

	return NewSlice(sorted)
}

// Unique returns a new slice containing only the first occurrence of each
// item, as determined by the key function.
func (slice Slice[T]) Unique[K comparable](key func(int, T) K) Slice[T] {
	seen := make(map[K]struct{}, len(slice))
	unique := make([]T, 0, len(slice))

	for i, v := range slice {
		k := key(i, v)

		if _, exists := seen[k]; !exists {
			seen[k] = struct{}{}
			unique = append(unique, v)
		}
	}

	return NewSlice(unique)
}

// Partition splits the slice into two: the first contains items for which
// f returns true, the second contains the rest.
func (slice Slice[T]) Partition(f func(int, T) bool) (Slice[T], Slice[T]) {
	matching := make([]T, 0, len(slice))
	rest := make([]T, 0, len(slice))

	for i, v := range slice {
		if f(i, v) {
			matching = append(matching, v)
		} else {
			rest = append(rest, v)
		}
	}

	return NewSlice(matching), NewSlice(rest)
}

// Concat returns a new slice with the items of other appended.
func (slice Slice[T]) Concat(other Slice[T]) Slice[T] {
	return NewSlice(append(slices.Clone([]T(slice)), other...))
}

// GroupBy groups items by the key returned by the key function, returning a
// [Map] of sub-slices.
func (slice Slice[T]) GroupBy[K comparable](key func(int, T) K) Map[K, Slice[T]] {
	grouped := make(map[K][]T, len(slice))

	for i, v := range slice {
		k := key(i, v)
		grouped[k] = append(grouped[k], v)
	}

	result := make(map[K]Slice[T], len(grouped))

	for k, items := range grouped {
		result[k] = NewSlice(items)
	}

	return NewMap(result)
}

// Nth returns a new slice consisting of every nth item, starting at the
// given offset. The offset defaults to zero if not provided.
// It panics if n is less than or equal to zero.
func (slice Slice[T]) Nth(n int, offset ...int) Slice[T] {
	if n <= 0 {
		panic("nth step must be greater than zero")
	}

	start := 0
	if len(offset) > 0 {
		start = offset[0]
	}

	var result []T

	for i := start; i < len(slice); i += n {
		result = append(result, slice[i])
	}

	return NewSlice(result)
}

// ForPage returns a new slice containing the items for the given
// one-indexed page number and page size.
func (slice Slice[T]) ForPage(page, perPage int) Slice[T] {
	return slice.Skip((page - 1) * perPage).Take(perPage)
}

// MarshalJSONTo encodes the slice as a JSON array into enc, identical to
// marshaling the underlying slice directly.
func (slice Slice[T]) MarshalJSONTo(enc *jsontext.Encoder) error {
	return json.MarshalEncode(enc, []T(slice))
}

// UnmarshalJSONFrom decodes a JSON array from dec into the slice.
func (slice *Slice[T]) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	return json.UnmarshalDecode(dec, (*[]T)(slice))
}

// MarshalJSON implements [encoding/json.Marshaler] for v1 compatibility.
// When used with [encoding/json/v2], [Slice.MarshalJSONTo] takes precedence.
func (slice Slice[T]) MarshalJSON() ([]byte, error) {
	return json.Marshal([]T(slice))
}

// UnmarshalJSON implements [encoding/json.Unmarshaler] for v1 compatibility.
// When used with [encoding/json/v2], [Slice.UnmarshalJSONFrom] takes precedence.
func (slice *Slice[T]) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, (*[]T)(slice))
}
