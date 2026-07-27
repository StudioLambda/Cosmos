package bootstrap

import (
	"fmt"

	"github.com/knadh/koanf/v2"
	"github.com/samber/do/v2"
	"github.com/studiolambda/cosmos/contract"
	"github.com/studiolambda/cosmos/framework/cache"
)

func NewCache(i do.Injector) (*contract.Cache, error) {
	driver := do.MustInvoke[contract.CacheDriver](i)

	return contract.NewCache(driver), nil
}

func NewCacheDriver(i do.Injector) (contract.CacheDriver, error) {
	k := do.MustInvoke[*koanf.Koanf](i)

	driver := k.String("cache.driver")

	switch driver {
	case "memory":
		config := do.MustInvoke[cache.MemoryConfig](i)

		return cache.NewMemory(config), nil
	case "redis":
		config := do.MustInvoke[cache.RedisConfig](i)

		return cache.NewRedis(config), nil
	}

	return nil, fmt.Errorf("unknown cache driver %q", driver)
}
