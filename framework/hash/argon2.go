package hash

import "github.com/matthewhartstonge/argon2"

type Argon2Config = argon2.Config

type Argon2 struct {
	config argon2.Config
}

func NewArgon2() *Argon2 {
	return NewArgon2With(argon2.DefaultConfig())
}

func NewArgon2With(config Argon2Config) *Argon2 {
	return &Argon2{
		config: config,
	}
}

func (h *Argon2) Hash(value []byte) ([]byte, error) {
	return h.config.HashEncoded(value)
}

func (h *Argon2) Check(value []byte, hash []byte) (bool, error) {
	return argon2.VerifyEncoded(value, hash)
}
