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
var ErrTargetNotCallable = errors.New("target not callable")
var ErrInvalidOutNum = errors.New("invalid number of return values")

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

func (c *Container) Make[T any](f any) (t T, e error) {
	rv := reflect.ValueOf(f)

	if rv.Kind() != reflect.Func {
		return t, fmt.Errorf("%w: expected func but got %T", ErrTargetNotCallable, f)
	}

	rt := rv.Type()

	outs := rt.NumOut()
	if outs < 1 || outs > 2 {
		return t, fmt.Errorf("%w: expected 1 or 2 but got %d", ErrInvalidOutNum, outs)
	}

	target := reflect.TypeFor[T]()

	if !rt.Out(0).AssignableTo(target) {
		return t, fmt.Errorf(
			"%w: expected %v but got %v",
			ErrInvalidConversion,
			target,
			rt.Out(0),
		)
	}

	if outs == 2 && !rt.Out(1).Implements(reflect.TypeFor[error]()) {
		return t, fmt.Errorf(
			"%w: second return must be error, got %v",
			ErrInvalidConversion,
			rt.Out(1),
		)
	}

	args := make([]reflect.Value, rt.NumIn())

	for i := range args {
		arg, err := c.ResolveType(rt.In(i))
		if err != nil {
			return t, err
		}
		args[i] = arg
	}

	results := rv.Call(args)

	if outs == 2 && !results[1].IsNil() {
		return t, results[1].Interface().(error)
	}

	return results[0].Interface().(T), nil
}

func (c *Container) MustMake[T any](f any) T {
	r, err := c.Make[T](f)

	if err != nil {
		panic(err)
	}

	return r
}

func (c *Container) ResolveType(t reflect.Type) (reflect.Value, error) {
	k := keyFromType(t)

	c.mutex.RLock()
	r, ok := c.resolvers[k]
	c.mutex.RUnlock()

	if !ok {
		return reflect.Value{}, fmt.Errorf("%w: %s", ErrServiceNotFound, k)
	}

	v, err := r(c) // no container lock held
	if err != nil {
		return reflect.Value{}, err
	}

	rv := reflect.ValueOf(v)

	if !rv.IsValid() {
		return reflect.Value{}, fmt.Errorf("%w: resolver returned nil", ErrInvalidConversion)
	}

	if !rv.Type().AssignableTo(t) {
		return reflect.Value{}, fmt.Errorf(
			"%w from %v to %v",
			ErrInvalidConversion,
			rv.Type(),
			t,
		)
	}

	return rv, nil
}

func (c *Container) Resolve[T any]() (t T, err error) {
	v, err := c.ResolveType(reflect.TypeFor[T]())

	if err != nil {
		return t, err
	}

	return v.Interface().(T), nil
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
