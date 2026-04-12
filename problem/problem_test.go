package problem_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/studiolambda/cosmos/problem"
)

func TestNewProblem(t *testing.T) {
	t.Parallel()

	err := errors.New("something failed")
	p := problem.NewProblem(err, http.StatusBadRequest)

	if p.Status != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, p.Status)
	}

	if p.Unwrap() != err {
		t.Fatalf("expected error %v, got %v", err, p.Unwrap())
	}
}

func TestNewProblemNilError(t *testing.T) {
	t.Parallel()

	p := problem.NewProblem(nil, http.StatusNotFound)

	if p.Status != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, p.Status)
	}

	if p.Unwrap() != nil {
		t.Fatalf("expected nil error, got %v", p.Unwrap())
	}
}

func TestAdditionalFound(t *testing.T) {
	t.Parallel()

	p := problem.Problem{Status: http.StatusBadRequest}
	p = p.With("key", "value")

	val, ok := p.Additional("key")

	if !ok {
		t.Fatalf("expected key to be found")
	}

	if val != "value" {
		t.Fatalf("expected value %q, got %q", "value", val)
	}
}

func TestAdditionalNotFound(t *testing.T) {
	t.Parallel()

	p := problem.Problem{Status: http.StatusBadRequest}

	val, ok := p.Additional("missing")

	if ok {
		t.Fatalf("expected key to not be found")
	}

	if val != nil {
		t.Fatalf("expected nil value, got %v", val)
	}
}

func TestAdditionalNotFoundNilMap(t *testing.T) {
	t.Parallel()

	p := problem.Problem{}

	val, ok := p.Additional("anything")

	if ok {
		t.Fatalf("expected key to not be found on nil map")
	}

	if val != nil {
		t.Fatalf("expected nil value, got %v", val)
	}
}

func TestWithNilMap(t *testing.T) {
	t.Parallel()

	p := problem.Problem{Status: http.StatusBadRequest}
	p = p.With("foo", "bar")

	val, ok := p.Additional("foo")

	if !ok {
		t.Fatalf("expected key to be found")
	}

	if val != "bar" {
		t.Fatalf("expected value %q, got %q", "bar", val)
	}
}

func TestWithExistingMap(t *testing.T) {
	t.Parallel()

	p := problem.Problem{Status: http.StatusBadRequest}
	p = p.With("first", 1)
	p = p.With("second", 2)

	val1, ok1 := p.Additional("first")
	val2, ok2 := p.Additional("second")

	if !ok1 || !ok2 {
		t.Fatalf("expected both keys to be found")
	}

	if val1 != 1 {
		t.Fatalf("expected first=1, got %v", val1)
	}

	if val2 != 2 {
		t.Fatalf("expected second=2, got %v", val2)
	}
}

func TestWithDoesNotMutateOriginal(t *testing.T) {
	t.Parallel()

	original := problem.Problem{Status: http.StatusBadRequest}
	original = original.With("key", "original")
	modified := original.With("key", "modified")

	val, _ := original.Additional("key")

	if val != "original" {
		t.Fatalf("original was mutated: expected %q, got %v", "original", val)
	}

	val2, _ := modified.Additional("key")

	if val2 != "modified" {
		t.Fatalf("modified has wrong value: expected %q, got %v", "modified", val2)
	}
}

func TestWithError(t *testing.T) {
	t.Parallel()

	p := problem.Problem{Status: http.StatusInternalServerError}
	err := errors.New("database error")
	p = p.WithError(err)

	if p.Unwrap() != err {
		t.Fatalf("expected error %v, got %v", err, p.Unwrap())
	}
}

func TestWithoutError(t *testing.T) {
	t.Parallel()

	err := errors.New("some error")
	p := problem.NewProblem(err, http.StatusInternalServerError)
	p = p.WithoutError()

	if p.Unwrap() != nil {
		t.Fatalf("expected nil error, got %v", p.Unwrap())
	}
}

func TestWithoutNilMap(t *testing.T) {
	t.Parallel()

	p := problem.Problem{Status: http.StatusBadRequest}
	p = p.Without("nonexistent")

	if p.Status != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, p.Status)
	}
}

func TestWithoutExistingKey(t *testing.T) {
	t.Parallel()

	p := problem.Problem{Status: http.StatusBadRequest}
	p = p.With("key", "value")
	p = p.Without("key")

	_, ok := p.Additional("key")

	if ok {
		t.Fatalf("expected key to be removed")
	}
}

func TestWithoutDoesNotMutateOriginal(t *testing.T) {
	t.Parallel()

	original := problem.Problem{Status: http.StatusBadRequest}
	original = original.With("key", "value")
	_ = original.Without("key")

	val, ok := original.Additional("key")

	if !ok {
		t.Fatalf("original was mutated: key was removed")
	}

	if val != "value" {
		t.Fatalf("original was mutated: expected %q, got %v", "value", val)
	}
}

func TestErrorWithErr(t *testing.T) {
	t.Parallel()

	err := errors.New("something broke")
	p := problem.NewProblem(err, http.StatusInternalServerError)

	got := p.Error()
	expected := "500 internal server error: something broke"

	if got != expected {
		t.Fatalf("expected %q, got %q", expected, got)
	}
}

func TestErrorWithoutErr(t *testing.T) {
	t.Parallel()

	p := problem.Problem{
		Status: http.StatusNotFound,
		Title:  "Resource Not Found",
	}

	got := p.Error()
	expected := "404 not found: resource not found"

	if got != expected {
		t.Fatalf("expected %q, got %q", expected, got)
	}
}

func TestErrorsNilError(t *testing.T) {
	t.Parallel()

	p := problem.Problem{}
	errs := p.Errors()

	if len(errs) != 0 {
		t.Fatalf("expected empty errors, got %d", len(errs))
	}
}

func TestErrorsSingleError(t *testing.T) {
	t.Parallel()

	err := errors.New("one error")
	p := problem.NewProblem(err, http.StatusBadRequest)
	errs := p.Errors()

	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d", len(errs))
	}

	if errs[0].Error() != "one error" {
		t.Fatalf("expected %q, got %q", "one error", errs[0].Error())
	}
}

func TestErrorsWrappedChain(t *testing.T) {
	t.Parallel()

	inner := errors.New("inner")
	outer := fmt.Errorf("outer: %w", inner)
	p := problem.NewProblem(outer, http.StatusInternalServerError)
	errs := p.Errors()

	if len(errs) != 1 {
		t.Fatalf("expected 1 error (fmt.Errorf wraps single), got %d", len(errs))
	}
}

func TestErrorsJoinedErrors(t *testing.T) {
	t.Parallel()

	err1 := errors.New("first")
	err2 := errors.New("second")
	joined := errors.Join(err1, err2)
	p := problem.NewProblem(joined, http.StatusInternalServerError)
	errs := p.Errors()

	if len(errs) != 2 {
		t.Fatalf("expected 2 errors, got %d", len(errs))
	}
}

func TestUnwrap(t *testing.T) {
	t.Parallel()

	err := errors.New("test error")
	p := problem.NewProblem(err, http.StatusBadRequest)

	if p.Unwrap() != err {
		t.Fatalf("expected %v, got %v", err, p.Unwrap())
	}
}

func TestUnwrapNil(t *testing.T) {
	t.Parallel()

	p := problem.Problem{}

	if p.Unwrap() != nil {
		t.Fatalf("expected nil, got %v", p.Unwrap())
	}
}

func TestWithStackTrace(t *testing.T) {
	t.Parallel()

	err := errors.New("root cause")
	p := problem.NewProblem(err, http.StatusInternalServerError)
	p = p.WithStackTrace()

	val, ok := p.Additional(problem.StackTraceKey)

	if !ok {
		t.Fatalf("expected stack_trace to be present")
	}

	traces, ok := val.([]string)

	if !ok {
		t.Fatalf("expected []string, got %T", val)
	}

	if len(traces) != 1 {
		t.Fatalf("expected 1 trace, got %d", len(traces))
	}

	if traces[0] != "root cause" {
		t.Fatalf("expected %q, got %q", "root cause", traces[0])
	}
}

func TestWithStackTraceNoError(t *testing.T) {
	t.Parallel()

	p := problem.Problem{Status: http.StatusBadRequest}
	p = p.WithStackTrace()

	val, ok := p.Additional(problem.StackTraceKey)

	if !ok {
		t.Fatalf("expected stack_trace key to be present")
	}

	traces, ok := val.([]string)

	if !ok {
		t.Fatalf("expected []string, got %T", val)
	}

	if len(traces) != 0 {
		t.Fatalf("expected 0 traces, got %d", len(traces))
	}
}

func TestWithoutStackTrace(t *testing.T) {
	t.Parallel()

	err := errors.New("root cause")
	p := problem.NewProblem(err, http.StatusInternalServerError)
	p = p.WithStackTrace()
	p = p.WithoutStackTrace()

	_, ok := p.Additional(problem.StackTraceKey)

	if ok {
		t.Fatalf("expected stack_trace to be removed")
	}
}

func TestMarshalJSON(t *testing.T) {
	t.Parallel()

	p := problem.Problem{
		Type:     "https://example.com/errors/test",
		Title:    "Test Error",
		Detail:   "Something went wrong",
		Status:   http.StatusBadRequest,
		Instance: "/test/123",
	}

	data, err := json.Marshal(p)

	if err != nil {
		t.Fatalf("failed to marshal: %s", err)
	}

	var decoded map[string]any

	err = json.Unmarshal(data, &decoded)

	if err != nil {
		t.Fatalf("failed to unmarshal result: %s", err)
	}

	if decoded["type"] != "https://example.com/errors/test" {
		t.Fatalf("expected type %q, got %v", "https://example.com/errors/test", decoded["type"])
	}

	if decoded["title"] != "Test Error" {
		t.Fatalf("expected title %q, got %v", "Test Error", decoded["title"])
	}

	if decoded["detail"] != "Something went wrong" {
		t.Fatalf("expected detail %q, got %v", "Something went wrong", decoded["detail"])
	}

	if decoded["status"] != float64(400) {
		t.Fatalf("expected status %v, got %v", float64(400), decoded["status"])
	}

	if decoded["instance"] != "/test/123" {
		t.Fatalf("expected instance %q, got %v", "/test/123", decoded["instance"])
	}
}

func TestMarshalJSONWithAdditional(t *testing.T) {
	t.Parallel()

	p := problem.Problem{
		Type:   "https://example.com/errors/test",
		Title:  "Test",
		Status: http.StatusBadRequest,
	}

	p = p.With("custom_field", "custom_value")

	data, err := json.Marshal(p)

	if err != nil {
		t.Fatalf("failed to marshal: %s", err)
	}

	var decoded map[string]any

	err = json.Unmarshal(data, &decoded)

	if err != nil {
		t.Fatalf("failed to unmarshal result: %s", err)
	}

	if decoded["custom_field"] != "custom_value" {
		t.Fatalf(
			"expected custom_field %q, got %v",
			"custom_value",
			decoded["custom_field"],
		)
	}
}

func TestUnmarshalJSON(t *testing.T) {
	t.Parallel()

	raw := `{
		"type": "https://example.com/errors/test",
		"title": "Test Error",
		"detail": "Something went wrong",
		"status": 400,
		"instance": "/test/123"
	}`

	var p problem.Problem

	err := json.Unmarshal([]byte(raw), &p)

	if err != nil {
		t.Fatalf("failed to unmarshal: %s", err)
	}

	if p.Type != "https://example.com/errors/test" {
		t.Fatalf("expected type %q, got %q", "https://example.com/errors/test", p.Type)
	}

	if p.Title != "Test Error" {
		t.Fatalf("expected title %q, got %q", "Test Error", p.Title)
	}

	if p.Detail != "Something went wrong" {
		t.Fatalf("expected detail %q, got %q", "Something went wrong", p.Detail)
	}

	if p.Status != 400 {
		t.Fatalf("expected status %d, got %d", 400, p.Status)
	}

	if p.Instance != "/test/123" {
		t.Fatalf("expected instance %q, got %q", "/test/123", p.Instance)
	}
}

func TestUnmarshalJSONWithAdditional(t *testing.T) {
	t.Parallel()

	raw := `{
		"type": "https://example.com/errors/test",
		"title": "Test",
		"status": 400,
		"custom": "extra"
	}`

	var p problem.Problem

	err := json.Unmarshal([]byte(raw), &p)

	if err != nil {
		t.Fatalf("failed to unmarshal: %s", err)
	}

	val, ok := p.Additional("custom")

	if !ok {
		t.Fatalf("expected custom key to be found")
	}

	if val != "extra" {
		t.Fatalf("expected %q, got %v", "extra", val)
	}
}

func TestUnmarshalJSONMissingFields(t *testing.T) {
	t.Parallel()

	raw := `{}`

	var p problem.Problem

	err := json.Unmarshal([]byte(raw), &p)

	if err != nil {
		t.Fatalf("failed to unmarshal: %s", err)
	}

	if p.Type != "" {
		t.Fatalf("expected empty type, got %q", p.Type)
	}

	if p.Title != "" {
		t.Fatalf("expected empty title, got %q", p.Title)
	}

	if p.Detail != "" {
		t.Fatalf("expected empty detail, got %q", p.Detail)
	}

	if p.Status != 0 {
		t.Fatalf("expected zero status, got %d", p.Status)
	}

	if p.Instance != "" {
		t.Fatalf("expected empty instance, got %q", p.Instance)
	}
}

func TestUnmarshalJSONInvalid(t *testing.T) {
	t.Parallel()

	var p problem.Problem

	err := p.UnmarshalJSON([]byte(`not valid json`))

	if err == nil {
		t.Fatalf("expected error for invalid JSON")
	}
}

func TestDefaultedAllEmpty(t *testing.T) {
	t.Parallel()

	p := problem.Problem{}
	req := httptest.NewRequest(http.MethodGet, "/test/path", nil)
	p = p.Defaulted(req)

	if p.Type != "about:blank" {
		t.Fatalf("expected type %q, got %q", "about:blank", p.Type)
	}

	if p.Status != http.StatusInternalServerError {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusInternalServerError,
			p.Status,
		)
	}

	if p.Title != "Internal Server Error" {
		t.Fatalf("expected title %q, got %q", "Internal Server Error", p.Title)
	}

	if p.Instance != "/test/path" {
		t.Fatalf("expected instance %q, got %q", "/test/path", p.Instance)
	}
}

func TestDefaultedAllFilled(t *testing.T) {
	t.Parallel()

	p := problem.Problem{
		Type:     "https://example.com/errors/custom",
		Title:    "Custom Title",
		Status:   http.StatusBadRequest,
		Instance: "/custom/instance",
	}

	req := httptest.NewRequest(http.MethodGet, "/test/path", nil)
	p = p.Defaulted(req)

	if p.Type != "https://example.com/errors/custom" {
		t.Fatalf(
			"expected type %q, got %q",
			"https://example.com/errors/custom",
			p.Type,
		)
	}

	if p.Status != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, p.Status)
	}

	if p.Title != "Custom Title" {
		t.Fatalf("expected title %q, got %q", "Custom Title", p.Title)
	}

	if p.Instance != "/custom/instance" {
		t.Fatalf("expected instance %q, got %q", "/custom/instance", p.Instance)
	}
}

func TestDefaultedPartialFilled(t *testing.T) {
	t.Parallel()

	p := problem.Problem{
		Status: http.StatusNotFound,
	}

	req := httptest.NewRequest(http.MethodGet, "/items/42", nil)
	p = p.Defaulted(req)

	if p.Type != "about:blank" {
		t.Fatalf("expected type %q, got %q", "about:blank", p.Type)
	}

	if p.Status != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, p.Status)
	}

	if p.Title != "Not Found" {
		t.Fatalf("expected title %q, got %q", "Not Found", p.Title)
	}

	if p.Instance != "/items/42" {
		t.Fatalf("expected instance %q, got %q", "/items/42", p.Instance)
	}
}

func TestServeHTTPProblemJSON(t *testing.T) {
	t.Parallel()

	p := problem.Problem{
		Type:   "https://example.com/errors/test",
		Title:  "Test Error",
		Detail: "Details here",
		Status: http.StatusBadRequest,
	}

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Accept", "application/problem+json")
	rec := httptest.NewRecorder()

	p.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}

	ct := rec.Header().Get("Content-Type")

	if ct != "application/problem+json" {
		t.Fatalf("expected content-type %q, got %q", "application/problem+json", ct)
	}

	var decoded map[string]any

	err := json.NewDecoder(rec.Body).Decode(&decoded)

	if err != nil {
		t.Fatalf("failed to decode response: %s", err)
	}

	if decoded["title"] != "Test Error" {
		t.Fatalf("expected title %q, got %v", "Test Error", decoded["title"])
	}
}

func TestServeHTTPJSON(t *testing.T) {
	t.Parallel()

	p := problem.Problem{
		Type:   "https://example.com/errors/test",
		Title:  "Test Error",
		Detail: "Details here",
		Status: http.StatusBadRequest,
	}

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Accept", "application/json")
	rec := httptest.NewRecorder()

	p.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}

	ct := rec.Header().Get("Content-Type")

	if ct != "application/json" {
		t.Fatalf("expected content-type %q, got %q", "application/json", ct)
	}

	var decoded map[string]any

	err := json.NewDecoder(rec.Body).Decode(&decoded)

	if err != nil {
		t.Fatalf("failed to decode response: %s", err)
	}

	if decoded["title"] != "Test Error" {
		t.Fatalf("expected title %q, got %v", "Test Error", decoded["title"])
	}
}

func TestServeHTTPTextFallback(t *testing.T) {
	t.Parallel()

	p := problem.Problem{
		Title:  "Test Error",
		Detail: "Something went wrong",
		Status: http.StatusBadRequest,
	}

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	p.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}

	body := rec.Body.String()

	if !strings.Contains(body, "400 Test Error") {
		t.Fatalf("expected body to contain %q, got %q", "400 Test Error", body)
	}

	if !strings.Contains(body, "Something went wrong") {
		t.Fatalf(
			"expected body to contain %q, got %q",
			"Something went wrong",
			body,
		)
	}
}

func TestServeHTTPTextFallbackWithUnknownAccept(t *testing.T) {
	t.Parallel()

	p := problem.Problem{
		Title:  "Test Error",
		Detail: "Something went wrong",
		Status: http.StatusBadRequest,
	}

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Accept", "text/html")
	rec := httptest.NewRecorder()

	p.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}

	body := rec.Body.String()

	if !strings.Contains(body, "400 Test Error") {
		t.Fatalf("expected text fallback, got %q", body)
	}
}

func TestServeHTTPTextWithStackTrace(t *testing.T) {
	t.Parallel()

	err := errors.New("root cause")
	p := problem.NewProblem(err, http.StatusInternalServerError)
	p.Title = "Server Error"
	p.Detail = "An error occurred"
	p = p.WithStackTrace()

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	p.ServeHTTP(rec, req)

	body := rec.Body.String()

	if !strings.Contains(body, "root cause") {
		t.Fatalf("expected stack trace in text body, got %q", body)
	}
}

func TestServeHTTPDev(t *testing.T) {
	t.Parallel()

	err := errors.New("debug error")
	p := problem.NewProblem(err, http.StatusInternalServerError)
	p.Title = "Server Error"

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Accept", "application/json")
	rec := httptest.NewRecorder()

	p.ServeHTTPDev(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusInternalServerError,
			rec.Code,
		)
	}

	var decoded map[string]any

	err2 := json.NewDecoder(rec.Body).Decode(&decoded)

	if err2 != nil {
		t.Fatalf("failed to decode response: %s", err2)
	}

	if _, ok := decoded[problem.StackTraceKey]; !ok {
		t.Fatalf("expected stack_trace in dev response")
	}
}

func TestServeHTTPDevTextFallback(t *testing.T) {
	t.Parallel()

	err := errors.New("debug text error")
	p := problem.NewProblem(err, http.StatusInternalServerError)
	p.Title = "Server Error"

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	p.ServeHTTPDev(rec, req)

	body := rec.Body.String()

	if !strings.Contains(body, "debug text error") {
		t.Fatalf("expected stack trace in dev text response, got %q", body)
	}
}

func TestHTTPStatus(t *testing.T) {
	t.Parallel()

	p := problem.Problem{Status: http.StatusTeapot}

	if p.HTTPStatus() != http.StatusTeapot {
		t.Fatalf("expected %d, got %d", http.StatusTeapot, p.HTTPStatus())
	}
}

func TestHTTPStatusZero(t *testing.T) {
	t.Parallel()

	p := problem.Problem{}

	if p.HTTPStatus() != 0 {
		t.Fatalf("expected 0, got %d", p.HTTPStatus())
	}
}

func TestServeHTTPDefaultsApplied(t *testing.T) {
	t.Parallel()

	p := problem.Problem{}

	req := httptest.NewRequest(http.MethodGet, "/my/path", nil)
	req.Header.Set("Accept", "application/problem+json")
	rec := httptest.NewRecorder()

	p.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusInternalServerError,
			rec.Code,
		)
	}

	var decoded map[string]any

	err := json.NewDecoder(rec.Body).Decode(&decoded)

	if err != nil {
		t.Fatalf("failed to decode response: %s", err)
	}

	if decoded["type"] != "about:blank" {
		t.Fatalf("expected type %q, got %v", "about:blank", decoded["type"])
	}

	if decoded["title"] != "Internal Server Error" {
		t.Fatalf(
			"expected title %q, got %v",
			"Internal Server Error",
			decoded["title"],
		)
	}

	if decoded["instance"] != "/my/path" {
		t.Fatalf("expected instance %q, got %v", "/my/path", decoded["instance"])
	}
}

func TestServeHTTPProblemJSONPreferredOverJSON(t *testing.T) {
	t.Parallel()

	p := problem.Problem{
		Status: http.StatusBadRequest,
		Title:  "Bad Request",
	}

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set(
		"Accept",
		"application/problem+json;q=1.0, application/json;q=0.9",
	)
	rec := httptest.NewRecorder()

	p.ServeHTTP(rec, req)

	ct := rec.Header().Get("Content-Type")

	if ct != "application/problem+json" {
		t.Fatalf("expected content-type %q, got %q", "application/problem+json", ct)
	}
}

func TestServeHTTPJSONPreferredOverProblemJSON(t *testing.T) {
	t.Parallel()

	p := problem.Problem{
		Status: http.StatusBadRequest,
		Title:  "Bad Request",
	}

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set(
		"Accept",
		"application/json;q=1.0, application/problem+json;q=0.5",
	)
	rec := httptest.NewRecorder()

	p.ServeHTTP(rec, req)

	ct := rec.Header().Get("Content-Type")

	if ct != "application/json" {
		t.Fatalf("expected content-type %q, got %q", "application/json", ct)
	}
}

func TestMarshalJSONEmpty(t *testing.T) {
	t.Parallel()

	p := problem.Problem{}

	data, err := json.Marshal(p)

	if err != nil {
		t.Fatalf("failed to marshal: %s", err)
	}

	var decoded map[string]any

	err = json.Unmarshal(data, &decoded)

	if err != nil {
		t.Fatalf("failed to unmarshal result: %s", err)
	}

	if decoded["type"] != "" {
		t.Fatalf("expected empty type, got %v", decoded["type"])
	}

	if decoded["status"] != float64(0) {
		t.Fatalf("expected status 0, got %v", decoded["status"])
	}
}

func TestMarshalUnmarshalRoundTrip(t *testing.T) {
	t.Parallel()

	original := problem.Problem{
		Type:     "https://example.com/errors/test",
		Title:    "Test Error",
		Detail:   "A detailed description",
		Status:   http.StatusConflict,
		Instance: "/resources/42",
	}

	original = original.With("trace_id", "abc123")

	data, err := json.Marshal(original)

	if err != nil {
		t.Fatalf("failed to marshal: %s", err)
	}

	var restored problem.Problem

	err = json.Unmarshal(data, &restored)

	if err != nil {
		t.Fatalf("failed to unmarshal: %s", err)
	}

	if restored.Type != original.Type {
		t.Fatalf("type mismatch: %q vs %q", original.Type, restored.Type)
	}

	if restored.Title != original.Title {
		t.Fatalf("title mismatch: %q vs %q", original.Title, restored.Title)
	}

	if restored.Detail != original.Detail {
		t.Fatalf("detail mismatch: %q vs %q", original.Detail, restored.Detail)
	}

	if restored.Status != original.Status {
		t.Fatalf("status mismatch: %d vs %d", original.Status, restored.Status)
	}

	if restored.Instance != original.Instance {
		t.Fatalf(
			"instance mismatch: %q vs %q",
			original.Instance,
			restored.Instance,
		)
	}

	val, ok := restored.Additional("trace_id")

	if !ok {
		t.Fatalf("expected trace_id in additional fields")
	}

	if val != "abc123" {
		t.Fatalf("expected trace_id %q, got %v", "abc123", val)
	}
}

func TestTextHandlerWithStackTraceNonStringSlice(t *testing.T) {
	t.Parallel()

	p := problem.Problem{
		Status: http.StatusInternalServerError,
		Title:  "Error",
		Detail: "Detail",
	}

	p = p.With(problem.StackTraceKey, 12345)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	p.ServeHTTP(rec, req)

	body := rec.Body.String()

	if !strings.Contains(body, "500 Error") {
		t.Fatalf("expected text output, got %q", body)
	}

	if strings.Contains(body, "12345") {
		t.Fatalf("non-string stack trace should not appear in text output")
	}
}

func TestProblemImplementsErrorInterface(t *testing.T) {
	t.Parallel()

	var err error = problem.Problem{
		Status: http.StatusBadRequest,
		Title:  "Bad Request",
	}

	if err.Error() == "" {
		t.Fatalf("expected non-empty error string")
	}
}

func TestWithStackTraceJoinedErrors(t *testing.T) {
	t.Parallel()

	err1 := errors.New("first error")
	err2 := errors.New("second error")
	joined := errors.Join(err1, err2)

	p := problem.NewProblem(joined, http.StatusInternalServerError)
	p = p.WithStackTrace()

	val, ok := p.Additional(problem.StackTraceKey)

	if !ok {
		t.Fatalf("expected stack_trace to be present")
	}

	traces, ok := val.([]string)

	if !ok {
		t.Fatalf("expected []string, got %T", val)
	}

	if len(traces) != 2 {
		t.Fatalf("expected 2 traces, got %d", len(traces))
	}
}

func TestUnmarshalJSONPartialFields(t *testing.T) {
	t.Parallel()

	raw := `{"status": 422, "detail": "Validation failed"}`

	var p problem.Problem

	err := json.Unmarshal([]byte(raw), &p)

	if err != nil {
		t.Fatalf("failed to unmarshal: %s", err)
	}

	if p.Status != 422 {
		t.Fatalf("expected status 422, got %d", p.Status)
	}

	if p.Detail != "Validation failed" {
		t.Fatalf("expected detail %q, got %q", "Validation failed", p.Detail)
	}

	if p.Type != "" {
		t.Fatalf("expected empty type, got %q", p.Type)
	}

	if p.Title != "" {
		t.Fatalf("expected empty title, got %q", p.Title)
	}
}

func TestMarshalJSONStandardFieldsCannotBeOverwritten(t *testing.T) {
	t.Parallel()

	p := problem.Problem{
		Type:     "https://example.com/errors/test",
		Title:    "Test Error",
		Detail:   "Real detail",
		Status:   http.StatusBadRequest,
		Instance: "/real/instance",
	}

	p = p.With("status", 999)
	p = p.With("type", "https://evil.com/hijack")
	p = p.With("title", "Hijacked Title")
	p = p.With("detail", "Hijacked detail")
	p = p.With("instance", "/hijacked/instance")

	data, err := json.Marshal(p)

	if err != nil {
		t.Fatalf("failed to marshal: %s", err)
	}

	var decoded map[string]any

	err = json.Unmarshal(data, &decoded)

	if err != nil {
		t.Fatalf("failed to unmarshal result: %s", err)
	}

	if decoded["status"] != float64(http.StatusBadRequest) {
		t.Fatalf("expected status %v, got %v", float64(http.StatusBadRequest), decoded["status"])
	}

	if decoded["type"] != "https://example.com/errors/test" {
		t.Fatalf("expected type %q, got %v", "https://example.com/errors/test", decoded["type"])
	}

	if decoded["title"] != "Test Error" {
		t.Fatalf("expected title %q, got %v", "Test Error", decoded["title"])
	}

	if decoded["detail"] != "Real detail" {
		t.Fatalf("expected detail %q, got %v", "Real detail", decoded["detail"])
	}

	if decoded["instance"] != "/real/instance" {
		t.Fatalf("expected instance %q, got %v", "/real/instance", decoded["instance"])
	}
}

func TestServeHTTPWithAcceptWildcardReturnsJSON(t *testing.T) {
	t.Parallel()

	p := problem.Problem{
		Type:   "https://example.com/errors/test",
		Title:  "Test Error",
		Detail: "Details here",
		Status: http.StatusBadRequest,
	}

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Accept", "*/*")
	rec := httptest.NewRecorder()

	p.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}

	ct := rec.Header().Get("Content-Type")

	if ct != "application/json" {
		t.Fatalf("expected content-type %q, got %q", "application/json", ct)
	}

	var decoded map[string]any

	err := json.NewDecoder(rec.Body).Decode(&decoded)

	if err != nil {
		t.Fatalf("failed to decode response: %s", err)
	}

	if decoded["title"] != "Test Error" {
		t.Fatalf("expected title %q, got %v", "Test Error", decoded["title"])
	}
}

func TestUnmarshalJSONWrongFieldTypes(t *testing.T) {
	t.Parallel()

	raw := `{
		"type": 123,
		"title": true,
		"detail": null,
		"status": "not a number",
		"instance": ["array"]
	}`

	var p problem.Problem

	err := json.Unmarshal([]byte(raw), &p)

	if err != nil {
		t.Fatalf("failed to unmarshal: %s", err)
	}

	if p.Type != "" {
		t.Fatalf("expected empty type for wrong type, got %q", p.Type)
	}

	if p.Title != "" {
		t.Fatalf("expected empty title for wrong type, got %q", p.Title)
	}

	if p.Detail != "" {
		t.Fatalf("expected empty detail for wrong type, got %q", p.Detail)
	}

	if p.Status != 0 {
		t.Fatalf("expected zero status for wrong type, got %d", p.Status)
	}

	if p.Instance != "" {
		t.Fatalf("expected empty instance for wrong type, got %q", p.Instance)
	}
}
