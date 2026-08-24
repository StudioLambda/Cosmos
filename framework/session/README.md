# framework/session

Cache-backed `contract.SessionDriver` persistence. HTTP middleware is in
`framework/middleware`.

## What this package provides

- Cache-backed `contract.SessionDriver` implementation.
- `NewCacheDriver(cache, config)` with a `*contract.Cache` facade.

## Behavior summary

- Sessions are loaded per request and persisted when changed.
- Regenerated session IDs cause old IDs to be deleted.
- Absolute session lifetime is enforced independently from sliding TTL.

Install `middleware.Session(driver, middleware.DefaultSessionConfig())` to load
and persist a session for each request. Retrieve it with `request.Session(r)`,
which returns `(*contract.Session, bool)`; `request.MustSession(r)` panics when
it is missing.

## Security notes

Regenerate session IDs after authentication and privilege changes.
