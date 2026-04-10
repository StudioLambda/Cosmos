package router_test

import (
	"net/http"
	"testing"

	"github.com/studiolambda/cosmos/router"
)

func TestRouterHas(t *testing.T) {
	t.Parallel()

	rt := router.New[http.HandlerFunc]()

	rt.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	rt.Get("/foo/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	rt.Get("/bar", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	rt.Get("/baz/{other...}", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	if route := "/"; !rt.Has(http.MethodGet, route) {
		t.Fatalf("router should have the %s route", route)
	}

	if route := "/foo"; !rt.Has(http.MethodGet, route) {
		t.Fatalf("router should have the %s route", route)
	}

	if route := "/foo/"; !rt.Has(http.MethodGet, route) {
		t.Fatalf("router should have the %s route", route)
	}

	if route := "/bar"; !rt.Has(http.MethodGet, route) {
		t.Fatalf("router should have the %s route", route)
	}

	if route := "/bar/"; !rt.Has(http.MethodGet, route) {
		t.Fatalf("router should have the %s route", route)
	}

	if route := "/baz"; !rt.Has(http.MethodGet, route) {
		t.Fatalf("router should have the %s route", route)
	}

	if route := "/baz/"; !rt.Has(http.MethodGet, route) {
		t.Fatalf("router should have the %s route", route)
	}

	if route := "/baz/foo/bar/baz"; !rt.Has(http.MethodGet, route) {
		t.Fatalf("router should have the %s route", route)
	}

	if route := "/baz/foo/bar/baz/"; !rt.Has(http.MethodGet, route) {
		t.Fatalf("router should have the %s route", route)
	}

	if route := "/not-found"; rt.Has(http.MethodGet, route) {
		t.Fatalf("router should not have the %s route", route)
	}
}

func TestRouterMatches(t *testing.T) {
	t.Parallel()

	rt := router.New[http.HandlerFunc]()

	rt.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	request, err := http.NewRequest(http.MethodGet, "/", nil)

	if err != nil {
		t.Fatalf("unable to create http request: %v", err)
	}

	if !rt.Matches(request) {
		t.Fatal("router does not have the route")
	}
}

func TestRouterHandler(t *testing.T) {
	t.Parallel()

	rt := router.New[http.HandlerFunc]()

	rt.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	request, err := http.NewRequest(http.MethodGet, "/", nil)

	if err != nil {
		t.Fatalf("failed to create http request: %v", err)
	}

	response := rt.Record(request)

	if expected := http.StatusOK; response.StatusCode != expected {
		t.Fatalf("expected status code %d but got %d", expected, response.StatusCode)
	}
}
