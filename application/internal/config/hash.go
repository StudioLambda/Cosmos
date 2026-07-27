package config

import (
	"github.com/knadh/koanf/v2"
	"github.com/matthewhartstonge/argon2"
	"github.com/samber/do/v2"
	"github.com/studiolambda/cosmos/framework/hash"
)

func NewBcryptHash(i do.Injector) (hash.BcryptConfig, error) {
	k := do.MustInvoke[*koanf.Koanf](i)

	config := hash.BcryptConfig{
		Cost: k.Int("hash.bcrypt.cost"),
	}

	return config, nil
}

func NewArgon2Hash(i do.Injector) (hash.Argon2Config, error) {
	k := do.MustInvoke[*koanf.Koanf](i)

	config := hash.Argon2Config{
		HashLength:  uint32(k.Int("hash.argon2.hash_length")),
		SaltLength:  uint32(k.Int("hash.argon2.salt_length")),
		TimeCost:    uint32(k.Int("hash.argon2.time_cost")),
		MemoryCost:  uint32(k.Int("hash.argon2.memory_cost")),
		Parallelism: uint8(k.Int("hash.argon2.parallelism")),
		Mode:        argon2.Mode(k.Int("hash.argon2.mode")),
		Version:     argon2.Version(k.Int("hash.argon2.version")),
	}

	return config, nil
}
