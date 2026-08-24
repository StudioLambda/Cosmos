package collection

// MapEntry represents a key-value pair yielded by a [TryLazyMap].
type MapEntry[K comparable, V any] struct {
	Key   K
	Value V
}

// NewMapEntry creates a new MapEntry with the given key and value.
func NewMapEntry[K comparable, V any](key K, value V) MapEntry[K, V] {
	return MapEntry[K, V]{Key: key, Value: value}
}
