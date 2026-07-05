package container

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"sync"
)

type containerKey struct{}

type Container struct {
	mutex     sync.RWMutex
	resolvers map[string]Resolver[any]
}

var ContainerKey containerKey
var ErrServiceNotFound = errors.New("service not found in container")
var ErrInvalidConversion = errors.New("invalid conversion")
var ErrCallableNotFunction = errors.New("callable must be a function")
var ErrVariadicCallableNotSupported = errors.New("variadic functions are not supported")

func NewContainer() *Container {
	return &Container{
		resolvers: make(map[string]Resolver[any]),
	}
}

func Must(c *Container, ok bool) *Container {
	if !ok {
		panic("unable to resolve container")
	}

	return c
}

func FromContext(ctx context.Context) (*Container, bool) {
	if c, ok := ctx.Value(ContainerKey).(*Container); ok {
		return c, true
	}

	return nil, false
}

func FromRequest(r *http.Request) (*Container, bool) {
	return FromContext(r.Context())
}

func Key[T any]() string {
	return keyFromType(reflect.TypeFor[T]())
}

func keyFromType(reflected reflect.Type) string {
	if reflected == nil {
		return ""
	}

	if reflected.Name() != "" {
		return reflected.PkgPath() + "." + reflected.Name()
	}

	switch reflected.Kind() {
	case reflect.Pointer:
		return "*" + keyFromType(reflected.Elem())
	case reflect.Slice:
		return "[]" + keyFromType(reflected.Elem())
	case reflect.Array:
		return fmt.Sprintf("[%d]%s", reflected.Len(), keyFromType(reflected.Elem()))
	case reflect.Map:
		return "map[" + keyFromType(reflected.Key()) + "]" + keyFromType(reflected.Elem())
	default:
		return reflected.String()
	}
}

func (c *Container) Context(ctx context.Context) context.Context {
	return context.WithValue(ctx, ContainerKey, c)
}

func (c *Container) MustResolve[T any]() T {
	v, err := c.Resolve[T]()

	if err != nil {
		panic(err)
	}

	return v
}

func (c *Container) Resolve[T any]() (t T, e error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	k := Key[T]()

	if r, ok := c.resolvers[k]; ok {
		v, err := r(c)

		if err != nil {
			return t, err
		}

		if res, ok := v.(T); ok {
			return res, nil
		}

		return t, fmt.Errorf("%w from %T to %T", ErrInvalidConversion, v, t)
	}

	return t, fmt.Errorf("%w: %s", ErrServiceNotFound, k)
}

func (c *Container) Register[T any](resolver Resolver[T]) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.resolvers[Key[T]()] = NewResolver(resolver)
}

func (c *Container) Singleton[T any](value any) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.resolvers[Key[T]()] = NewSingleton(value)
}

func (c *Container) LazySingleton[T any](resolver Resolver[T]) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.resolvers[Key[T]()] = NewLazySingleton(resolver)
}
