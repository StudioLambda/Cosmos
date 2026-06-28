package collection

import (
	"iter"
	"slices"
)

type LazyCollection[T any] struct {
	seq iter.Seq[T]
}

func NewLazy[T any](seq iter.Seq[T]) *LazyCollection[T] {
	return &LazyCollection[T]{
		seq: seq,
	}
}

func (c *LazyCollection[T]) Slice() []T {
	return slices.Collect(c.seq)
}

func (c *LazyCollection[T]) IsEmpty() bool {
	for range c.seq {
		return false
	}

	return true
}

func (c *LazyCollection[T]) Every(f func(T) bool) bool {
	for v := range c.seq {
		if !f(v) {
			return false
		}
	}

	return true
}

func (c *LazyCollection[T]) Each(f func(T)) {
	for v := range c.seq {
		f(v)
	}
}

func (c *LazyCollection[T]) TapEach(f func(T)) *LazyCollection[T] {
	return NewLazy(func(yield func(T) bool) {
		for v := range c.seq {
			f(v)

			if !yield(v) {
				return
			}
		}
	})
}

func (c *LazyCollection[T]) Filter(f func(T) bool) *LazyCollection[T] {
	return NewLazy(func(yield func(T) bool) {
		for v := range c.seq {
			if f(v) {
				if !yield(v) {
					return
				}
			}
		}
	})
}

func (c *LazyCollection[T]) Map[K any](f func(T) K) *LazyCollection[K] {
	return NewLazy(func(yield func(K) bool) {
		for v := range c.seq {
			if !yield(f(v)) {
				return
			}
		}
	})
}

func (c *LazyCollection[T]) FirstWhere(f func(T) bool) (T, bool) {
	for v := range c.seq {
		if f(v) {
			return v, true
		}
	}

	var zero T

	return zero, false
}

func (c *LazyCollection[T]) Take(limit int) *LazyCollection[T] {
	return NewLazy(func(yield func(T) bool) {
		if limit <= 0 {
			return
		}

		count := 0

		for v := range c.seq {
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

func (c *LazyCollection[T]) Chunk(size int) iter.Seq[[]T] {
	if size <= 0 {
		panic("chunk size must be greater than zero")
	}

	return func(yield func([]T) bool) {
		var chunk []T

		for v := range c.seq {
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

func (c *LazyCollection[T]) Reduce[K any](f func(K, T) K, initial K) K {
	acc := initial

	for v := range c.seq {
		acc = f(acc, v)
	}

	return acc
}

func (c *LazyCollection[T]) Contains(f func(T) bool) bool {
	_, found := c.FirstWhere(f)

	return found
}
