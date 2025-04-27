package orbit_test

import (
	"net/http"
	"testing"

	"github.com/studiolambda/cosmos/orbit"
)

func TestRouterHas(t *testing.T) {
	router := orbit.New[http.HandlerFunc]()

	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	router.Get("/foo/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	router.Get("/bar", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	router.Get("/baz/{other...}", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	if route := "/"; !router.Has(http.MethodGet, route) {
		t.Fatalf("router should have the %s route", route)
	}

	if route := "/foo"; !router.Has(http.MethodGet, route) {
		t.Fatalf("router should have the %s route", route)
	}

	if route := "/foo/"; !router.Has(http.MethodGet, route) {
		t.Fatalf("router should have the %s route", route)
	}

	if route := "/bar"; !router.Has(http.MethodGet, route) {
		t.Fatalf("router should have the %s route", route)
	}

	if route := "/bar/"; !router.Has(http.MethodGet, route) {
		t.Fatalf("router should have the %s route", route)
	}

	if route := "/baz"; !router.Has(http.MethodGet, route) {
		t.Fatalf("router should have the %s route", route)
	}

	if route := "/baz/"; !router.Has(http.MethodGet, route) {
		t.Fatalf("router should have the %s route", route)
	}

	if route := "/baz/foo/bar/baz"; !router.Has(http.MethodGet, route) {
		t.Fatalf("router should have the %s route", route)
	}

	if route := "/baz/foo/bar/baz/"; !router.Has(http.MethodGet, route) {
		t.Fatalf("router should have the %s route", route)
	}

	if route := "/not-found"; router.Has(http.MethodGet, route) {
		t.Fatalf("router should not have the %s route", route)
	}
}

func TestRouterMatches(t *testing.T) {
	router := orbit.New[http.HandlerFunc]()

	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	request, err := http.NewRequest(http.MethodGet, "/", nil)

	if err != nil {
		t.Fatalf("unable to create http request: %v", err)
	}

	if !router.Matches(request) {
		t.Fatal("router does not have the route")
	}
}

func TestRouterHandler(t *testing.T) {
	router := orbit.New[http.HandlerFunc]()

	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	request, err := http.NewRequest(http.MethodGet, "/", nil)

	if err != nil {
		t.Fatalf("failed to create http request: %v", err)
	}

	response := router.Record(request)

	if expected := http.StatusOK; response.Code != expected {
		t.Fatalf("expected status code %d but got %d", expected, response.Code)
	}
}
