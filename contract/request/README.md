# contract/request

Helpers for reading HTTP request data safely and consistently.

## What this package covers

- Header and cookie accessors.
- Route parameter and query helpers, including typed integer parsing.
- Body readers for bytes, strings, JSON, and XML.
- Size-limited and strict JSON/XML decoding helpers.
- Session and hooks retrieval from request context.
- Correlation ID retrieval.

## When to use it

Use this package in handlers and middleware when you want concise request
parsing without re-implementing common boilerplate.

## Security notes

Prefer size-limited helpers (`LimitedJSON`, `LimitedXML`, `LimitedBytes`) over
unbounded variants for untrusted input.

Prefer strict decoders when rejecting unknown fields is required.
