package hash

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

// Bcrypt implements contract.Hasher using the bcrypt algorithm.
// It is an acceptable alternative to Argon2 when compatibility
// with existing bcrypt hashes is required.
//
// WARNING: Bcrypt hash operations are intentionally CPU intensive.
// Endpoints that trigger hashing (login, registration, password
// change) should implement rate limiting to prevent
// denial-of-service attacks through hash computation abuse.
type Bcrypt struct {
	options BcryptOptions
}

// BcryptOptions configures the bcrypt hasher. The cost parameter
// controls the computational expense of hashing; higher values
// are more secure but slower.
type BcryptOptions struct {
	cost int
}

// DefaultBcryptCost is the default bcrypt cost factor, matching
// the bcrypt library's own default of 10.
const DefaultBcryptCost = 10

// NewBcrypt creates a Bcrypt hasher with the default cost factor.
func NewBcrypt() *Bcrypt {
	return NewBcryptWith(BcryptOptions{
		cost: DefaultBcryptCost,
	})
}

// NewBcryptWith creates a Bcrypt hasher with the given options,
// allowing a custom cost factor.
func NewBcryptWith(options BcryptOptions) *Bcrypt {
	return &Bcrypt{
		options: options,
	}
}

// Hash produces a bcrypt hash of the given value using the
// configured cost factor.
func (hasher *Bcrypt) Hash(value []byte) ([]byte, error) {
	hash, err := bcrypt.GenerateFromPassword(value, hasher.options.cost)

	if err != nil {
		return nil, err
	}

	return hash, nil
}

// Check verifies that the given plaintext value matches the
// bcrypt hash. Returns (false, nil) on a mismatch and (false, err)
// on an unexpected error.
func (hasher *Bcrypt) Check(value []byte, hash []byte) (bool, error) {
	err := bcrypt.CompareHashAndPassword(hash, value)

	if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	return true, nil
}
