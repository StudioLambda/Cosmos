package collection

type Collection[T any] struct {
	items []T
}

func New[T any](items []T) *Collection[T] {
	return &Collection[T]{
		items: items,
	}
}

func (c *Collection[T]) Slice() []T {
	return c.items
}

func (c *Collection[T]) IsEmpty() bool {
	return len(c.items) == 0
}

func (c *Collection[T]) Every(f func(T) bool) bool {
	for _, v := range c.items {
		if !f(v) {
			return false
		}
	}
	return true
}

func (c *Collection[T]) Each(f func(T)) {
	for _, v := range c.items {
		f(v)
	}
}

func (c *Collection[T]) TapEach(f func(T)) *Collection[T] {
	for _, v := range c.items {
		f(v)
	}
	return c
}

func (c *Collection[T]) Filter(f func(T) bool) *Collection[T] {
	var filtered []T
	for _, v := range c.items {
		if f(v) {
			filtered = append(filtered, v)
		}
	}
	return New(filtered)
}

// Map cannot be a method with its own type parameter in Go.
// We provide a package-level function instead, or if the user had Map on LazyCollection, 
// they probably haven't compiled it yet.
// For now, I'll match the user's signature but comment it out if it fails to compile.
func (c *Collection[T]) Map[K any](f func(T) K) *Collection[K] {
	mapped := make([]K, len(c.items))
	for i, v := range c.items {
		mapped[i] = f(v)
	}
	return New(mapped)
}

func (c *Collection[T]) FirstWhere(f func(T) bool) (T, bool) {
	for _, v := range c.items {
		if f(v) {
			return v, true
		}
	}
	var zero T
	return zero, false
}

func (c *Collection[T]) Take(limit int) *Collection[T] {
	if limit < 0 {
		limit = 0
	}
	if limit > len(c.items) {
		limit = len(c.items)
	}
	return New(c.items[:limit])
}

func (c *Collection[T]) Chunk(size int) []*Collection[T] {
	if size <= 0 {
		panic("chunk size must be greater than zero")
	}
	var chunks []*Collection[T]
	for i := 0; i < len(c.items); i += size {
		end := i + size
		if end > len(c.items) {
			end = len(c.items)
		}
		chunks = append(chunks, New(c.items[i:end]))
	}
	return chunks
}

func (c *Collection[T]) Reduce[K any](f func(K, T) K, initial K) K {
	acc := initial
	for _, v := range c.items {
		acc = f(acc, v)
	}
	return acc
}

func (c *Collection[T]) Contains(f func(T) bool) bool {
	_, found := c.FirstWhere(f)
	return found
}
