package problem_test

import (
	"encoding/json/v2"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/studiolambda/cosmos/problem"
)

func TestWithContextValuesComposesValues(t *testing.T) {
	t.Parallel()

	ctx := problem.WithContextValues(t.Context(), map[string]any{"request_id": "first"})
	ctx = problem.WithContextValues(ctx, map[string]any{
		"request_id":     "second",
		"correlation_id": "correlation",
	})
	req := httptest.NewRequest(http.MethodGet, "/", nil).WithContext(ctx)
	req.Header.Set("Accept", "application/json")
	rec := httptest.NewRecorder()

	problem.Details{Status: http.StatusBadRequest}.ServeHTTP(rec, req)

	var body map[string]any
	if err := json.UnmarshalRead(rec.Body, &body); err != nil {
		t.Fatalf("failed to decode response: %s", err)
	}

	if body["request_id"] != "second" {
		t.Fatalf("expected latest request_id, got %v", body["request_id"])
	}

	if body["correlation_id"] != "correlation" {
		t.Fatalf("expected correlation_id, got %v", body["correlation_id"])
	}
}

func TestWithContextValuesCopiesInput(t *testing.T) {
	t.Parallel()

	values := map[string]any{"request_id": "original"}
	ctx := problem.WithContextValues(t.Context(), values)
	values["request_id"] = "modified"
	req := httptest.NewRequest(http.MethodGet, "/", nil).WithContext(ctx)
	req.Header.Set("Accept", "application/json")
	rec := httptest.NewRecorder()

	problem.Details{Status: http.StatusBadRequest}.ServeHTTP(rec, req)

	var body map[string]any
	if err := json.UnmarshalRead(rec.Body, &body); err != nil {
		t.Fatalf("failed to decode response: %s", err)
	}

	if body["request_id"] != "original" {
		t.Fatalf("expected original request_id, got %v", body["request_id"])
	}
}

func TestServeHTTPExplicitAdditionalOverridesContextValue(t *testing.T) {
	t.Parallel()

	ctx := problem.WithContextValues(t.Context(), map[string]any{"request_id": "context"})
	req := httptest.NewRequest(http.MethodGet, "/", nil).WithContext(ctx)
	req.Header.Set("Accept", "application/json")
	rec := httptest.NewRecorder()

	problem.Details{Status: http.StatusBadRequest}.
		With("request_id", "explicit").
		ServeHTTP(rec, req)

	var body map[string]any
	if err := json.UnmarshalRead(rec.Body, &body); err != nil {
		t.Fatalf("failed to decode response: %s", err)
	}

	if body["request_id"] != "explicit" {
		t.Fatalf("expected explicit request_id, got %v", body["request_id"])
	}
}

func TestServeHTTPIgnoresStandardContextMembers(t *testing.T) {
	t.Parallel()

	ctx := problem.WithContextValues(t.Context(), map[string]any{
		"status":     http.StatusTeapot,
		"title":      "Context Title",
		"request_id": "request",
	})
	req := httptest.NewRequest(http.MethodGet, "/", nil).WithContext(ctx)
	req.Header.Set("Accept", "application/json")
	rec := httptest.NewRecorder()

	problem.Details{
		Status: http.StatusBadRequest,
		Title:  "Bad Request",
	}.ServeHTTP(rec, req)

	var body map[string]any
	if err := json.UnmarshalRead(rec.Body, &body); err != nil {
		t.Fatalf("failed to decode response: %s", err)
	}

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}

	if body["title"] != "Bad Request" {
		t.Fatalf("expected explicit title, got %v", body["title"])
	}

	if body["request_id"] != "request" {
		t.Fatalf("expected request_id, got %v", body["request_id"])
	}
}
