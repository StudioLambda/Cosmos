package contract

// correlationIDKey is a private type used as a context key to avoid collisions.
type correlationIDKey struct{}

// CorrelationIDKey is the context key used to store and retrieve
// the correlation ID from a context.Context.
var CorrelationIDKey = correlationIDKey{}
