package config

import (
	"github.com/knadh/koanf/v2"
	"github.com/samber/do/v2"
	"github.com/studiolambda/cosmos/framework/cache"
)

func NewMemoryCache(i do.Injector) (cache.MemoryConfig, error) {
	k := do.MustInvoke[*koanf.Koanf](i)

	config := cache.MemoryConfig{
		Expiration: k.Duration("cache.memory.expiration"),
		Cleanup:    k.Duration("cache.memory.cleanup"),
	}

	return config, nil
}

func NewRedisCache(i do.Injector) (cache.RedisConfig, error) {
	k := do.MustInvoke[*koanf.Koanf](i)

	config := cache.RedisConfig{
		Network:  k.String("cache.redis.network"),
		Addr:     k.String("cache.redis.addr"),
		Username: k.String("cache.redis.username"),
		Password: k.String("cache.redis.password"),
		DB:       k.Int("cache.redis.db"),
	}

	return config, nil
}
