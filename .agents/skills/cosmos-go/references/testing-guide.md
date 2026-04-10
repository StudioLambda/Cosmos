# Testing Guide

Testing conventions for the Cosmos codebase. Load this reference when
writing or reviewing test files.

## Test Structure

### Atomic Test Functions

Each test function covers one scenario. No table-driven tests.
No subtests via `t.Run` for simple cases.

```go
func TestProblemWithAddsValue(t *testing.T) {
	t.Parallel()

	problem := Problem{Status: http.StatusNotFound}
	result := problem.With("user_id", 42)

	value, ok := result.Additional("user_id")

	require.True(t, ok)
	require.Equal(t, 42, value)
}

func TestProblemWithPreservesOriginal(t *testing.T) {
	t.Parallel()

	original := Problem{Status: http.StatusNotFound}
	_ = original.With("key", "value")

	_, ok := original.Additional("key")

	require.False(t, ok)
}
```

Key rules:
- One concern per test function.
- `t.Parallel()` as the first line unless shared state prevents it.
- Descriptive name: `TestTypeName` + verb/behaviour description.
- CamelCase names, no underscores.

### Black-Box Testing

Use external test packages (`package foo_test`) by default. This tests
the public API surface and catches accidental coupling to internals.

```go
package session_test

import (
	"testing"

	"github.com/studiolambda/cosmos/framework/session"

	"github.com/stretchr/testify/require"
)
```

Only use `package foo` (white-box) when testing unexported helpers that
have complex logic worth verifying independently.

## Assertions

Use testify's `require` package for all assertions. Never use `assert`
(which continues execution on failure) unless you explicitly need
multiple independent checks in one test.

```go
// Fatal on failure -- stops the test immediately.
require.NoError(t, err)
require.Equal(t, expected, actual)
require.True(t, condition)
require.Nil(t, value)
require.Len(t, slice, 3)

// Argument order: (t, expected, actual) -- never reversed.
require.Equal(t, http.StatusOK, response.StatusCode)
```

Never discard errors in tests. Always assert them:

```go
// Wrong: silently discards constructor error.
e, _ := crypto.NewAES(key)

// Correct: fails the test if constructor returns an error.
e, err := crypto.NewAES(key)
require.NoError(t, err)
```

## HTTP Handler Testing

Use `httptest.NewRequest` for request creation and
`httptest.NewRecorder` for response capture.

```go
func TestHandlerReturnsOK(t *testing.T) {
	t.Parallel()

	handler := framework.Handler(func(w http.ResponseWriter, r *http.Request) error {
		return response.JSON(w, http.StatusOK, map[string]string{"ok": "true"})
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
}
```

For router-level testing, use the `Record` helper:

```go
func TestRouterGetRoot(t *testing.T) {
	t.Parallel()

	r := router.New[http.HandlerFunc]()
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	res := r.Record(req)

	require.Equal(t, http.StatusOK, res.StatusCode)
}
```

## Test Naming

Descriptive CamelCase names that communicate the scenario:

```
TestSessionRegenerateChangesID
TestSessionHasExpiredAfterTTL
TestRouterHasMatchesTrailingSlash
TestCSRFBlocksCrossOriginPost
TestAESEncryptDecryptRoundTrip
```

Avoid generic names like `TestItWorks`, `TestBasic`, or `TestHappyPath`.

## Anti-Patterns

- **No table-driven tests.** Each scenario is its own function.
- **No `t.Run` subtests** for simple cases. Use separate test functions.
- **No shared mutable state** between tests. Duplicate setup.
- **No `_ =` for errors** in tests. Always `require.NoError`.
- **No raw `if/t.Fatalf`** when testify is available.
- **No `defer` for cleanup.** Use `t.Cleanup()` instead.
- **Never reverse `require.Equal` args.** Always `(t, expected, actual)`.
