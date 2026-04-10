---
name: cosmos-go
description: >
  Project-specific Go skill for the Cosmos monorepo. Must use when writing,
  editing, planning, or reviewing any Go (.go) files in this workspace.
  Covers code style, naming, documentation, error handling, immutability
  patterns, testing, and module architecture. Use this skill instead of the
  global Go skill whenever working inside the Cosmos codebase.
---

# Cosmos Go

Go conventions and patterns for the Cosmos HTTP framework monorepo.
Go 1.25+ workspace with four modules: contract, router, problem, framework.

## Project Structure

```
cosmos/
├── go.work              # Workspace: use + replace directives
├── contract/            # Interfaces (zero deps) - the foundation
├── router/              # Generic HTTP router (zero deps)
├── problem/             # RFC 9457 problem details (zero deps)
│   └── internal/        # Accept header parsing (not exported)
└── framework/           # Full framework (depends on all above)
    ├── cache/           # contract.Cache implementations
    ├── crypto/          # contract.Encrypter implementations
    ├── database/        # contract.Database implementation
    ├── event/           # contract.EventBus implementations
    ├── hash/            # contract.Hasher implementations
    ├── middleware/       # Ready-made middleware
    └── session/         # Session management
```

**Dependency direction:** contract (zero deps) -> router, problem (standalone) -> framework (uses all).

Always run tests from workspace root: `go test ./...`
Test a single module: `go test ./router/...`

## Formatting

Standard `gofmt`. No exceptions.

### Import Groups

Separate groups with a blank line, alphabetical within each group:

```go
import (
    "encoding/json"
    "fmt"
    "net/http"

    "github.com/studiolambda/cosmos/problem/internal"

    "github.com/stretchr/testify/assert"
)
```

- Group 1: Standard library
- Group 2: Internal cosmos modules
- Group 3: External dependencies

When only stdlib imports exist, use a single group.

## Naming

### Receiver Names

Use **full-word receiver names** matching the type name in lowercase. This is a firm convention.

```go
// Correct
func (problem Problem) With(key string, value any) Problem { ... }
func (router *Router[H]) Group(pattern string, fn func(*Router[H])) { ... }
func (accept Accept) Quality(media string) float64 { ... }

// Wrong
func (p Problem) With(key string, value any) Problem { ... }
func (r *Router[H]) Group(pattern string, fn func(*Router[H])) { ... }
```

Rationale: Short names create ambiguity in methods >10 lines and make grep/search
less useful. `r` could be a Router, Request, Reader, or Redis client. Full words
eliminate that class of confusion entirely.

### General Naming

- Descriptive names over abbreviations: `subrouter` not `sr`, `handler` not `h` (exception: named return values like `h H, ok bool` are fine).
- PascalCase for exported, camelCase for unexported.
- Constructor: `New` for single-type packages, `NewTypeName` when multiple types exist.
- Symmetric method pairs: `With`/`Without`, `WithError`/`WithoutError`.
- Past participle for methods returning a modified copy: `Defaulted`, `Grouped`.

## Functions

- **Small and focused.** Target <30 lines body, <5 cyclomatic complexity.
- **Single purpose.** If you need a comment saying "now do the second thing," extract a function.
- **Early returns over nesting.** Use guard clauses for preconditions, errors, and edge cases.
- **Zero `else` after early returns.** The happy path flows downward without indentation.

```go
// Correct: early return, no else
func (problem Problem) With(key string, value any) Problem {
    if problem.additional == nil {
        problem.additional = map[string]any{key: value}

        return problem
    }

    problem.additional = maps.Clone(problem.additional)
    problem.additional[key] = value

    return problem
}

// Wrong: unnecessary else
func (problem Problem) With(key string, value any) Problem {
    if problem.additional == nil {
        problem.additional = map[string]any{key: value}
    } else {
        problem.additional = maps.Clone(problem.additional)
        problem.additional[key] = value
    }

    return problem
}
```

## Blank Lines

Use blank lines to separate logical blocks within functions:

- After variable declarations before logic.
- Between guard clauses / early-return blocks.
- Between distinct logical steps.
- After `struct {` opening when fields have doc-comment blocks.

Do not use blank lines between tightly coupled one-liners (e.g. sequential `delete()` calls).

## Documentation

### Doc Comments

All exported symbols get doc comments. Start with the symbol name:

```go
// ParseAccept creates a new [Accept] based on a
// [http.Request] by parsing its headers.
func ParseAccept(request *http.Request) Accept {
```

Unexported functions with non-obvious logic also get doc comments:

```go
// stackTrace creates a stack trace of all the errors found
// that have been either Joined or Wrapped using [errors.Join]
// or [fmt.Errorf] with `%w` directive.
func stackTrace(err error) []error {
```

### Godoc Links

Use bracket syntax to cross-reference types and functions:

```go
// With adds a new additional value to the given key.
// Use [Problem.Without] to remove values.
// See [NewProblem] for creating problems from errors.
```

### Struct Field Comments

For structs with >3 fields, document each field with a comment block above it.
For structs with 1-3 fields, inline comments or a single doc comment on the type suffice.

### Sentence Completeness

Doc comments must be complete sentences. End with a period. Do not leave
truncated sentences.

## Error Handling

- Early-return on error: `if err != nil { return err }`.
- Use `errors.Is()` and `errors.As()` for error inspection, never type assertions.
- When intentionally discarding an error, use explicit `_ =` and comment the reason:

```go
// Error is intentionally discarded: headers are already written,
// nothing useful can be done if encoding fails.
_ = json.NewEncoder(w).Encode(problem)
```

- Never silently swallow errors without a documented reason.
- Error strings are lowercase, no punctuation at the end.

## Design Patterns

### Immutability (Copy-on-Write)

For types holding mutable internal state (maps, slices), prefer value receivers
and methods that return modified copies rather than mutating in place.

Clone mutable fields before modification to prevent aliasing:

```go
func (problem Problem) With(key string, value any) Problem {
    problem.additional = maps.Clone(problem.additional)
    problem.additional[key] = value

    return problem
}
```

This enables safe derivation from package-level variables: `ErrNotFound.With("id", 42)`.

Use `maps.Clone`, `slices.Clone` from the standard library. Never hand-roll copy loops.

When mutation is the natural API (like `Router.Use()`), document the contrast
with the immutable alternative clearly.

### Encapsulation

Unexported struct fields, exported methods. Implementation details go in
`internal/` packages when they serve a single parent package.

### Interface Satisfaction

Satisfy interfaces implicitly (duck typing). Do not add explicit compile-time checks
like `var _ Interface = Type{}` unless the satisfaction is non-obvious or critical.

Common interfaces to implement: `error`, `http.Handler`, `json.Marshaler`/`json.Unmarshaler`.

### Generics

Use type constraints meaningfully. Prefer named constraints over inline unions:

```go
type Middleware[H http.Handler] = func(H) H    // type alias preserves assignability
type Router[H http.Handler] struct { ... }      // constrained to http.Handler
```

### Modern Standard Library

Prefer modern Go stdlib functions:

- `strings.CutPrefix`, `strings.CutSuffix` over manual `HasPrefix` + `TrimPrefix` combinations.
- `strings.SplitSeq` (iterator-based) over `strings.Split` when not materializing the full slice.
- `slices.SortFunc`, `slices.Clone` over hand-rolled sort/copy.
- `maps.Clone`, `maps.Copy` over manual map iteration.
- `path.Join` for URL path construction.

## Testing

See [references/testing-guide.md](references/testing-guide.md) for complete patterns and examples.

Summary:

- **Atomic test functions.** One concern per test function. No table-driven tests.
- **Parallel by default.** Call `t.Parallel()` in every test unless shared state prevents it.
- **Black-box testing.** Use `package foo_test` (external test package).
- **Testify assertions.** Use `require` for fatal checks, `assert` for non-fatal.
- **Descriptive names.** `TestProblemWithAddsAdditionalValue`, `TestRouterHasMatchesTrailingSlash`.
- **httptest for HTTP.** `httptest.NewRequest` + `httptest.NewRecorder` or `Router.Record()`.

## When to Load References

- **Writing or reviewing Go source code:** See [references/code-patterns.md](references/code-patterns.md) for annotated examples of all major patterns (immutability, middleware, constructors, content negotiation, recursive resolution).
- **Writing or reviewing tests:** See [references/testing-guide.md](references/testing-guide.md) for test structure, assertion patterns, HTTP testing, and anti-patterns.
