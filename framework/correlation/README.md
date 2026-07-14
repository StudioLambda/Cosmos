# framework/correlation

Request correlation ID propagation for tracing and logging.

## What this package provides

- Middleware that ensures every request has a correlation ID.
- Log handler adapter that injects correlation ID into structured logs.
- Utilities to retrieve correlation IDs from request context.

## Typical usage

Install middleware early, then wrap your logger handler so all downstream logs
include the same request correlation value.
