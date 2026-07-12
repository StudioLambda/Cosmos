# contract/response

Helpers for writing HTTP responses with consistent headers and encoding.

## What this package covers

- JSON, XML, text, and HTML responses.
- Streamed and server-sent event responses.
- File/static responses.
- Redirect helpers including safe relative-path validation.

## When to use it

Use this package from handlers to keep response writing explicit and concise,
while staying compatible with `net/http`.

## Security notes

Use `SafeRedirect` for user-influenced targets to avoid open redirect issues.
