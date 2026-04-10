package hash

import "github.com/matthewhartstonge/argon2"

// Argon2Config is an alias for argon2.Config, exposing the full set
// of tuning parameters (memory, iterations, parallelism) without
// requiring a direct import of the argon2 package.
type Argon2Config = argon2.Config

// Argon2 implements contract.Hasher using the Argon2id algorithm.
// It is the recommended hasher for password storage due to its
// resistance to GPU and side-channel attacks.
type Argon2 struct {
	config argon2.Config
}

// NewArgon2 creates an Argon2 hasher with the library's default
// configuration (Argon2id, sensible memory/iteration defaults).
func NewArgon2() *Argon2 {
	return NewArgon2With(argon2.DefaultConfig())
}

// NewArgon2With creates an Argon2 hasher using the provided
// configuration, allowing full control over memory cost,
// iteration count, and parallelism.
func NewArgon2With(config Argon2Config) *Argon2 {
	return &Argon2{
		config: config,
	}
}

// Hash produces an Argon2id encoded hash of the given value.
// The returned byte slice contains the full encoded string
// including algorithm parameters and salt. The plaintext value
// is zeroed from memory after hashing completes.
func (hasher *Argon2) Hash(value []byte) ([]byte, error) {
	defer zeroBytes(value)

	return hasher.config.HashEncoded(value)
}

// Check verifies that the given plaintext value matches the
// previously hashed encoded output. The plaintext value is
// zeroed from memory after verification completes.
func (hasher *Argon2) Check(value []byte, hash []byte) (bool, error) {
	defer zeroBytes(value)

	return argon2.VerifyEncoded(value, hash)
}
