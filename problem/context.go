package problem

import (
	"context"
	"maps"
	"sync"
)

type contextValuesKey struct{}

// ContextValues stores RFC 9457 extension members for a request.
// All methods are safe for concurrent use.
type ContextValues struct {
	mutex  sync.RWMutex
	values map[string]any
}

// NewContextValues creates an empty request problem-values container.
func NewContextValues() *ContextValues {
	return &ContextValues{values: make(map[string]any)}
}

// Add merges values into the container. Values with the same key replace
// previously added values.
func (contextValues *ContextValues) Add(values map[string]any) {
	if len(values) == 0 {
		return
	}

	contextValues.mutex.Lock()
	defer contextValues.mutex.Unlock()

	if contextValues.values == nil {
		contextValues.values = make(map[string]any, len(values))
	}

	maps.Copy(contextValues.values, values)
}

// Values returns a snapshot of the values in the container.
func (contextValues *ContextValues) Values() map[string]any {
	contextValues.mutex.RLock()
	defer contextValues.mutex.RUnlock()

	return maps.Clone(contextValues.values)
}

// WithContextValuesContainer returns a context containing the supplied
// request problem-values container.
func WithContextValuesContainer(ctx context.Context, values *ContextValues) context.Context {
	if values == nil {
		return ctx
	}

	return context.WithValue(ctx, contextValuesKey{}, values)
}

// ContextValuesFrom retrieves the request problem-values container from a
// context without panicking.
func ContextValuesFrom(ctx context.Context) (*ContextValues, bool) {
	values, ok := ctx.Value(contextValuesKey{}).(*ContextValues)

	return values, ok
}

// WithContextValues returns a context containing values to include as RFC 9457
// extension members when [Details.ServeHTTP] is called. Calls compose: values
// from later calls replace same-named values from earlier calls.
//
// Standard RFC 9457 member names are ignored when serving. Values explicitly
// added with [Details.With] take precedence over context values.
func WithContextValues(ctx context.Context, values map[string]any) context.Context {
	if len(values) == 0 {
		return ctx
	}

	contextValues, ok := ContextValuesFrom(ctx)
	if !ok {
		contextValues = NewContextValues()
		ctx = WithContextValuesContainer(ctx, contextValues)
	}

	contextValues.Add(values)

	return ctx
}

func isStandardMember(key string) bool {
	switch key {
	case "type", "title", "detail", "status", "instance":
		return true
	default:
		return false
	}
}

func (details Details) withContextValues(ctx context.Context) Details {
	contextValues, ok := ContextValuesFrom(ctx)
	if !ok {
		return details
	}

	values := contextValues.Values()
	if len(values) == 0 {
		return details
	}

	additional := maps.Clone(details.Additional)
	if additional == nil {
		additional = make(map[string]any, len(values))
	}

	for key, value := range values {
		if isStandardMember(key) {
			continue
		}

		if _, ok := additional[key]; ok {
			continue
		}

		additional[key] = value
	}

	details.Additional = additional

	return details
}
