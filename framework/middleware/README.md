# framework/middleware

Built-in middleware for `github.com/studiolambda/cosmos/framework`.

## Included middleware

- Panic recovery (`Recover`, `RecoverWith`)
- Structured logging (`Logger`)
- CORS and CSRF protection
- Secure response headers
- Rate limiting
- Context value injection (`Provide`, `ProvideWith`)
- Adapter for stdlib middleware (`HTTP`)

## Ordering guidance

Place recovery and logging near the beginning of the chain so downstream
failures are consistently captured.
