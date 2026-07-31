// Package argon2 provides an Argon2id [contract.Hasher].
package argon2

import (
	"github.com/matthewhartstonge/argon2"
	"github.com/studiolambda/cosmos/contract"
	"github.com/studiolambda/cosmos/framework/hash/internal/secure"
)

// Argon2Config mirrors argon2.Config, exposing the full set
// of tuning parameters (memory, iterations, parallelism) without
// requiring a direct import of the argon2 package.
type Argon2Config argon2.Config

// Argon2 implements contract.Hasher using the Argon2id algorithm.
// It is the recommended hasher for password storage due to its
// resistance to GPU and side-channel attacks.
//
// WARNING: Argon2 hash operations are intentionally CPU and memory
// intensive. Endpoints that trigger hashing (login, registration,
// password change) should implement rate limiting to prevent
// denial-of-service attacks through hash computation abuse.
type Argon2 struct {
	config argon2.Config
}

// DefaultArgon2Config returns the default Argon2 hasher configuration.
func DefaultArgon2Config() Argon2Config {
	return Argon2Config(argon2.DefaultConfig())
}

// FromConfiguration populates the Argon2 configuration from configuration.
func (config *Argon2Config) FromConfiguration(configuration *contract.Configuration) {
	config.HashLength = uint32(configuration.GetOr("hash_length", 0))
	config.SaltLength = uint32(configuration.GetOr("salt_length", 0))
	config.TimeCost = uint32(configuration.GetOr("time_cost", 0))
	config.MemoryCost = uint32(configuration.GetOr("memory_cost", 0))
	config.Parallelism = uint8(configuration.GetOr("parallelism", 0))
	config.Mode = argon2.Mode(configuration.GetOr("mode", 0))
	config.Version = argon2.Version(configuration.GetOr("version", 0))
}

// NewArgon2 creates an Argon2 hasher using the provided
// configuration, allowing full control over memory cost,
// iteration count, and parallelism.
func NewArgon2(config Argon2Config) *Argon2 {
	return &Argon2{
		config: argon2.Config(config),
	}
}

// Hash produces an Argon2id encoded hash of the given value.
// The returned byte slice contains the full encoded string
// including algorithm parameters and salt. The input value is
// zeroed after hashing as a security measure. Callers must not
// reuse the value slice after calling Hash.
func (hasher *Argon2) Hash(value []byte) ([]byte, error) {
	defer secure.Clear(value)

	return hasher.config.HashEncoded(value)
}

// Check verifies that the given plaintext value matches the
// previously hashed encoded output. The input value is zeroed
// after verification as a security measure. Callers must not
// reuse the value slice after calling Check.
func (hasher *Argon2) Check(value []byte, hash []byte) (bool, error) {
	defer secure.Clear(value)

	return argon2.VerifyEncoded(value, hash)
}

// NeedsRehash reports whether the given hash was created with different
// parameters than the current configuration, indicating it should be
// re-hashed on the next successful authentication.
func (hasher *Argon2) NeedsRehash(hash []byte) bool {
	raw, err := argon2.Decode(hash)

	if err != nil {
		return true
	}

	return raw.Config.TimeCost != hasher.config.TimeCost ||
		raw.Config.MemoryCost != hasher.config.MemoryCost ||
		raw.Config.Parallelism != hasher.config.Parallelism ||
		raw.Config.HashLength != hasher.config.HashLength
}
