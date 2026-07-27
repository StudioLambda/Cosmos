package bootstrap

import (
	"fmt"

	"github.com/knadh/koanf/v2"
	"github.com/samber/do/v2"
	"github.com/studiolambda/cosmos/contract"
	"github.com/studiolambda/cosmos/framework/hash"
)

func NewHasher(i do.Injector) (contract.Hasher, error) {
	k := do.MustInvoke[*koanf.Koanf](i)

	driver := k.String("hash.driver")

	switch driver {
	case "bcrypt":
		config := do.MustInvoke[hash.BcryptConfig](i)

		return hash.NewBcrypt(config), nil
	case "argon2":
		config := do.MustInvoke[hash.Argon2Config](i)

		return hash.NewArgon2(config), nil
	}

	return nil, fmt.Errorf("unknown hash driver %q", driver)
}
