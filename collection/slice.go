package collection

import (
	"encoding/json/jsontext"
	"encoding/json/v2"
	"iter"
	"slices"
)

// Slice is an eager, in-memory collection of items of type T.
type Slice[T any] struct {
	items []T
}

// NewSlice creates a new [Slice] from the given slice.
func NewSlice[T any](items []T) Slice[T] {
	return Slice[T]{items: items}
}

// Iter returns an [iter.Seq] over the items in the slice, suitable for
// use in a for-range loop.
//
//	for v := range slice.Iter() { ... }
func (slice Slice[T]) Iter() iter.Seq[T] {
	return slices.Values(slice.items)
}

// Lazy returns a [LazySlice] backed by the items in the slice.
func (slice Slice[T]) Lazy() LazySlice[T] {
	return NewLazySlice(slices.Values(slice.items))
}

// Items returns the underlying slice of items.
func (slice Slice[T]) Items() []T {
	return slice.items
}

// IsEmpty reports whether the slice contains no items.
func (slice Slice[T]) IsEmpty() bool {
	return len(slice.items) == 0
}

// IsNotEmpty reports whether the slice contains at least one item.
func (slice Slice[T]) IsNotEmpty() bool {
	return len(slice.items) > 0
}

// Len returns the number of items in the slice.
func (slice Slice[T]) Len() int {
	return len(slice.items)
}

// Every reports whether f returns true for every item in the slice.
func (slice Slice[T]) Every(f func(T) bool) bool {
	for _, v := range slice.items {
		if !f(v) {
			return false
		}
	}

	return true
}

// Each calls f for every item in the slice.
func (slice Slice[T]) Each(f func(T)) {
	for _, v := range slice.items {
		f(v)
	}
}

// TapEach calls f for every item and returns the slice unchanged.
func (slice Slice[T]) TapEach(f func(T)) Slice[T] {
	for _, v := range slice.items {
		f(v)
	}

	return slice
}

// Filter returns a new slice containing only items for which f returns true.
func (slice Slice[T]) Filter(f func(T) bool) Slice[T] {
	var filtered []T

	for _, v := range slice.items {
		if f(v) {
			filtered = append(filtered, v)
		}
	}

	return NewSlice(filtered)
}

// Reject returns a new slice containing only items for which f returns false.
func (slice Slice[T]) Reject(f func(T) bool) Slice[T] {
	return slice.Filter(func(v T) bool { return !f(v) })
}

// Map transforms each item using f and returns a new slice of type K.
func (slice Slice[T]) Map[K any](f func(T) K) Slice[K] {
	mapped := make([]K, len(slice.items))

	for i, v := range slice.items {
		mapped[i] = f(v)
	}

	return NewSlice(mapped)
}

// FirstWhere returns the first item for which f returns true, along with a boolean
// indicating whether such an item was found.
func (slice Slice[T]) FirstWhere(f func(T) bool) (result T, ok bool) {
	for _, v := range slice.items {
		if f(v) {
			return v, true
		}
	}

	return result, false
}

// First returns the first item in the slice, along with a boolean
// indicating whether the slice is non-empty.
func (slice Slice[T]) First() (result T, ok bool) {
	if len(slice.items) == 0 {
		return result, false
	}

	return slice.items[0], true
}

// Last returns the last item in the slice, along with a boolean
// indicating whether the slice is non-empty.
func (slice Slice[T]) Last() (result T, ok bool) {
	if len(slice.items) == 0 {
		return result, false
	}

	return slice.items[len(slice.items)-1], true
}

// Contains reports whether any item in the slice satisfies f.
func (slice Slice[T]) Contains(f func(T) bool) bool {
	_, found := slice.FirstWhere(f)

	return found
}

// Take returns a new slice with at most limit items from the start.
func (slice Slice[T]) Take(limit int) Slice[T] {
	limit = max(0, min(limit, len(slice.items)))

	return NewSlice(slice.items[:limit])
}

// Skip returns a new slice with the first n items removed.
func (slice Slice[T]) Skip(n int) Slice[T] {
	n = max(0, min(n, len(slice.items)))

	return NewSlice(slice.items[n:])
}

// TakeWhile returns a new slice with items from the start until f first
// returns false.
func (slice Slice[T]) TakeWhile(f func(T) bool) Slice[T] {
	var taken []T

	for _, v := range slice.items {
		if !f(v) {
			break
		}

		taken = append(taken, v)
	}

	return NewSlice(taken)
}

// SkipWhile returns a new slice with leading items removed until f first
// returns false.
func (slice Slice[T]) SkipWhile(f func(T) bool) Slice[T] {
	i := 0

	for i < len(slice.items) && f(slice.items[i]) {
		i++
	}

	return NewSlice(slice.items[i:])
}

// Chunk splits the slice into consecutive sub-slices of the given size.
// It panics if size is less than or equal to zero.
func (slice Slice[T]) Chunk(size int) []Slice[T] {
	if size <= 0 {
		panic("chunk size must be greater than zero")
	}

	var chunks []Slice[T]

	for i := 0; i < len(slice.items); i += size {
		end := min(i+size, len(slice.items))
		chunks = append(chunks, NewSlice(slice.items[i:end]))
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

	for i := 0; i+size <= len(slice.items); i += s {
		windows = append(windows, NewSlice(slice.items[i:i+size]))
	}

	return windows
}

// Reduce reduces the slice to a single value of type K by applying f
// to each item starting from initial.
func (slice Slice[T]) Reduce[K any](f func(K, T) K, initial K) K {
	acc := initial

	for _, v := range slice.items {
		acc = f(acc, v)
	}

	return acc
}

// Reverse returns a new slice with items in reversed order.
func (slice Slice[T]) Reverse() Slice[T] {
	reversed := slices.Clone(slice.items)
	slices.Reverse(reversed)

	return NewSlice(reversed)
}

// Sort returns a new slice with items sorted using the given comparison function.
func (slice Slice[T]) Sort(cmp func(T, T) int) Slice[T] {
	sorted := slices.Clone(slice.items)
	slices.SortFunc(sorted, cmp)

	return NewSlice(sorted)
}

// Unique returns a new slice containing only the first occurrence of each
// item, as determined by the key function.
func (slice Slice[T]) Unique[K comparable](key func(T) K) Slice[T] {
	seen := make(map[K]struct{})
	var unique []T

	for _, v := range slice.items {
		k := key(v)

		if _, exists := seen[k]; !exists {
			seen[k] = struct{}{}
			unique = append(unique, v)
		}
	}

	return NewSlice(unique)
}

// Partition splits the slice into two: the first contains items for which
// f returns true, the second contains the rest.
func (slice Slice[T]) Partition(f func(T) bool) (Slice[T], Slice[T]) {
	var matching, rest []T

	for _, v := range slice.items {
		if f(v) {
			matching = append(matching, v)
		} else {
			rest = append(rest, v)
		}
	}

	return NewSlice(matching), NewSlice(rest)
}

// Concat returns a new slice with the items of other appended.
func (slice Slice[T]) Concat(other Slice[T]) Slice[T] {
	return NewSlice(append(slices.Clone(slice.items), other.items...))
}

// GroupBy groups items by the key returned by the key function, returning a
// [Map] of sub-slices.
func (slice Slice[T]) GroupBy[K comparable](key func(T) K) Map[K, Slice[T]] {
	grouped := make(map[K][]T)

	for _, v := range slice.items {
		k := key(v)
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

	for i := start; i < len(slice.items); i += n {
		result = append(result, slice.items[i])
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
	return json.MarshalEncode(enc, slice.items)
}

// UnmarshalJSONFrom decodes a JSON array from dec into the slice.
func (slice *Slice[T]) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	return json.UnmarshalDecode(dec, &slice.items)
}

// MarshalJSON implements [encoding/json.Marshaler] for v1 compatibility.
// When used with [encoding/json/v2], [Slice.MarshalJSONTo] takes precedence.
func (slice Slice[T]) MarshalJSON() ([]byte, error) {
	return json.Marshal(slice.items)
}

// UnmarshalJSON implements [encoding/json.Unmarshaler] for v1 compatibility.
// When used with [encoding/json/v2], [Slice.UnmarshalJSONFrom] takes precedence.
func (slice *Slice[T]) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &slice.items)
}
