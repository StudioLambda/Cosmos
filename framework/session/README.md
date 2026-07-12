# framework/session

Session middleware and cache-backed session persistence.

## What this package provides

- HTTP middleware for loading, committing, and rotating sessions.
- Cache-backed `contract.SessionDriver` implementation.
- Cookie/session lifecycle configuration.

## Behavior summary

- Sessions are loaded per request and persisted when changed.
- Regenerated session IDs cause old IDs to be deleted.
- Absolute session lifetime is enforced independently from sliding TTL.

## Security notes

Regenerate session IDs after authentication and privilege changes.
