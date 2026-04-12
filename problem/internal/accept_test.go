package internal_test

import (
	"net/http"
	"testing"

	"github.com/studiolambda/cosmos/problem/internal"
)

func TestAcceptAccepts(t *testing.T) {
	t.Parallel()

	request, err := http.NewRequest("GET", "/", nil)

	if err != nil {
		t.Fatalf("failed to create request: %s", err)
	}

	request.Header.Add("Accept", "application/json, text/*")

	accept := internal.ParseAccept(request)

	if expected := "application/json"; !accept.Accepts(expected) {
		t.Fatalf("failed to accept media type: %s", expected)
	}

	if expected := "text/html"; !accept.Accepts(expected) {
		t.Fatalf("failed to accept media type: %s", expected)
	}

	if expected := "text/*"; !accept.Accepts(expected) {
		t.Fatalf("failed to accept media type: %s", expected)
	}

	if expected := "foo/bar"; accept.Accepts(expected) {
		t.Fatalf("failed to not accept media type: %s", expected)
	}
}

func TestAcceptAcceptsWithMultipleHeaderValues(t *testing.T) {
	t.Parallel()

	request, err := http.NewRequest("GET", "/", nil)

	if err != nil {
		t.Fatalf("failed to create request: %s", err)
	}

	request.Header.Add("Accept", "application/json")
	request.Header.Add("Accept", "text/*")

	accept := internal.ParseAccept(request)

	if expected := "application/json"; !accept.Accepts(expected) {
		t.Fatalf("failed to accept media type: %s", expected)
	}

	if expected := "text/html"; !accept.Accepts(expected) {
		t.Fatalf("failed to accept media type: %s", expected)
	}

	if expected := "text/*"; !accept.Accepts(expected) {
		t.Fatalf("failed to accept media type: %s", expected)
	}

	if expected := "foo/bar"; accept.Accepts(expected) {
		t.Fatalf("failed to not accept media type: %s", expected)
	}
}

func TestAcceptOrder(t *testing.T) {
	t.Parallel()

	request, err := http.NewRequest("GET", "/", nil)

	if err != nil {
		t.Fatalf("failed to create request: %s", err)
	}

	request.Header.Add("Accept", "application/json, text/*, foo/bar;q=0.3, another/*;q=0.4, bar/baz;q=0.5")

	accept := internal.ParseAccept(request)
	order := accept.Order()

	if expected := 5; expected != len(order) {
		t.Fatalf("failed order len: %d, expected %d", len(order), expected)
	}

	if expected := "application/json"; expected != order[0] {
		t.Fatalf("failed order element: %s, expected %s", order[0], expected)
	}

	if expected := "text/*"; expected != order[1] {
		t.Fatalf("failed order element: %s, expected %s", order[1], expected)
	}

	if expected := "bar/baz"; expected != order[2] {
		t.Fatalf("failed order element: %s, expected %s", order[2], expected)
	}

	if expected := "another/*"; expected != order[3] {
		t.Fatalf("failed order element: %s, expected %s", order[3], expected)
	}

	if expected := "foo/bar"; expected != order[4] {
		t.Fatalf("failed order element: %s, expected %s", order[4], expected)
	}
}

func TestAcceptQualityFound(t *testing.T) {
	t.Parallel()

	request, err := http.NewRequest("GET", "/", nil)

	if err != nil {
		t.Fatalf("failed to create request: %s", err)
	}

	request.Header.Add("Accept", "application/json;q=0.8")

	accept := internal.ParseAccept(request)
	quality := accept.Quality("application/json")

	if quality != 0.8 {
		t.Fatalf("expected quality 0.8, got %f", quality)
	}
}

func TestAcceptQualityNotFound(t *testing.T) {
	t.Parallel()

	request, err := http.NewRequest("GET", "/", nil)

	if err != nil {
		t.Fatalf("failed to create request: %s", err)
	}

	request.Header.Add("Accept", "application/json")

	accept := internal.ParseAccept(request)
	quality := accept.Quality("text/html")

	if quality != 0 {
		t.Fatalf("expected quality 0, got %f", quality)
	}
}

func TestAcceptQualityDefault(t *testing.T) {
	t.Parallel()

	request, err := http.NewRequest("GET", "/", nil)

	if err != nil {
		t.Fatalf("failed to create request: %s", err)
	}

	request.Header.Add("Accept", "application/json")

	accept := internal.ParseAccept(request)
	quality := accept.Quality("application/json")

	if quality != 1.0 {
		t.Fatalf("expected quality 1.0, got %f", quality)
	}
}

func TestParseAcceptMalformedMediaType(t *testing.T) {
	t.Parallel()

	request, err := http.NewRequest("GET", "/", nil)

	if err != nil {
		t.Fatalf("failed to create request: %s", err)
	}

	request.Header.Add("Accept", "a/b/c/d, application/json")

	accept := internal.ParseAccept(request)

	if !accept.Accepts("application/json") {
		t.Fatalf("expected application/json to be accepted")
	}

	order := accept.Order()

	if len(order) != 1 {
		t.Fatalf("expected 1 valid media type, got %d", len(order))
	}
}

func TestAcceptOrderEmpty(t *testing.T) {
	t.Parallel()

	request, err := http.NewRequest("GET", "/", nil)

	if err != nil {
		t.Fatalf("failed to create request: %s", err)
	}

	accept := internal.ParseAccept(request)
	order := accept.Order()

	if len(order) != 0 {
		t.Fatalf("expected empty order, got %d elements", len(order))
	}
}

func TestAcceptAcceptsWildcardInMedia(t *testing.T) {
	t.Parallel()

	request, err := http.NewRequest("GET", "/", nil)

	if err != nil {
		t.Fatalf("failed to create request: %s", err)
	}

	request.Header.Add("Accept", "application/json")

	accept := internal.ParseAccept(request)

	if !accept.Accepts("application/*") {
		t.Fatalf("expected application/* to match application/json")
	}
}

func TestAcceptWildcardDoesNotMatchAcrossSlash(t *testing.T) {
	t.Parallel()

	request, err := http.NewRequest("GET", "/", nil)

	if err != nil {
		t.Fatalf("failed to create request: %s", err)
	}

	request.Header.Add("Accept", "application/*")

	accept := internal.ParseAccept(request)

	if accept.Accepts("applicationx/json") {
		t.Fatalf("application/* should not match applicationx/json")
	}

	if !accept.Accepts("application/json") {
		t.Fatalf("application/* should match application/json")
	}
}

func TestAcceptFullWildcardMatchesAny(t *testing.T) {
	t.Parallel()

	request, err := http.NewRequest("GET", "/", nil)

	if err != nil {
		t.Fatalf("failed to create request: %s", err)
	}

	request.Header.Add("Accept", "*/*")

	accept := internal.ParseAccept(request)

	if !accept.Accepts("application/json") {
		t.Fatalf("expected */* to match application/json")
	}

	if !accept.Accepts("text/html") {
		t.Fatalf("expected */* to match text/html")
	}

	if !accept.Accepts("application/problem+json") {
		t.Fatalf("expected */* to match application/problem+json")
	}
}

func TestAcceptNoAcceptHeader(t *testing.T) {
	t.Parallel()

	request, err := http.NewRequest("GET", "/", nil)

	if err != nil {
		t.Fatalf("failed to create request: %s", err)
	}

	accept := internal.ParseAccept(request)

	if accept.Accepts("application/json") {
		t.Fatalf("expected no media to be accepted with empty Accept header")
	}
}
