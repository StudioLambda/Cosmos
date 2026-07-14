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
	config BcryptConfig
}

// BcryptConfig configures the bcrypt hasher. The Cost parameter
// controls the computational expense of hashing; higher values
// are more secure but slower.
type BcryptConfig struct {
	Cost int
}

// DefaultBcryptCost is the default bcrypt cost factor.
// OWASP recommends a minimum cost of 12 for password hashing
// to provide adequate resistance against brute-force attacks.
const DefaultBcryptCost = 12

// NewBcrypt creates a Bcrypt hasher with the default cost factor.
func NewBcrypt() *Bcrypt {
	return NewBcryptWith(BcryptConfig{
		Cost: DefaultBcryptCost,
	})
}

// NewBcryptWith creates a Bcrypt hasher with the given configuration,
// allowing a custom cost factor. The cost is clamped to a minimum
// of bcrypt.MinCost (4). Costs below 12 are not recommended for
// production use per OWASP guidelines.
func NewBcryptWith(config BcryptConfig) *Bcrypt {
	if config.Cost < bcrypt.MinCost {
		config.Cost = bcrypt.MinCost
	}

	return &Bcrypt{
		config: config,
	}
}

// Hash generates a bcrypt hash of the given value.
// Bcrypt has a maximum password length of 72 bytes. Passwords exceeding
// this limit will result in an error from the underlying bcrypt library.
// The input value is zeroed after hashing as a security measure.
// Callers must not reuse the value slice after calling Hash.
func (hasher *Bcrypt) Hash(value []byte) ([]byte, error) {
	defer zeroBytes(value)

	hash, err := bcrypt.GenerateFromPassword(value, hasher.config.Cost)

	if err != nil {
		return nil, err
	}

	return hash, nil
}

// Check verifies that the given plaintext value matches the
// bcrypt hash. Returns (false, nil) on a mismatch and (false, err)
// on an unexpected error. The input value is zeroed after verification
// as a security measure. Callers must not reuse the value slice after
// calling Check.
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

// NeedsRehash reports whether the given hash was created with different
// parameters than the current configuration, indicating it should be
// re-hashed on the next successful authentication.
func (hasher *Bcrypt) NeedsRehash(hash []byte) bool {
	cost, err := bcrypt.Cost(hash)

	if err != nil {
		return true
	}

	return cost != hasher.config.Cost
}
