package problem

import (
	"errors"
	"fmt"
	"testing"
)

func TestStackTraceNilError(t *testing.T) {
	t.Parallel()

	result := stackTrace(nil)

	if len(result) != 0 {
		t.Fatalf("expected empty slice, got %d elements", len(result))
	}
}

func TestStackTraceSimpleError(t *testing.T) {
	t.Parallel()

	err := errors.New("simple error")
	result := stackTrace(err)

	if len(result) != 1 {
		t.Fatalf("expected 1 error, got %d", len(result))
	}

	if result[0].Error() != "simple error" {
		t.Fatalf("expected %q, got %q", "simple error", result[0].Error())
	}
}

func TestStackTraceWrappedError(t *testing.T) {
	t.Parallel()

	inner := errors.New("inner")
	outer := fmt.Errorf("outer: %w", inner)
	result := stackTrace(outer)

	if len(result) != 1 {
		t.Fatalf("expected 1 error (wrapped single), got %d", len(result))
	}

	if result[0] != outer {
		t.Fatalf("expected outer error, got %v", result[0])
	}
}

func TestStackTraceJoinedErrors(t *testing.T) {
	t.Parallel()

	err1 := errors.New("first")
	err2 := errors.New("second")
	err3 := errors.New("third")
	joined := errors.Join(err1, err2, err3)

	result := stackTrace(joined)

	if len(result) != 3 {
		t.Fatalf("expected 3 errors, got %d", len(result))
	}

	if result[0].Error() != "first" {
		t.Fatalf("expected %q, got %q", "first", result[0].Error())
	}

	if result[1].Error() != "second" {
		t.Fatalf("expected %q, got %q", "second", result[1].Error())
	}

	if result[2].Error() != "third" {
		t.Fatalf("expected %q, got %q", "third", result[2].Error())
	}
}

func TestStackTraceNestedJoinedErrors(t *testing.T) {
	t.Parallel()

	err1 := errors.New("a")
	err2 := errors.New("b")
	err3 := errors.New("c")
	inner := errors.Join(err1, err2)
	outer := errors.Join(inner, err3)

	result := stackTrace(outer)

	if len(result) != 3 {
		t.Fatalf("expected 3 errors, got %d", len(result))
	}

	if result[0].Error() != "a" {
		t.Fatalf("expected %q, got %q", "a", result[0].Error())
	}

	if result[1].Error() != "b" {
		t.Fatalf("expected %q, got %q", "b", result[1].Error())
	}

	if result[2].Error() != "c" {
		t.Fatalf("expected %q, got %q", "c", result[2].Error())
	}
}

func TestStackTraceJoinedWithWrapped(t *testing.T) {
	t.Parallel()

	inner := errors.New("root")
	wrapped := fmt.Errorf("wrapped: %w", inner)
	other := errors.New("other")
	joined := errors.Join(wrapped, other)

	result := stackTrace(joined)

	if len(result) != 2 {
		t.Fatalf("expected 2 errors, got %d", len(result))
	}

	if result[0].Error() != "wrapped: root" {
		t.Fatalf("expected %q, got %q", "wrapped: root", result[0].Error())
	}

	if result[1].Error() != "other" {
		t.Fatalf("expected %q, got %q", "other", result[1].Error())
	}
}
