// Package collection provides generic eager and lazy slice and map collections.
//
// # Eager collections
//
// [Slice] and [Map] wrap in-memory Go slices and maps. Their constructors and
// Items methods convert without copying, so the wrapper and the supplied or
// returned collection alias the same storage. Mutating either is observable
// through the other. Operations that allocate a result return independent
// storage; operations that select a portion of a Slice can retain its backing
// array.
//
// # Lazy collections
//
// [LazySlice] and [LazyMap] are iterator function types. Transformations such
// as Filter and Map compose lazy collections and run when the result is
// iterated. They can be ranged over directly, and Items or Eager materializes
// all yielded values into a Go collection or eager wrapper. A lazy collection
// invokes its underlying iterator for each traversal. It is replayable only
// when that iterator is replayable; callers must not assume that a traversal
// can be repeated or that it has no side effects.
//
// LazySlice callbacks receive the index from their source sequence. LazySlice
// transformations yield dense output indices, starting at zero, even when
// filtering, expanding, reordering, or combining source items.
//
// [TryLazySlice] and [TryLazyMap] are the error-aware lazy forms. They yield a
// value and an error, and a value yielded with a non-nil error must be ignored.
// Their Items and Eager methods stop at the first error and return values
// collected before it. ItemsAll and EagerAll consume the full sequence, retain
// successful values, and join yielded errors. Like the non-error lazy forms,
// their iteration and replayability depend on their underlying iterator.
//
// Slice callbacks receive an index, while map callbacks receive keys and
// values. Map iteration order follows the underlying iterator and is not
// guaranteed for Go maps.
package collection
