# framework/cache

Concrete cache backends implementing `contract.CacheDriver`.

## Implementations

- `Memory`: in-process cache backed by `go-cache`.
- `Redis`: Redis-backed cache driver.

## When to use it

Use `Memory` for local development, tests, and single-node deployments.
Use `Redis` for shared cache state across processes/instances.
