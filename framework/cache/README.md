# framework/cache

Concrete raw-byte cache backends implementing `contract.CacheDriver`.

## Implementations

- `cache/memory`: in-process cache backed by `go-cache`.
- `cache/redis`: Redis-backed cache driver.

## When to use it

Use `Memory` for local development, tests, and single-node deployments.
Use `Redis` for shared cache state across processes/instances.

Wrap a backend with `contract.NewCache` for typed JSON values and helpers such
as `Remember`.
