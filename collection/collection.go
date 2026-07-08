package collection

import (
	"encoding/json/jsontext"
	"encoding/json/v2"
	"iter"
	"slices"
)

// Collection is an eager, in-memory collection of items of type T.
type Collection[T any] struct {
	items []T
}

// New creates a new [Collection] from the given slice.
func New[T any](items []T) Collection[T] {
	return Collection[T]{items: items}
}

// Iter returns an [iter.Seq] over the items in the collection, suitable for
// use in a for-range loop.
//
//	for v := range collection.Iter() { ... }
func (collection Collection[T]) Iter() iter.Seq[T] {
	return slices.Values(collection.items)
}

// Lazy returns a [LazyCollection] backed by the items in the collection.
func (collection Collection[T]) Lazy() LazyCollection[T] {
	return NewLazy(slices.Values(collection.items))
}

// Slice returns the underlying slice of items.
func (collection Collection[T]) Slice() []T {
	return collection.items
}

// IsEmpty reports whether the collection contains no items.
func (collection Collection[T]) IsEmpty() bool {
	return len(collection.items) == 0
}

// IsNotEmpty reports whether the collection contains at least one item.
func (collection Collection[T]) IsNotEmpty() bool {
	return len(collection.items) > 0
}

// Len returns the number of items in the collection.
func (collection Collection[T]) Len() int {
	return len(collection.items)
}

// Every reports whether f returns true for every item in the collection.
func (collection Collection[T]) Every(f func(T) bool) bool {
	for _, v := range collection.items {
		if !f(v) {
			return false
		}
	}

	return true
}

// Each calls f for every item in the collection.
func (collection Collection[T]) Each(f func(T)) {
	for _, v := range collection.items {
		f(v)
	}
}

// TapEach calls f for every item and returns the collection unchanged.
func (collection Collection[T]) TapEach(f func(T)) Collection[T] {
	for _, v := range collection.items {
		f(v)
	}

	return collection
}

// Filter returns a new collection containing only items for which f returns true.
func (collection Collection[T]) Filter(f func(T) bool) Collection[T] {
	var filtered []T

	for _, v := range collection.items {
		if f(v) {
			filtered = append(filtered, v)
		}
	}

	return New(filtered)
}

// Reject returns a new collection containing only items for which f returns false.
func (collection Collection[T]) Reject(f func(T) bool) Collection[T] {
	return collection.Filter(func(v T) bool { return !f(v) })
}

// Map transforms each item using f and returns a new collection of type K.
func (collection Collection[T]) Map[K any](f func(T) K) Collection[K] {
	mapped := make([]K, len(collection.items))

	for i, v := range collection.items {
		mapped[i] = f(v)
	}

	return New(mapped)
}

// FirstWhere returns the first item for which f returns true, along with a boolean
// indicating whether such an item was found.
func (collection Collection[T]) FirstWhere(f func(T) bool) (result T, ok bool) {
	for _, v := range collection.items {
		if f(v) {
			return v, true
		}
	}

	return result, false
}

// First returns the first item in the collection, along with a boolean
// indicating whether the collection is non-empty.
func (collection Collection[T]) First() (result T, ok bool) {
	if len(collection.items) == 0 {
		return result, false
	}

	return collection.items[0], true
}

// Last returns the last item in the collection, along with a boolean
// indicating whether the collection is non-empty.
func (collection Collection[T]) Last() (result T, ok bool) {
	if len(collection.items) == 0 {
		return result, false
	}

	return collection.items[len(collection.items)-1], true
}

// Contains reports whether any item in the collection satisfies f.
func (collection Collection[T]) Contains(f func(T) bool) bool {
	_, found := collection.FirstWhere(f)

	return found
}

// Take returns a new collection with at most limit items from the start.
func (collection Collection[T]) Take(limit int) Collection[T] {
	limit = max(0, min(limit, len(collection.items)))

	return New(collection.items[:limit])
}

// Skip returns a new collection with the first n items removed.
func (collection Collection[T]) Skip(n int) Collection[T] {
	n = max(0, min(n, len(collection.items)))

	return New(collection.items[n:])
}

// TakeWhile returns a new collection with items from the start until f first
// returns false.
func (collection Collection[T]) TakeWhile(f func(T) bool) Collection[T] {
	var taken []T

	for _, v := range collection.items {
		if !f(v) {
			break
		}

		taken = append(taken, v)
	}

	return New(taken)
}

// SkipWhile returns a new collection with leading items removed until f first
// returns false.
func (collection Collection[T]) SkipWhile(f func(T) bool) Collection[T] {
	i := 0

	for i < len(collection.items) && f(collection.items[i]) {
		i++
	}

	return New(collection.items[i:])
}

// Chunk splits the collection into consecutive sub-collections of the given size.
// It panics if size is less than or equal to zero.
func (collection Collection[T]) Chunk(size int) []Collection[T] {
	if size <= 0 {
		panic("chunk size must be greater than zero")
	}

	var chunks []Collection[T]

	for i := 0; i < len(collection.items); i += size {
		end := min(i+size, len(collection.items))
		chunks = append(chunks, New(collection.items[i:end]))
	}

	return chunks
}

// Sliding returns a slice of overlapping windows of the given size, advancing
// by step items between each window. The step defaults to one if not provided.
// It panics if size or step is less than or equal to zero.
func (collection Collection[T]) Sliding(size int, step ...int) []Collection[T] {
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

	var windows []Collection[T]

	for i := 0; i+size <= len(collection.items); i += s {
		windows = append(windows, New(collection.items[i:i+size]))
	}

	return windows
}

// Reduce reduces the collection to a single value of type K by applying f
// to each item starting from initial.
func (collection Collection[T]) Reduce[K any](f func(K, T) K, initial K) K {
	acc := initial

	for _, v := range collection.items {
		acc = f(acc, v)
	}

	return acc
}

// Reverse returns a new collection with items in reversed order.
func (collection Collection[T]) Reverse() Collection[T] {
	reversed := slices.Clone(collection.items)
	slices.Reverse(reversed)

	return New(reversed)
}

// Sort returns a new collection with items sorted using the given comparison function.
func (collection Collection[T]) Sort(cmp func(T, T) int) Collection[T] {
	sorted := slices.Clone(collection.items)
	slices.SortFunc(sorted, cmp)

	return New(sorted)
}

// Unique returns a new collection containing only the first occurrence of each
// item, as determined by the key function.
func (collection Collection[T]) Unique[K comparable](key func(T) K) Collection[T] {
	seen := make(map[K]struct{})
	var unique []T

	for _, v := range collection.items {
		k := key(v)

		if _, exists := seen[k]; !exists {
			seen[k] = struct{}{}
			unique = append(unique, v)
		}
	}

	return New(unique)
}

// Partition splits the collection into two: the first contains items for which
// f returns true, the second contains the rest.
func (collection Collection[T]) Partition(f func(T) bool) (Collection[T], Collection[T]) {
	var matching, rest []T

	for _, v := range collection.items {
		if f(v) {
			matching = append(matching, v)
		} else {
			rest = append(rest, v)
		}
	}

	return New(matching), New(rest)
}

// Concat returns a new collection with the items of other appended.
func (collection Collection[T]) Concat(other Collection[T]) Collection[T] {
	return New(append(slices.Clone(collection.items), other.items...))
}

// GroupBy groups items by the key returned by the key function, returning a
// map of sub-collections.
func (collection Collection[T]) GroupBy[K comparable](key func(T) K) map[K]Collection[T] {
	grouped := make(map[K][]T)

	for _, v := range collection.items {
		k := key(v)
		grouped[k] = append(grouped[k], v)
	}

	result := make(map[K]Collection[T], len(grouped))

	for k, items := range grouped {
		result[k] = New(items)
	}

	return result
}

// Nth returns a new collection consisting of every nth item, starting at the
// given offset. The offset defaults to zero if not provided.
// It panics if n is less than or equal to zero.
func (collection Collection[T]) Nth(n int, offset ...int) Collection[T] {
	if n <= 0 {
		panic("nth step must be greater than zero")
	}

	start := 0
	if len(offset) > 0 {
		start = offset[0]
	}

	var result []T

	for i := start; i < len(collection.items); i += n {
		result = append(result, collection.items[i])
	}

	return New(result)
}

// ForPage returns a new collection containing the items for the given
// one-indexed page number and page size.
func (collection Collection[T]) ForPage(page, perPage int) Collection[T] {
	return collection.Skip((page - 1) * perPage).Take(perPage)
}

// MarshalJSONTo encodes the collection as a JSON array into enc, identical to
// marshaling the underlying slice directly.
func (collection Collection[T]) MarshalJSONTo(enc *jsontext.Encoder) error {
	return json.MarshalEncode(enc, collection.items)
}

// UnmarshalJSONFrom decodes a JSON array from dec into the collection.
func (collection *Collection[T]) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	return json.UnmarshalDecode(dec, &collection.items)
}

// MarshalJSON implements [encoding/json.Marshaler] for v1 compatibility.
// When used with [encoding/json/v2], [Collection.MarshalJSONTo] takes precedence.
func (collection Collection[T]) MarshalJSON() ([]byte, error) {
	return json.Marshal(collection.items)
}

// UnmarshalJSON implements [encoding/json.Unmarshaler] for v1 compatibility.
// When used with [encoding/json/v2], [Collection.UnmarshalJSONFrom] takes precedence.
func (collection *Collection[T]) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &collection.items)
}
