# framework/event

Event bus adapters implementing `contract.EventBus`.

## Implementations

- `event/memory` (in-process)
- `event/redis`
- `event/nats`
- `event/amqp`
- `event/mqtt`

## Topic matching

Each implementation owns its transport-specific topic matching behavior.

## When to use it

Use `Memory` in tests and local-only deployments. Use broker-backed adapters for
multi-process distribution.
