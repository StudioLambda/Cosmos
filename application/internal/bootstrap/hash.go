package bootstrap

import (
	"github.com/studiolambda/cosmos/contract"
	"github.com/studiolambda/cosmos/framework/hash/argon2"
)

// NewHasher creates an Argon2id hasher from hash.argon2 configuration.
func NewHasher(configuration *contract.Configuration) contract.Hasher {
	config := configuration.From[argon2.Argon2Config]("hash.argon2")

	return argon2.NewArgon2(config)
}
