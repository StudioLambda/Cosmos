# Code Patterns

Annotated examples from the Cosmos codebase. Load this reference when writing
or reviewing Go source code in the workspace.

## Immutable Copy-on-Write (Problem Package)

Types with mutable internal state (maps, slices) use value receivers and
return modified copies. Clone mutable fields before mutation to prevent
aliasing with the original.

```go
// With adds a new additional value to the given key.
// It returns a new Problem, leaving the original untouched.
// This enables safe derivation from package-level variables:
//
//	ErrNotFound.With("user_id", 42)
func (problem Problem) With(key string, value any) Problem {
	if problem.additional == nil {
		problem.additional = map[string]any{key: value}

		return problem
	}

	problem.additional = maps.Clone(problem.additional)
	problem.additional[key] = value

	return problem
}
```

Key elements:
- Value receiver (`Problem`, not `*Problem`).
- `maps.Clone` before writing to the map.
- Guard clause for nil map with early return.
- Symmetric pair: `With` / `Without`, `WithError` / `WithoutError`.

The only exception for pointer receivers on immutable types is when required
by an interface (`json.Unmarshaler` requires `*Problem`).

## Constructor Pattern

Single-type packages use `New`. Multi-type packages use `NewTypeName`.
Explicitly initialize all fields, even zero values, to signal intent.

```go
// New creates a fresh router with an empty middleware stack
// and a newly allocated http.ServeMux for route registration.
func New[H http.Handler]() *Router[H] {
	return &Router[H]{
		native:      http.NewServeMux(),
		pattern:     "",
		parent:      nil,
		middlewares: make([]Middleware[H], 0),
	}
}
```

Key elements:
- `make([]T, 0)` instead of `nil` for slices that will be appended to.
- All fields listed, even when zero-valued (`pattern: ""`, `parent: nil`).
- Generic type parameter propagated to the return type.

## Middleware / Decorator Pattern

Middleware is a type alias (not a defined type) to preserve assignability.
Application order is reversed so that the first registered middleware
executes first on the request path.

```go
// Middleware is a function that wraps a handler with additional
// behaviour. The type alias preserves direct assignability from
// plain function literals without explicit conversion.
type Middleware[H http.Handler] = func(H) H

// wrap applies all registered middlewares to the handler in
// reverse order. This ensures the first middleware added via
// Use() is the outermost layer and executes first on the
// request path.
func (router *Router[H]) wrap(handler H) H {
	for i := len(router.middlewares) - 1; i >= 0; i-- {
		handler = router.middlewares[i](handler)
	}

	return handler
}
```

## Group / Callback Pattern

Scoped configuration via callback functions. The sub-router inherits
the parent's middleware (via `slices.Clone`) and pattern prefix (via
`path.Join`), but mutations to the sub-router do not affect the parent.

```go
// Group creates a sub-router with the given pattern prefix.
// Routes registered inside the callback inherit the prefix
// and any middleware already applied to the parent router.
func (router *Router[H]) Group(pattern string, subrouter func(*Router[H])) {
	subrouter(&Router[H]{
		native:      nil, // resolved via parent's mux()
		pattern:     path.Join(router.pattern, pattern),
		parent:      router,
		middlewares: slices.Clone(router.middlewares),
	})
}
```

## Content Negotiation Dispatch

Map of content types to handlers, iterated in Accept header quality order.
Falls back to a default when no match is found.

```go
func (problem Problem) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	problem = problem.Defaulted(r)
	responses := map[string]http.Handler{
		"application/problem+json": http.HandlerFunc(problem.jsonProblemHandler),
		"application/json":         http.HandlerFunc(problem.jsonHandler),
	}

	for _, media := range internal.ParseAccept(r).Order() {
		if response, ok := responses[media]; ok {
			response.ServeHTTP(w, r)
			return
		}
	}

	// Fall back to plain text when no JSON type is accepted.
	http.HandlerFunc(problem.textHandler).ServeHTTP(w, r)
}
```

## Recursive Resolution

Sub-routers form a tree via `parent` pointers. Shared state (the `http.ServeMux`)
is resolved by walking up to the root. Only the root router holds the actual mux.

```go
// mux returns the native http.ServeMux used internally.
// Sub-routers delegate to their parent recursively, ensuring
// all routes in the tree end up registered on the same mux.
func (router *Router[H]) mux() *http.ServeMux {
	if router.parent != nil {
		return router.parent.mux()
	}

	return router.native
}
```

## Interface Satisfaction

Satisfy interfaces implicitly. The compiler enforces correctness at usage
sites. Common interfaces implemented in the codebase:

- `error` via `Error() string` and `Unwrap() error`.
- `http.Handler` via `ServeHTTP(http.ResponseWriter, *http.Request)`.
- `json.Marshaler` / `json.Unmarshaler` for custom JSON encoding.
- `HTTPStatus` via `HTTPStatus() int` for error-to-status-code mapping.

## Import Organization

Three groups separated by blank lines, alphabetical within each:

```go
import (
	// Standard library
	"encoding/json"
	"fmt"
	"net/http"

	// Internal cosmos modules
	"github.com/studiolambda/cosmos/contract"
	"github.com/studiolambda/cosmos/problem"

	// External third-party
	"github.com/stretchr/testify/require"
)
```

When only stdlib imports exist, use a single group with no blank lines.

## Modern Standard Library

Prefer modern Go stdlib functions over manual equivalents:

```go
// Prefer CutSuffix over HasSuffix + TrimSuffix
if pattern, ok := strings.CutSuffix(pattern, "/"); ok {
	router.register(method, pattern, handler)
}

// Prefer SplitSeq (iterator) over Split (materializes full slice)
for line := range strings.SplitSeq(header, ",") {
	// process each line lazily
}

// Prefer slices/maps package over hand-rolled loops
problem.additional = maps.Clone(problem.additional)
values := slices.Clone(accept.values)
slices.SortFunc(values, func(a, b acceptPair) int { ... })
```

## Early Returns and Guard Clauses

Every branching function follows the same discipline: guard clauses at
the top, happy path flowing downward without indentation, zero `else`
blocks after early returns.

```go
// Defaulted fills in zero-value fields with sensible defaults.
// Each field is guarded independently, keeping the logic flat.
func (problem Problem) Defaulted(request *http.Request) Problem {
	if problem.Type == "" {
		problem.Type = "about:blank"
	}

	if problem.Status == 0 {
		problem.Status = http.StatusInternalServerError
	}

	if problem.Title == "" {
		problem.Title = http.StatusText(problem.Status)
	}

	if problem.Instance == "" {
		problem.Instance = request.URL.String()
	}

	if traces := problem.Errors(); problem.Detail == "" && len(traces) > 0 {
		problem.Detail = traces[0].Error()
	}

	return problem
}
```

## Encapsulation

Unexported struct fields, exported methods. Implementation details that
serve a single parent package go in `internal/` sub-packages.

```go
type Router[H http.Handler] struct {
	native      *http.ServeMux   // unexported: internal wiring
	pattern     string           // unexported: managed via Group/Method
	parent      *Router[H]       // unexported: tree structure
	middlewares []Middleware[H]   // unexported: managed via Use/With
}
```

External consumers interact exclusively through exported methods.
The `internal/` package restriction prevents other modules from
importing implementation details like Accept header parsing.
