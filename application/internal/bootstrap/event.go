package bootstrap

import (
	"fmt"

	"github.com/knadh/koanf/v2"
	"github.com/samber/do/v2"
	"github.com/studiolambda/cosmos/contract"
	"github.com/studiolambda/cosmos/framework/event"
)

func NewEvents(i do.Injector) (*contract.Events, error) {
	driver := do.MustInvoke[contract.EventDriver](i)

	return contract.NewEvents(driver), nil
}

func NewEventDriver(i do.Injector) (contract.EventDriver, error) {
	k := do.MustInvoke[*koanf.Koanf](i)

	driver := k.String("event.driver")

	switch driver {
	case "memory":
		config := do.MustInvoke[event.MemoryBrokerConfig](i)

		return event.NewMemoryBroker(config), nil
	case "amqp":
		config := do.MustInvoke[event.AMQPBrokerConfig](i)

		return event.NewAMQPBroker(config)
	case "mqtt":
		config := do.MustInvoke[event.MQTTBrokerConfig](i)

		return event.NewMQTTBroker(config)
	case "nats":
		config := do.MustInvoke[event.NATSBrokerConfig](i)

		return event.NewNATSBroker(config)
	case "redis":
		config := do.MustInvoke[event.RedisBrokerConfig](i)

		return event.NewRedisBroker(config), nil
	}

	return nil, fmt.Errorf("unknown event driver %q", driver)
}
