package contract

// Hasher defines the interface for hashing and verifying hashed values.
// Implementations of Hasher are responsible for generating cryptographic hashes
// and verifying that values match their corresponding hashes.
type Hasher interface {
	// Hash computes a cryptographic hash of the given byte slice and returns the hash.
	// It returns an error if the hashing operation fails.
	Hash(value []byte) ([]byte, error)

	// Check verifies that the given value matches the provided hash.
	// It returns true if the value and hash match, false otherwise.
	// It returns an error if the verification operation fails.
	Check(value []byte, hash []byte) (bool, error)
}
