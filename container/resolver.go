package container

import "sync"

type Resolver[T any] func(c *Container) (T, error)

func NewResolver[T any](r Resolver[T]) Resolver[any] {
	return func(c *Container) (any, error) {
		return r(c)
	}
}

func NewSingleton(v any) Resolver[any] {
	return func(c *Container) (any, error) {
		return v, nil
	}
}

func NewLazySingleton[T any](r Resolver[T]) Resolver[any] {
	var once sync.Once
	var singleton T
	var resolveErr error

	return func(c *Container) (any, error) {
		once.Do(func() {
			singleton, resolveErr = r(c)
		})

		if resolveErr != nil {
			var zero T

			return zero, resolveErr
		}

		return singleton, nil
	}
}
