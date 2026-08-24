package collection

import (
	"encoding/json/jsontext"
	"encoding/json/v2"
	"errors"
	"iter"
	"slices"
)

// TryLazySlice is a lazily-evaluated collection backed by an [iter.Seq2].
//
// The yielded value must be ignored whenever the yielded error is non-nil.
// Being a function type, it can be used directly in a for-range loop:
//
//	for v, err := range tryLazySlice { ... }
type TryLazySlice[T any] func(yield func(T, error) bool)

// NewTryLazySlice creates a new [TryLazySlice] from the given iterator function.
func NewTryLazySlice[T any](f func(yield func(T, error) bool)) TryLazySlice[T] {
	return TryLazySlice[T](f)
}

// Eager materializes the lazy slice into an eager [Slice].
//
// It stops on the first error and returns the items collected before that error.
func (tryLazySlice TryLazySlice[T]) Eager() (Slice[T], error) {
	items, err := tryLazySlice.Items()

	return NewSlice(items), err
}

// EagerAll materializes the lazy slice into an eager [Slice].
//
// It consumes the entire sequence, collecting all successful items and joining
// all errors with [errors.Join].
func (tryLazySlice TryLazySlice[T]) EagerAll() (Slice[T], error) {
	items, err := tryLazySlice.ItemsAll()

	return NewSlice(items), err
}

// Items materializes all successful items from the iterator into a Go slice.
//
// It stops on the first error and returns the items collected before that error.
func (tryLazySlice TryLazySlice[T]) Items() ([]T, error) {
	items := make([]T, 0)

	for v, err := range tryLazySlice {
		if err != nil {
			return items, err
		}

		items = append(items, v)
	}

	return items, nil
}

// ItemsAll materializes all successful items from the iterator into a Go slice.
//
// It consumes the entire sequence and joins all errors with [errors.Join].
func (tryLazySlice TryLazySlice[T]) ItemsAll() ([]T, error) {
	items := make([]T, 0)
	var errs []error

	for v, err := range tryLazySlice {
		if err != nil {
			errs = append(errs, err)

			continue
		}

		items = append(items, v)
	}

	return items, errors.Join(errs...)
}

// ItemsAs materializes all successful items into the given slice type.
//
// It stops on the first error and returns the items collected before that error.
func (tryLazySlice TryLazySlice[T]) ItemsAs[S ~[]T]() (S, error) {
	items, err := tryLazySlice.Items()

	return S(items), err
}

// ItemsAsAll materializes all successful items into the given slice type.
//
// It consumes the entire sequence and joins all errors with [errors.Join].
func (tryLazySlice TryLazySlice[T]) ItemsAsAll[S ~[]T]() (S, error) {
	items, err := tryLazySlice.ItemsAll()

	return S(items), err
}

// Iter returns the underlying [iter.Seq2].
func (tryLazySlice TryLazySlice[T]) Iter() iter.Seq2[T, error] {
	return iter.Seq2[T, error](tryLazySlice)
}

// IsEmpty reports whether the iterator yields no successful items.
//
// It stops on the first error.
func (tryLazySlice TryLazySlice[T]) IsEmpty() (bool, error) {
	for _, err := range tryLazySlice {
		if err != nil {
			return false, err
		}

		return false, nil
	}

	return true, nil
}

// IsNotEmpty reports whether the iterator yields at least one successful item.
//
// It stops on the first error.
func (tryLazySlice TryLazySlice[T]) IsNotEmpty() (bool, error) {
	result, err := tryLazySlice.IsEmpty()

	return !result, err
}

// Len returns the total number of successful items yielded by the iterator.
//
// It stops on the first error.
func (tryLazySlice TryLazySlice[T]) Len() (int, error) {
	count := 0

	for _, err := range tryLazySlice {
		if err != nil {
			return count, err
		}

		count++
	}

	return count, nil
}

// Every reports whether f returns true for every successful item yielded by the iterator.
//
// It stops on the first error.
func (tryLazySlice TryLazySlice[T]) Every(f func(int, T) (bool, error)) (bool, error) {
	index := 0

	for v, err := range tryLazySlice {
		if err != nil {
			return false, err
		}

		ok, err := f(index, v)
		if err != nil {
			return false, err
		}

		if !ok {
			return false, nil
		}
		index++
	}

	return true, nil
}

// Each calls f for every successful item yielded by the iterator.
//
// It stops on the first error, whether yielded by the iterator or returned by f.
func (tryLazySlice TryLazySlice[T]) Each(f func(int, T) error) error {
	index := 0

	for v, err := range tryLazySlice {
		if err != nil {
			return err
		}

		if err := f(index, v); err != nil {
			return err
		}
		index++
	}

	return nil
}

// EachAll calls f for every successful item yielded by the iterator.
//
// It consumes the entire sequence and joins all iterator and callback errors
// with [errors.Join].
func (tryLazySlice TryLazySlice[T]) EachAll(f func(int, T) error) error {
	var errs []error

	index := 0

	for v, err := range tryLazySlice {
		if err != nil {
			errs = append(errs, err)

			continue
		}

		if err := f(index, v); err != nil {
			errs = append(errs, err)
		}
		index++
	}

	return errors.Join(errs...)
}

// TapEach returns a lazy slice that calls f on each successful item as it is consumed.
func (tryLazySlice TryLazySlice[T]) TapEach(f func(int, T) error) TryLazySlice[T] {
	return NewTryLazySlice(func(yield func(T, error) bool) {
		index := 0

		for v, err := range tryLazySlice {
			if err != nil {
				if !tryYieldError(yield, err) {
					return
				}

				continue
			}

			if err := f(index, v); err != nil {
				if !tryYieldError(yield, err) {
					return
				}

				continue
			}

			if !yield(v, nil) {
				return
			}

			index++
		}
	})
}

// Filter returns a lazy slice containing only successful items for which f returns true.
func (tryLazySlice TryLazySlice[T]) Filter(f func(int, T) (bool, error)) TryLazySlice[T] {
	return NewTryLazySlice(func(yield func(T, error) bool) {
		index := 0

		for v, err := range tryLazySlice {
			if err != nil {
				if !tryYieldError(yield, err) {
					return
				}

				continue
			}

			ok, err := f(index, v)
			if err != nil {
				if !tryYieldError(yield, err) {
					return
				}

				continue
			}

			if ok {
				if !yield(v, nil) {
					return
				}
			}

			index++
		}
	})
}

// Reject returns a lazy slice containing only successful items for which f returns false.
func (tryLazySlice TryLazySlice[T]) Reject(f func(int, T) (bool, error)) TryLazySlice[T] {
	return tryLazySlice.Filter(func(index int, v T) (bool, error) {
		ok, err := f(index, v)
		if err != nil {
			return false, err
		}

		return !ok, nil
	})
}

// Map transforms each successful item using f and returns a new lazy slice of type K.
func (tryLazySlice TryLazySlice[T]) Map[K any](f func(int, T) (K, error)) TryLazySlice[K] {
	return NewTryLazySlice(func(yield func(K, error) bool) {
		index := 0

		for v, err := range tryLazySlice {
			if err != nil {
				if !tryYieldError(yield, err) {
					return
				}

				continue
			}

			mapped, err := f(index, v)
			if err != nil {
				if !tryYieldError(yield, err) {
					return
				}

				continue
			}

			if !yield(mapped, nil) {
				return
			}

			index++
		}
	})
}

// FlatMap transforms each successful item into zero or more items and yields the flattened results lazily.
func (tryLazySlice TryLazySlice[T]) FlatMap[K any](f func(int, T) ([]K, error)) TryLazySlice[K] {
	return NewTryLazySlice(func(yield func(K, error) bool) {
		index := 0

		for v, err := range tryLazySlice {
			if err != nil {
				if !tryYieldError(yield, err) {
					return
				}

				continue
			}

			mapped, err := f(index, v)
			if err != nil {
				if !tryYieldError(yield, err) {
					return
				}

				continue
			}

			for _, item := range mapped {
				if !yield(item, nil) {
					return
				}
			}

			index++
		}
	})
}

// KeyBy indexes yielded items by the key returned from f.
// When multiple items map to the same key, the last item wins.
//
// It stops on the first error.
func (tryLazySlice TryLazySlice[T]) KeyBy[K comparable](f func(int, T) (K, error)) (Map[K, T], error) {
	indexed := make(map[K]T)

	index := 0

	for v, err := range tryLazySlice {
		if err != nil {
			return NewMap(indexed), err
		}

		key, err := f(index, v)
		if err != nil {
			return NewMap(indexed), err
		}

		indexed[key] = v
		index++
	}

	return NewMap(indexed), nil
}

// CountBy counts yielded items by the key returned from f.
//
// It stops on the first error.
func (tryLazySlice TryLazySlice[T]) CountBy[K comparable](f func(int, T) (K, error)) (Map[K, int], error) {
	counts := make(map[K]int)

	index := 0

	for v, err := range tryLazySlice {
		if err != nil {
			return NewMap(counts), err
		}

		key, err := f(index, v)
		if err != nil {
			return NewMap(counts), err
		}

		counts[key]++
		index++
	}

	return NewMap(counts), nil
}

// FirstWhere returns the first successful item for which f returns true.
//
// It stops on the first error.
func (tryLazySlice TryLazySlice[T]) FirstWhere(f func(int, T) (bool, error)) (result T, ok bool, err error) {
	index := 0

	for v, err := range tryLazySlice {
		if err != nil {
			return result, false, err
		}

		matched, err := f(index, v)
		if err != nil {
			return result, false, err
		}

		if matched {
			return v, true, nil
		}

		index++
	}

	return result, false, nil
}

// First returns the first successful item yielded by the iterator.
//
// It stops on the first error.
func (tryLazySlice TryLazySlice[T]) First() (result T, ok bool, err error) {
	for v, err := range tryLazySlice {
		if err != nil {
			return result, false, err
		}

		return v, true, nil
	}

	return result, false, nil
}

// Last returns the last successful item yielded by the iterator.
//
// It stops on the first error.
func (tryLazySlice TryLazySlice[T]) Last() (result T, ok bool, err error) {
	for v, err := range tryLazySlice {
		if err != nil {
			return result, ok, err
		}

		result = v
		ok = true
	}

	return result, ok, nil
}

// Contains reports whether any successful item yielded by the iterator satisfies f.
//
// It stops on the first error.
func (tryLazySlice TryLazySlice[T]) Contains(f func(int, T) (bool, error)) (bool, error) {
	_, found, err := tryLazySlice.FirstWhere(f)

	return found, err
}

// Take returns a lazy slice that yields at most limit successful items.
func (tryLazySlice TryLazySlice[T]) Take(limit int) TryLazySlice[T] {
	return NewTryLazySlice(func(yield func(T, error) bool) {
		if limit <= 0 {
			return
		}

		count := 0

		for v, err := range tryLazySlice {
			if err != nil {
				if !tryYieldError(yield, err) {
					return
				}

				continue
			}

			if !yield(v, nil) {
				return
			}

			count++

			if count >= limit {
				return
			}
		}
	})
}

// Skip returns a lazy slice that skips the first n successful items.
func (tryLazySlice TryLazySlice[T]) Skip(n int) TryLazySlice[T] {
	return NewTryLazySlice(func(yield func(T, error) bool) {
		skipped := 0

		for v, err := range tryLazySlice {
			if err != nil {
				if !tryYieldError(yield, err) {
					return
				}

				continue
			}

			if skipped < n {
				skipped++

				continue
			}

			if !yield(v, nil) {
				return
			}
		}
	})
}

// TakeWhile returns a lazy slice that yields successful items until f first returns false.
func (tryLazySlice TryLazySlice[T]) TakeWhile(f func(int, T) (bool, error)) TryLazySlice[T] {
	return NewTryLazySlice(func(yield func(T, error) bool) {
		index := 0

		for v, err := range tryLazySlice {
			if err != nil {
				if !tryYieldError(yield, err) {
					return
				}

				continue
			}

			ok, err := f(index, v)
			if err != nil {
				if !tryYieldError(yield, err) {
					return
				}

				continue
			}

			if !ok {
				return
			}

			if !yield(v, nil) {
				return
			}

			index++
		}
	})
}

// SkipWhile returns a lazy slice that skips successful items until f first returns false,
// then yields all remaining successful items.
func (tryLazySlice TryLazySlice[T]) SkipWhile(f func(int, T) (bool, error)) TryLazySlice[T] {
	return NewTryLazySlice(func(yield func(T, error) bool) {
		skipping := true

		index := 0

		for v, err := range tryLazySlice {
			if err != nil {
				if !tryYieldError(yield, err) {
					return
				}

				continue
			}

			if skipping {
				ok, err := f(index, v)
				if err != nil {
					if !tryYieldError(yield, err) {
						return
					}

					continue
				}

				if ok {
					index++
					continue
				}

				skipping = false
			}

			if !yield(v, nil) {
				return
			}

			index++
		}
	})
}

// TakeUntil returns a lazy slice that yields successful items until f first returns true.
func (tryLazySlice TryLazySlice[T]) TakeUntil(f func(int, T) (bool, error)) TryLazySlice[T] {
	return tryLazySlice.TakeWhile(func(index int, v T) (bool, error) {
		ok, err := f(index, v)
		if err != nil {
			return false, err
		}

		return !ok, nil
	})
}

// SkipUntil returns a lazy slice that skips successful items until f first returns true,
// then yields all remaining successful items.
func (tryLazySlice TryLazySlice[T]) SkipUntil(f func(int, T) (bool, error)) TryLazySlice[T] {
	return tryLazySlice.SkipWhile(func(index int, v T) (bool, error) {
		ok, err := f(index, v)
		if err != nil {
			return false, err
		}

		return !ok, nil
	})
}

// Chunk returns an iterator over consecutive slices of the given size.
// It panics if size is less than or equal to zero.
func (tryLazySlice TryLazySlice[T]) Chunk(size int) iter.Seq2[[]T, error] {
	if size <= 0 {
		panic("chunk size must be greater than zero")
	}

	return func(yield func([]T, error) bool) {
		var chunk []T

		for v, err := range tryLazySlice {
			if err != nil {
				if !tryYieldError(yield, err) {
					return
				}

				continue
			}

			chunk = append(chunk, v)

			if len(chunk) == size {
				if !yield(chunk, nil) {
					return
				}

				chunk = nil
			}
		}

		if len(chunk) > 0 {
			yield(chunk, nil)
		}
	}
}

// Sliding returns an iterator over overlapping windows of the given size,
// advancing by step successful items between each window. The step defaults to one
// if not provided. Each yielded slice is a fresh copy.
// It panics if size or step is less than or equal to zero.
func (tryLazySlice TryLazySlice[T]) Sliding(size int, step ...int) iter.Seq2[[]T, error] {
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

	return func(yield func([]T, error) bool) {
		window := make([]T, 0, size)

		for v, err := range tryLazySlice {
			if err != nil {
				if !tryYieldError(yield, err) {
					return
				}

				continue
			}

			window = append(window, v)

			if len(window) == size {
				if !yield(slices.Clone(window), nil) {
					return
				}

				if s >= len(window) {
					window = nil

					continue
				}

				window = window[s:]
			}
		}
	}
}

// Reduce reduces the lazy slice to a single value of type K by applying f
// to each successful item starting from initial.
//
// It stops on the first error.
func (tryLazySlice TryLazySlice[T]) Reduce[K any](f func(K, int, T) (K, error), initial K) (K, error) {
	acc := initial

	index := 0

	for v, err := range tryLazySlice {
		if err != nil {
			return acc, err
		}

		acc, err = f(acc, index, v)
		if err != nil {
			return acc, err
		}

		index++
	}

	return acc, nil
}

// Reverse materializes the iterator and returns a new lazy slice with items in reversed order.
//
// If materialization fails, the returned sequence yields that error and no items.
func (tryLazySlice TryLazySlice[T]) Reverse() TryLazySlice[T] {
	return NewTryLazySlice(func(yield func(T, error) bool) {
		items, err := tryLazySlice.Items()
		if err != nil {
			tryYieldError(yield, err)

			return
		}

		slices.Reverse(items)

		for _, v := range items {
			if !yield(v, nil) {
				return
			}
		}
	})
}

// Sort materializes the iterator, sorts items using the given comparison function,
// and returns a new lazy slice over the sorted items.
//
// If materialization fails, the returned sequence yields that error and no items.
func (tryLazySlice TryLazySlice[T]) Sort(cmp func(T, T) int) TryLazySlice[T] {
	return NewTryLazySlice(func(yield func(T, error) bool) {
		items, err := tryLazySlice.Items()
		if err != nil {
			tryYieldError(yield, err)

			return
		}

		slices.SortFunc(items, cmp)

		for _, v := range items {
			if !yield(v, nil) {
				return
			}
		}
	})
}

// Unique returns a lazy slice containing only the first successful occurrence of each
// item, as determined by the key function.
func (tryLazySlice TryLazySlice[T]) Unique[K comparable](key func(int, T) (K, error)) TryLazySlice[T] {
	return NewTryLazySlice(func(yield func(T, error) bool) {
		seen := make(map[K]struct{})

		index := 0

		for v, err := range tryLazySlice {
			if err != nil {
				if !tryYieldError(yield, err) {
					return
				}

				continue
			}

			k, err := key(index, v)
			if err != nil {
				if !tryYieldError(yield, err) {
					return
				}

				continue
			}

			if _, exists := seen[k]; exists {
				index++
				continue
			}

			seen[k] = struct{}{}

			if !yield(v, nil) {
				return
			}

			index++
		}
	})
}

// Partition splits the iterator into two lazy slices: the first contains items for which
// f returns true, the second contains the rest.
// It fully consumes the sequence to allow both results to be independently iterated.
//
// If materialization fails, both returned sequences yield that error and no items.
func (tryLazySlice TryLazySlice[T]) Partition(f func(int, T) (bool, error)) (TryLazySlice[T], TryLazySlice[T]) {
	var matching, rest []T

	index := 0

	for v, err := range tryLazySlice {
		if err != nil {
			return tryErrorSlice[T](err), tryErrorSlice[T](err)
		}

		matched, err := f(index, v)
		if err != nil {
			return tryErrorSlice[T](err), tryErrorSlice[T](err)
		}

		if matched {
			matching = append(matching, v)
			index++
			continue
		}

		rest = append(rest, v)
		index++
	}

	return NewSlice(matching).Try(), NewSlice(rest).Try()
}

// Concat returns a lazy slice that yields all items from this slice followed by all items from other.
func (tryLazySlice TryLazySlice[T]) Concat(other TryLazySlice[T]) TryLazySlice[T] {
	return NewTryLazySlice(func(yield func(T, error) bool) {
		for v, err := range tryLazySlice {
			if !yield(v, err) {
				return
			}
		}

		for v, err := range other {
			if !yield(v, err) {
				return
			}
		}
	})
}

// GroupBy groups all successful items by the key returned by the key function,
// returning a [Map] of sub-slices. It fully consumes the sequence.
//
// It stops on the first error.
func (tryLazySlice TryLazySlice[T]) GroupBy[K comparable](key func(int, T) (K, error)) (Map[K, Slice[T]], error) {
	grouped := make(map[K][]T)

	index := 0

	for v, err := range tryLazySlice {
		if err != nil {
			return tryGroupByResult(grouped), err
		}

		k, err := key(index, v)
		if err != nil {
			return tryGroupByResult(grouped), err
		}

		grouped[k] = append(grouped[k], v)
		index++
	}

	return tryGroupByResult(grouped), nil
}

// Nth returns a lazy slice consisting of every nth successful item, starting at the
// given offset. The offset defaults to zero if not provided.
// It panics if n is less than or equal to zero.
func (tryLazySlice TryLazySlice[T]) Nth(n int, offset ...int) TryLazySlice[T] {
	if n <= 0 {
		panic("nth step must be greater than zero")
	}

	start := 0
	if len(offset) > 0 {
		start = offset[0]
	}

	return NewTryLazySlice(func(yield func(T, error) bool) {
		i := 0

		for v, err := range tryLazySlice {
			if err != nil {
				if !tryYieldError(yield, err) {
					return
				}

				continue
			}

			if i >= start && (i-start)%n == 0 {
				if !yield(v, nil) {
					return
				}
			}

			i++
		}
	})
}

// ForPage returns a lazy slice containing the successful items for the given
// one-indexed page number and page size.
func (tryLazySlice TryLazySlice[T]) ForPage(page, perPage int) TryLazySlice[T] {
	return tryLazySlice.Skip((page - 1) * perPage).Take(perPage)
}

// MarshalJSONTo materializes the iterator and encodes it as a JSON array into enc.
func (tryLazySlice TryLazySlice[T]) MarshalJSONTo(enc *jsontext.Encoder) error {
	items, err := tryLazySlice.Items()
	if err != nil {
		return err
	}

	return json.MarshalEncode(enc, items)
}

// UnmarshalJSONFrom decodes a JSON array from dec into the lazy slice,
// backed by the resulting slice.
func (tryLazySlice *TryLazySlice[T]) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	var items []T

	if err := json.UnmarshalDecode(dec, &items); err != nil {
		return err
	}

	*tryLazySlice = NewSlice(items).Try()

	return nil
}

// MarshalJSON implements [encoding/json.Marshaler] for v1 compatibility.
// When used with [encoding/json/v2], [TryLazySlice.MarshalJSONTo] takes precedence.
func (tryLazySlice TryLazySlice[T]) MarshalJSON() ([]byte, error) {
	items, err := tryLazySlice.Items()
	if err != nil {
		return nil, err
	}

	return json.Marshal(items)
}

// UnmarshalJSON implements [encoding/json.Unmarshaler] for v1 compatibility.
// When used with [encoding/json/v2], [TryLazySlice.UnmarshalJSONFrom] takes precedence.
func (tryLazySlice *TryLazySlice[T]) UnmarshalJSON(data []byte) error {
	var items []T

	if err := json.Unmarshal(data, &items); err != nil {
		return err
	}

	*tryLazySlice = NewSlice(items).Try()

	return nil
}

// tryErrorSlice returns a [TryLazySlice] that yields the given error and no items.
func tryErrorSlice[T any](err error) TryLazySlice[T] {
	return NewTryLazySlice(func(yield func(T, error) bool) {
		tryYieldError(yield, err)
	})
}

// tryGroupByResult converts grouped items into a [Map] of eager [Slice] values.
func tryGroupByResult[K comparable, T any](grouped map[K][]T) Map[K, Slice[T]] {
	result := make(map[K]Slice[T], len(grouped))

	for k, items := range grouped {
		result[k] = NewSlice(items)
	}

	return NewMap(result)
}

// tryYieldError yields err with the zero value of T.
func tryYieldError[T any](yield func(T, error) bool, err error) bool {
	var zero T

	return yield(zero, err)
}
