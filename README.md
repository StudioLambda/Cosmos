# Cosmos

Cosmos is a lightweight, powerful, and flexible collection of http modules designed to work well together while maintaining great compatibility with the standard library. In fact,
the modules are designed to be a thin wrapper on top of them, without reinventing the wheel.

## Packages

- **router**: A HTTP router based on top of `http.ServeMux` with a familiar API.
- **framework**: A HTTP framework based on a modified version of a `http.Handler` that includes support for error handling, middlewares, helpers and more.
- **problem**: A pure go implementation of the [Problem's API](https://datatracker.ietf.org/doc/html/rfc9457) that works well in any HTTP framework.
- **contract**: A collection of common used service interfaces such as databases or caches.
- **service/\***: A collection of services that implement the contracts.

## Local Development

use go workspaces to manage the local development.

```sh
go work init
go work use ./framework ./problem ./router ./contract ./service/cache/memory ./service/cache/redis ./service/database/sql
```

Then replace deps as needed:

```sh
go work edit -replace=github.com/studiolambda/cosmos/router@vX.X.X=./router
go work edit -replace=github.com/studiolambda/cosmos/framework@vX.X.X=./framework
```
