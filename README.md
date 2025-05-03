# Cosmos

Cosmos is a lightweight, powerful, and flexible collection of http modules designed to work well together while maintaining great compatibility with the standard library. In fact,
the modules are designed to be a thin wrapper on top of them, without reinventing the wheel.

## Packages

- **Orbit**: A HTTP router based on top of `http.ServeMux` with a familiar API.
- **Nova**: A HTTP framework based on a modified version of a `http.Handler` that includes support for error handling, middlewares, helpers and more.
- **Atlas**: An application layer for Nova, it bootstraps the app, handles graceful shutdown and more.
- **Fracture**: A pure go implementation of the [Problem's API](https://datatracker.ietf.org/doc/html/rfc9457) that works well in any HTTP framework.

## Local Development

use go workspaces to manage the local development.

```sh
go work init
go work use ./nova ./fracture ./orbit ./atlas
```

Then replace deps as needed:

```sh
go work edit -replace=github.com/studiolambda/cosmos/orbit@vX.X.X=./orbit
go work edit -replace=github.com/studiolambda/cosmos/nova@vX.X.X=./nova
```
