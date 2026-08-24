# contract/mock

Generated test doubles for interfaces in `github.com/studiolambda/cosmos/contract`.

## Purpose

This package is for tests only. It provides mocks for cache, database,
session, events, crypto, hash, and hooks contracts.

## Generation

Mocks are generated from contract interfaces via `mockery` configuration in
`contract/.mockery.yml`.

```bash
cd contract
go generate ./...
```
