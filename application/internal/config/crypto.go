package config

import (
	"github.com/knadh/koanf/v2"
	"github.com/samber/do/v2"
	"github.com/studiolambda/cosmos/framework/crypto"
)

func NewAES(i do.Injector) (crypto.AESConfig, error) {
	k := do.MustInvoke[*koanf.Koanf](i)

	config := crypto.AESConfig{
		Key: k.Bytes("crypto.aes.key"),
	}

	return config, nil
}

func NewChaCha20(i do.Injector) (crypto.ChaCha20Config, error) {
	k := do.MustInvoke[*koanf.Koanf](i)

	config := crypto.ChaCha20Config{
		Key: k.Bytes("crypto.chacha20.key"),
	}

	return config, nil
}
