package container

type Resolver[T any] = func(c *Container) (T, error)

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
	var singleton T
	var resolved bool

	return func(c *Container) (any, error) {
		if !resolved {
			v, err := r(c)

			if err != nil {
				return v, err
			}

			singleton = v
			resolved = true
		}

		return singleton, nil
	}
}
