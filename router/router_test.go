package router_test

import (
	"net/http"
	"testing"

	"github.com/studiolambda/cosmos/router"
)

func TestRouterHas(t *testing.T) {
	r := router.New[http.HandlerFunc]()

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	r.Get("/foo/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	r.Get("/bar", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	r.Get("/baz/{other...}", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	if route := "/"; !r.Has(http.MethodGet, route) {
		t.Fatalf("router should have the %s route", route)
	}

	if route := "/foo"; !r.Has(http.MethodGet, route) {
		t.Fatalf("router should have the %s route", route)
	}

	if route := "/foo/"; !r.Has(http.MethodGet, route) {
		t.Fatalf("router should have the %s route", route)
	}

	if route := "/bar"; !r.Has(http.MethodGet, route) {
		t.Fatalf("router should have the %s route", route)
	}

	if route := "/bar/"; !r.Has(http.MethodGet, route) {
		t.Fatalf("router should have the %s route", route)
	}

	if route := "/baz"; !r.Has(http.MethodGet, route) {
		t.Fatalf("router should have the %s route", route)
	}

	if route := "/baz/"; !r.Has(http.MethodGet, route) {
		t.Fatalf("router should have the %s route", route)
	}

	if route := "/baz/foo/bar/baz"; !r.Has(http.MethodGet, route) {
		t.Fatalf("router should have the %s route", route)
	}

	if route := "/baz/foo/bar/baz/"; !r.Has(http.MethodGet, route) {
		t.Fatalf("router should have the %s route", route)
	}

	if route := "/not-found"; r.Has(http.MethodGet, route) {
		t.Fatalf("router should not have the %s route", route)
	}
}

func TestRouterMatches(t *testing.T) {
	r := router.New[http.HandlerFunc]()

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	request, err := http.NewRequest(http.MethodGet, "/", nil)

	if err != nil {
		t.Fatalf("unable to create http request: %v", err)
	}

	if !r.Matches(request) {
		t.Fatal("router does not have the route")
	}
}

func TestRouterHandler(t *testing.T) {
	r := router.New[http.HandlerFunc]()

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	request, err := http.NewRequest(http.MethodGet, "/", nil)

	if err != nil {
		t.Fatalf("failed to create http request: %v", err)
	}

	response := r.Record(request)

	if expected := http.StatusOK; response.StatusCode != expected {
		t.Fatalf("expected status code %d but got %d", expected, response.StatusCode)
	}
}
