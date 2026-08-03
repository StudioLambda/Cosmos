# Problem (RFC 9457)

Cosmos `problem` is a pure Go RFC 9457 implementation.

```bash
go get github.com/studiolambda/cosmos/problem
```

`problem.Details` implements:

- `error`
- `http.Handler`

---

## Define reusable templates

```go
var ErrNotFound = problem.Details{
	Type:   "https://api.example.com/errors/not-found",
	Title:  "Resource Not Found",
	Status: http.StatusNotFound,
}

var ErrForbidden = problem.Details{
	Type:   "https://api.example.com/errors/forbidden",
	Title:  "Forbidden",
	Status: http.StatusForbidden,
}
```

Use copy-on-write modifiers per request:

```go
return ErrNotFound.WithError(err).With("user_id", id)
```

This is the recommended consumer pattern: define `problem.Details` as top-level
error templates and derive per-request values with `WithError` and `With`.

---

## Core API

```go
p := ErrInternal.WithError(err)

p = p.With("trace_id", traceID)
p = p.Without("trace_id")

p = p.WithoutError()
```

All modifier methods return new values (immutable/copy-on-write style).

---

## Serving behavior

```go
p.ServeHTTP(w, r)
```

### Negotiation order

`ServeHTTP` supports:

- `application/problem+json`
- `application/json`
- fallback text/plain

If `Accept` includes `*/*`, JSON is preferred.

---

## Defaulting rules

`Defaulted(r)` fills missing values with:

- `Type`: `about:blank`
- `Status`: `500`
- `Title`: `http.StatusText(status)`
- `Instance`: `r.URL.Path`

Important: `Detail` is **not** auto-populated from `err.Error()`.

---

## JSON representation

JSON is a flat object with RFC fields + additional fields.
Wrapped `err` is not serialized directly.

## Error integration

```go
p := ErrNotFound.WithError(fmt.Errorf("query failed: %w", sql.ErrNoRows))

if errors.Is(p, sql.ErrNoRows) {
	// true
}

root := errors.Unwrap(p)
_ = root
```

---

## Framework interaction

When returned from a `framework.Handler`, `Details` is treated as `http.Handler`
and renders itself through the framework error pipeline.

```go
func getUser(w http.ResponseWriter, r *http.Request) error {
	return ErrNotFound.With("resource", "user")
}
```

---

## Gotchas

- Always capture returned value from `With*` / `Without*` methods.
- Wrapped errors are never serialized.
