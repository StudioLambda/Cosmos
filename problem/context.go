package problem

import (
	"context"
	"maps"
)

type contextValuesKey struct{}

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

	merged, _ := ctx.Value(contextValuesKey{}).(map[string]any)
	merged = maps.Clone(merged)

	if merged == nil {
		merged = make(map[string]any, len(values))
	}

	maps.Copy(merged, values)

	return context.WithValue(ctx, contextValuesKey{}, merged)
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
	values, _ := ctx.Value(contextValuesKey{}).(map[string]any)
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
