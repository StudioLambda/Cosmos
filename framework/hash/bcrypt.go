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

// DefaultBcryptCost is the default bcrypt cost factor.
// OWASP recommends a minimum cost of 12 for password hashing
// to provide adequate resistance against brute-force attacks.
const DefaultBcryptCost = 12

// NewBcrypt creates a Bcrypt hasher with the default cost factor.
func NewBcrypt() *Bcrypt {
	return NewBcryptWith(BcryptOptions{
		cost: DefaultBcryptCost,
	})
}

// NewBcryptWith creates a Bcrypt hasher with the given options,
// allowing a custom cost factor. The cost is clamped to a minimum
// of bcrypt.MinCost (4). Costs below 12 are not recommended for
// production use per OWASP guidelines.
func NewBcryptWith(options BcryptOptions) *Bcrypt {
	if options.cost < bcrypt.MinCost {
		options.cost = bcrypt.MinCost
	}

	return &Bcrypt{
		options: options,
	}
}

// Hash produces a bcrypt hash of the given value using the
// configured cost factor. The plaintext value is zeroed from
// memory after hashing completes.
func (hasher *Bcrypt) Hash(value []byte) ([]byte, error) {
	defer zeroBytes(value)

	hash, err := bcrypt.GenerateFromPassword(value, hasher.options.cost)

	if err != nil {
		return nil, err
	}

	return hash, nil
}

// Check verifies that the given plaintext value matches the
// bcrypt hash. Returns (false, nil) on a mismatch and (false, err)
// on an unexpected error. The plaintext value is zeroed from
// memory after verification completes.
func (hasher *Bcrypt) Check(value []byte, hash []byte) (bool, error) {
	defer zeroBytes(value)

	err := bcrypt.CompareHashAndPassword(hash, value)

	if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	return true, nil
}
