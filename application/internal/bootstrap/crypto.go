package bootstrap

import (
	"fmt"

	"github.com/knadh/koanf/v2"
	"github.com/samber/do/v2"
	"github.com/studiolambda/cosmos/contract"
	"github.com/studiolambda/cosmos/framework/crypto"
)

func NewCrypto(i do.Injector) (contract.Encrypter, error) {
	k := do.MustInvoke[*koanf.Koanf](i)

	driver := k.String("crypto.driver")

	switch driver {
	case "aes":
		config := do.MustInvoke[crypto.AESConfig](i)

		return crypto.NewAES(config)
	case "chacha20":
		config := do.MustInvoke[crypto.ChaCha20Config](i)

		return crypto.NewChaCha20(config)
	}

	return nil, fmt.Errorf("unknown crypto driver %q", driver)
}
