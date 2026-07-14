# framework/event

Event bus adapters implementing `contract.EventBus`.

## Implementations

- Memory (in-process)
- Redis
- NATS
- AMQP
- MQTT

## Topic matching

The package includes wildcard topic matching helpers used by memory and broker
adapters (`*` for one segment, `#` for many).

## When to use it

Use `Memory` in tests and local-only deployments. Use broker-backed adapters for
multi-process distribution.
