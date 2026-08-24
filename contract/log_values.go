package contract

import (
	"context"
	"maps"
	"sync"
)

type logValuesKey struct{}

// LogValues stores structured logging values for a request.
// All methods are safe for concurrent use.
type LogValues struct {
	mutex  sync.RWMutex
	values map[string]any
}

// NewLogValues creates an empty request log-values container.
func NewLogValues() *LogValues {
	return &LogValues{values: make(map[string]any)}
}

// WithLogValuesContainer returns a context containing the supplied request
// log-values container.
func WithLogValuesContainer(ctx context.Context, values *LogValues) context.Context {
	if values == nil {
		return ctx
	}

	return context.WithValue(ctx, logValuesKey{}, values)
}

// Add merges values into the container. Values with the same key replace
// previously added values.
func (logValues *LogValues) Add(values map[string]any) {
	if len(values) == 0 {
		return
	}

	logValues.mutex.Lock()
	defer logValues.mutex.Unlock()

	maps.Copy(logValues.values, values)
}

// Values returns a snapshot of the values in the container.
func (logValues *LogValues) Values() map[string]any {
	logValues.mutex.RLock()
	defer logValues.mutex.RUnlock()

	return maps.Clone(logValues.values)
}

// WithLogValues returns a context containing values to add to request logs.
// Calls compose: values from later calls replace same-named values from earlier
// calls.
func WithLogValues(ctx context.Context, values map[string]any) context.Context {
	if len(values) == 0 {
		return ctx
	}

	logValues, ok := LogValuesFrom(ctx)
	if !ok {
		logValues = NewLogValues()
		ctx = WithLogValuesContainer(ctx, logValues)
	}

	logValues.Add(values)

	return ctx
}

// LogValuesFrom retrieves the request log-values container from a context
// without panicking.
func LogValuesFrom(ctx context.Context) (*LogValues, bool) {
	logValues, ok := ctx.Value(logValuesKey{}).(*LogValues)

	return logValues, ok
}
