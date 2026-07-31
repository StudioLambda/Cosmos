package bootstrap

import (
	"github.com/studiolambda/cosmos/contract"
	"github.com/studiolambda/cosmos/framework/event/memory"
)

// NewEvents creates an in-memory event bus from event.memory configuration.
func NewEvents(configuration *contract.Configuration) *contract.Events {
	config := configuration.From[memory.MemoryBrokerConfig]("event.memory")
	driver := memory.NewMemoryBroker(config)

	return contract.NewEvents(driver)
}
