package contract

// correlationIDKey is a private type used as a context key to avoid collisions.
type correlationIDKey struct{}

// CorrelationIDKey is the context key used to store and retrieve
// the correlation ID from a context.Context.
//
// Example:
//
//	ctx := context.WithValue(context.Background(), contract.CorrelationIDKey, "req-123")
//	id, _ := ctx.Value(contract.CorrelationIDKey).(string)
//	_ = id
var CorrelationIDKey = correlationIDKey{}
