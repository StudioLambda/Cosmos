# framework/middleware

Built-in middleware for `github.com/studiolambda/cosmos/framework`.

## Included middleware

- Panic recovery (`Recover`, `RecoverWith`)
- Structured logging (`Logger`)
- CORS and CSRF protection
- Secure response headers
- Rate limiting
- Context value injection (`Provide`, `ProvideWith`)
- Adapter for standard `func(http.Handler) http.Handler` middleware (`HTTP`)

## Ordering guidance

Place recovery and logging near the beginning of the chain so downstream
failures are consistently captured.

`middleware.HTTP` adapts standard middleware; `framework.HTTP` adapts a
standard `http.Handler` for a Cosmos route.
