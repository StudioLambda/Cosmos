package contract

// Hasher defines the interface for hashing and verifying hashed values.
// Implementations of Hasher are responsible for generating cryptographic hashes
// and verifying that values match their corresponding hashes.
//
// Both Hash and Check zero the input value slice after use as a security
// measure. Callers must not reuse the value slice after calling either method.
type Hasher interface {
	// Hash computes a cryptographic hash of the given byte slice and returns the hash.
	// The input value is zeroed after hashing as a security measure.
	// Callers must not reuse the value slice after calling Hash.
	// It returns an error if the hashing operation fails.
	Hash(value []byte) ([]byte, error)

	// Check verifies that the given value matches the provided hash.
	// The input value is zeroed after verification as a security measure.
	// Callers must not reuse the value slice after calling Check.
	// It returns true if the value and hash match, false otherwise.
	// It returns an error if the verification operation fails.
	Check(value []byte, hash []byte) (bool, error)
}

// Rehashable extends [Hasher] with the ability to detect stale hash parameters.
// Implementations should return true when the given hash was produced with
// different parameters than the current configuration, indicating the value
// should be re-hashed on the next successful authentication.
type Rehashable interface {
	NeedsRehash(hash []byte) bool
}
