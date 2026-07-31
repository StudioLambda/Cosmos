package bootstrap

import (
	"github.com/studiolambda/cosmos/contract"
	"github.com/studiolambda/cosmos/framework/cache/memory"
)

// NewCache creates an in-memory cache from cache.memory configuration.
func NewCache(configuration *contract.Configuration) (*contract.Cache, error) {
	config := configuration.From[memory.MemoryConfig]("cache.memory")
	driver := memory.NewMemory(config)

	return contract.NewCache(driver), nil
}
