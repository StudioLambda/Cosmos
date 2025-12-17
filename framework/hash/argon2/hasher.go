package argon2

import "github.com/matthewhartstonge/argon2"

type Config = argon2.Config

type Hasher struct {
	config argon2.Config
}

func NewHasher() *Hasher {
	return NewHasherWith(argon2.DefaultConfig())
}

func NewHasherWith(config Config) *Hasher {
	return &Hasher{
		config: config,
	}
}

func (h *Hasher) Hash(value []byte) ([]byte, error) {
	return h.config.HashEncoded(value)
}

func (h *Hasher) Check(value []byte, hash []byte) (bool, error) {
	return argon2.VerifyEncoded(value, hash)
}
