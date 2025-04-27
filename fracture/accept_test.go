package fracture_test

import (
	"net/http"
	"testing"

	"github.com/studiolambda/cosmos/fracture"
)

func TestAcceptAccepts(t *testing.T) {
	request, err := http.NewRequest("GET", "/", nil)

	if err != nil {
		t.Fatalf("failed to create request: %s", err)
	}

	request.Header.Add("Accept", "application/json, text/*")

	accept := fracture.ParseAccept(request)

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
	request, err := http.NewRequest("GET", "/", nil)

	if err != nil {
		t.Fatalf("failed to create request: %s", err)
	}

	request.Header.Add("Accept", "application/json")
	request.Header.Add("Accept", "text/*")

	accept := fracture.ParseAccept(request)

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
	request, err := http.NewRequest("GET", "/", nil)

	if err != nil {
		t.Fatalf("failed to create request: %s", err)
	}

	request.Header.Add("Accept", "application/json, text/*, foo/bar;q=0.3, another/*;q=0.4, bar/baz;q=0.5")

	accept := fracture.ParseAccept(request)
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
