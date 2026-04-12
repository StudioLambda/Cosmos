package router_test

import (
	"net/http"
	"net/http/httptest"
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

func TestPostRegistersRoute(t *testing.T) {
	t.Parallel()

	rt := router.New[http.HandlerFunc]()

	rt.Post("/items", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	})

	if !rt.Has(http.MethodPost, "/items") {
		t.Fatal("router should have POST /items")
	}

	req := httptest.NewRequest(http.MethodPost, "/items", nil)
	res := rt.Record(req)

	if res.StatusCode != http.StatusCreated {
		t.Fatalf("expected status %d but got %d", http.StatusCreated, res.StatusCode)
	}
}

func TestPutRegistersRoute(t *testing.T) {
	t.Parallel()

	rt := router.New[http.HandlerFunc]()

	rt.Put("/items/{id}", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	if !rt.Has(http.MethodPut, "/items/42") {
		t.Fatal("router should have PUT /items/{id}")
	}

	req := httptest.NewRequest(http.MethodPut, "/items/42", nil)
	res := rt.Record(req)

	if res.StatusCode != http.StatusNoContent {
		t.Fatalf("expected status %d but got %d", http.StatusNoContent, res.StatusCode)
	}
}

func TestPatchRegistersRoute(t *testing.T) {
	t.Parallel()

	rt := router.New[http.HandlerFunc]()

	rt.Patch("/items/{id}", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	if !rt.Has(http.MethodPatch, "/items/1") {
		t.Fatal("router should have PATCH /items/{id}")
	}
}

func TestDeleteRegistersRoute(t *testing.T) {
	t.Parallel()

	rt := router.New[http.HandlerFunc]()

	rt.Delete("/items/{id}", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	if !rt.Has(http.MethodDelete, "/items/1") {
		t.Fatal("router should have DELETE /items/{id}")
	}
}

func TestHeadRegistersRoute(t *testing.T) {
	t.Parallel()

	rt := router.New[http.HandlerFunc]()

	rt.Head("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	if !rt.Has(http.MethodHead, "/health") {
		t.Fatal("router should have HEAD /health")
	}
}

func TestConnectRegistersRoute(t *testing.T) {
	t.Parallel()

	rt := router.New[http.HandlerFunc]()

	rt.Connect("/proxy", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	if !rt.Has(http.MethodConnect, "/proxy") {
		t.Fatal("router should have CONNECT /proxy")
	}
}

func TestOptionsRegistersRoute(t *testing.T) {
	t.Parallel()

	rt := router.New[http.HandlerFunc]()

	rt.Options("/items", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	if !rt.Has(http.MethodOptions, "/items") {
		t.Fatal("router should have OPTIONS /items")
	}
}

func TestTraceRegistersRoute(t *testing.T) {
	t.Parallel()

	rt := router.New[http.HandlerFunc]()

	rt.Trace("/debug", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	if !rt.Has(http.MethodTrace, "/debug") {
		t.Fatal("router should have TRACE /debug")
	}
}

func TestMethodsRegistersMultipleMethods(t *testing.T) {
	t.Parallel()

	rt := router.New[http.HandlerFunc]()

	methods := []string{http.MethodGet, http.MethodPost}

	rt.Methods(methods, "/items", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	if !rt.Has(http.MethodGet, "/items") {
		t.Fatal("router should have GET /items")
	}

	if !rt.Has(http.MethodPost, "/items") {
		t.Fatal("router should have POST /items")
	}

	if rt.Has(http.MethodDelete, "/items") {
		t.Fatal("router should not have DELETE /items")
	}
}

func TestAnyRegistersAllStandardMethods(t *testing.T) {
	t.Parallel()

	rt := router.New[http.HandlerFunc]()

	rt.Any("/webhook", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	})

	expected := []string{
		http.MethodGet,
		http.MethodHead,
		http.MethodPost,
		http.MethodPut,
		http.MethodPatch,
		http.MethodDelete,
		http.MethodOptions,
	}

	for _, method := range expected {
		if !rt.Has(method, "/webhook") {
			t.Fatalf("router should have %s /webhook", method)
		}
	}

	// TRACE and CONNECT are intentionally excluded from Any().
	if rt.Has(http.MethodTrace, "/webhook") {
		t.Fatal("Any() should not register TRACE")
	}

	if rt.Has(http.MethodConnect, "/webhook") {
		t.Fatal("Any() should not register CONNECT")
	}
}

func TestUseAppliesMiddleware(t *testing.T) {
	t.Parallel()

	rt := router.New[http.HandlerFunc]()

	rt.Use(func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Middleware", "applied")
			next(w, r)
		}
	})

	rt.Get("/test", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	res := rt.Record(req)

	if res.Header.Get("X-Middleware") != "applied" {
		t.Fatal("middleware header should be set")
	}

	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected status %d but got %d", http.StatusOK, res.StatusCode)
	}
}

func TestUseMiddlewareExecutionOrder(t *testing.T) {
	t.Parallel()

	rt := router.New[http.HandlerFunc]()

	rt.Use(func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Order", "A")
			next(w, r)
		}
	})

	rt.Use(func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Order", w.Header().Get("X-Order")+"B")
			next(w, r)
		}
	})

	rt.Get("/order", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Order", w.Header().Get("X-Order")+"H")
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/order", nil)
	res := rt.Record(req)

	if order := res.Header.Get("X-Order"); order != "ABH" {
		t.Fatalf("expected middleware order ABH but got %s", order)
	}
}

func TestUseEmptyMiddlewareList(t *testing.T) {
	t.Parallel()

	rt := router.New[http.HandlerFunc]()

	rt.Use()

	rt.Get("/empty", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/empty", nil)
	res := rt.Record(req)

	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected status %d but got %d", http.StatusOK, res.StatusCode)
	}
}

func TestWithCreatesSubRouterWithMiddleware(t *testing.T) {
	t.Parallel()

	rt := router.New[http.HandlerFunc]()

	sub := rt.With(func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Auth", "true")
			next(w, r)
		}
	})

	sub.Get("/protected", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	rt.Get("/public", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Protected route should have the middleware header.
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	res := rt.Record(req)

	if res.Header.Get("X-Auth") != "true" {
		t.Fatal("protected route should have X-Auth header")
	}

	// Public route should not have the middleware header.
	req = httptest.NewRequest(http.MethodGet, "/public", nil)
	res = rt.Record(req)

	if res.Header.Get("X-Auth") != "" {
		t.Fatal("public route should not have X-Auth header")
	}
}

func TestWithDoesNotModifyParentRouter(t *testing.T) {
	t.Parallel()

	rt := router.New[http.HandlerFunc]()

	rt.Use(func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Parent", "true")
			next(w, r)
		}
	})

	_ = rt.With(func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Child", "true")
			next(w, r)
		}
	})

	rt.Get("/parent-route", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/parent-route", nil)
	res := rt.Record(req)

	if res.Header.Get("X-Parent") != "true" {
		t.Fatal("parent middleware should be applied")
	}

	if res.Header.Get("X-Child") != "" {
		t.Fatal("child middleware should not affect parent routes")
	}
}

func TestWithInheritsParentMiddleware(t *testing.T) {
	t.Parallel()

	rt := router.New[http.HandlerFunc]()

	rt.Use(func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Base", "true")
			next(w, r)
		}
	})

	sub := rt.With(func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Extra", "true")
			next(w, r)
		}
	})

	sub.Get("/both", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/both", nil)
	res := rt.Record(req)

	if res.Header.Get("X-Base") != "true" {
		t.Fatal("sub-router should inherit parent middleware")
	}

	if res.Header.Get("X-Extra") != "true" {
		t.Fatal("sub-router should apply its own middleware")
	}
}

func TestGroupPrefixesRoutes(t *testing.T) {
	t.Parallel()

	rt := router.New[http.HandlerFunc]()

	rt.Group("/api", func(api *router.Router[http.HandlerFunc]) {
		api.Get("/users", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
	})

	if !rt.Has(http.MethodGet, "/api/users") {
		t.Fatal("router should have GET /api/users")
	}

	if rt.Has(http.MethodGet, "/users") {
		t.Fatal("router should not have GET /users without prefix")
	}
}

func TestGroupInheritsMiddleware(t *testing.T) {
	t.Parallel()

	rt := router.New[http.HandlerFunc]()

	rt.Use(func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Global", "true")
			next(w, r)
		}
	})

	rt.Group("/api", func(api *router.Router[http.HandlerFunc]) {
		api.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
	})

	req := httptest.NewRequest(http.MethodGet, "/api/ping", nil)
	res := rt.Record(req)

	if res.Header.Get("X-Global") != "true" {
		t.Fatal("group should inherit parent middleware")
	}
}

func TestGroupWithLocalMiddleware(t *testing.T) {
	t.Parallel()

	rt := router.New[http.HandlerFunc]()

	rt.Group("/admin", func(admin *router.Router[http.HandlerFunc]) {
		admin.Use(func(next http.HandlerFunc) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("X-Admin", "true")
				next(w, r)
			}
		})

		admin.Get("/dashboard", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
	})

	rt.Get("/public", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/admin/dashboard", nil)
	res := rt.Record(req)

	if res.Header.Get("X-Admin") != "true" {
		t.Fatal("group middleware should apply to group routes")
	}

	req = httptest.NewRequest(http.MethodGet, "/public", nil)
	res = rt.Record(req)

	if res.Header.Get("X-Admin") != "" {
		t.Fatal("group middleware should not leak to parent routes")
	}
}

func TestNestedGroups(t *testing.T) {
	t.Parallel()

	rt := router.New[http.HandlerFunc]()

	rt.Group("/api", func(api *router.Router[http.HandlerFunc]) {
		api.Group("/v1", func(v1 *router.Router[http.HandlerFunc]) {
			v1.Get("/users", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})
		})
	})

	if !rt.Has(http.MethodGet, "/api/v1/users") {
		t.Fatal("router should have GET /api/v1/users")
	}
}

func TestGroupedClonesRouter(t *testing.T) {
	t.Parallel()

	rt := router.New[http.HandlerFunc]()

	rt.Grouped(func(sub *router.Router[http.HandlerFunc]) {
		sub.Use(func(next http.HandlerFunc) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("X-Grouped", "true")
				next(w, r)
			}
		})

		sub.Get("/grouped-route", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
	})

	rt.Get("/ungrouped-route", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// The grouped route should have its middleware applied.
	req := httptest.NewRequest(http.MethodGet, "/grouped-route", nil)
	res := rt.Record(req)

	if res.Header.Get("X-Grouped") != "true" {
		t.Fatal("grouped route should have X-Grouped header")
	}

	// The ungrouped route should not have the grouped middleware.
	req = httptest.NewRequest(http.MethodGet, "/ungrouped-route", nil)
	res = rt.Record(req)

	if res.Header.Get("X-Grouped") != "" {
		t.Fatal("ungrouped route should not have X-Grouped header")
	}
}

func TestCloneCreatesIndependentRouter(t *testing.T) {
	t.Parallel()

	rt := router.New[http.HandlerFunc]()

	clone := rt.Clone()

	clone.Use(func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Clone", "true")
			next(w, r)
		}
	})

	clone.Get("/clone-route", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	rt.Get("/original-route", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Clone route should have the clone middleware.
	req := httptest.NewRequest(http.MethodGet, "/clone-route", nil)
	res := rt.Record(req)

	if res.Header.Get("X-Clone") != "true" {
		t.Fatal("clone route should have X-Clone header")
	}

	// Original route should not have the clone middleware.
	req = httptest.NewRequest(http.MethodGet, "/original-route", nil)
	res = rt.Record(req)

	if res.Header.Get("X-Clone") != "" {
		t.Fatal("original route should not have X-Clone header")
	}
}

func TestCloneSharesServeMux(t *testing.T) {
	t.Parallel()

	rt := router.New[http.HandlerFunc]()

	clone := rt.Clone()

	clone.Get("/from-clone", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Routes registered on clone should be visible from the parent.
	if !rt.Has(http.MethodGet, "/from-clone") {
		t.Fatal("parent should see routes registered on clone")
	}
}

func TestHandlerReturnsRegisteredHandler(t *testing.T) {
	t.Parallel()

	rt := router.New[http.HandlerFunc]()

	rt.Get("/found", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler, ok := rt.Handler(http.MethodGet, "/found")

	if !ok {
		t.Fatal("Handler should find GET /found")
	}

	if handler == nil {
		t.Fatal("returned handler should not be nil")
	}
}

func TestHandlerNotFound(t *testing.T) {
	t.Parallel()

	rt := router.New[http.HandlerFunc]()

	_, ok := rt.Handler(http.MethodGet, "/missing")

	if ok {
		t.Fatal("Handler should not find unregistered route")
	}
}

func TestHandlerMatchReturnsHandler(t *testing.T) {
	t.Parallel()

	rt := router.New[http.HandlerFunc]()

	rt.Post("/submit", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	})

	req := httptest.NewRequest(http.MethodPost, "/submit", nil)
	handler, ok := rt.HandlerMatch(req)

	if !ok {
		t.Fatal("HandlerMatch should find POST /submit")
	}

	if handler == nil {
		t.Fatal("returned handler should not be nil")
	}
}

func TestHandlerMatchNotFound(t *testing.T) {
	t.Parallel()

	rt := router.New[http.HandlerFunc]()

	req := httptest.NewRequest(http.MethodGet, "/nowhere", nil)
	_, ok := rt.HandlerMatch(req)

	if ok {
		t.Fatal("HandlerMatch should not find unregistered route")
	}
}

func TestMatchesNoMatch(t *testing.T) {
	t.Parallel()

	rt := router.New[http.HandlerFunc]()

	rt.Get("/exists", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/exists", nil)

	if rt.Matches(req) {
		t.Fatal("should not match POST when only GET is registered")
	}
}

func TestRecordReturnsResponse(t *testing.T) {
	t.Parallel()

	rt := router.New[http.HandlerFunc]()

	rt.Get("/hello", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("hello world"))
	})

	req := httptest.NewRequest(http.MethodGet, "/hello", nil)
	res := rt.Record(req)

	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected status %d but got %d", http.StatusOK, res.StatusCode)
	}

	if ct := res.Header.Get("Content-Type"); ct != "text/plain" {
		t.Fatalf("expected Content-Type text/plain but got %s", ct)
	}
}

func TestServeHTTPDispatchesRequest(t *testing.T) {
	t.Parallel()

	rt := router.New[http.HandlerFunc]()

	rt.Get("/serve", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	})

	req := httptest.NewRequest(http.MethodGet, "/serve", nil)
	rr := httptest.NewRecorder()

	rt.ServeHTTP(rr, req)

	if rr.Code != http.StatusAccepted {
		t.Fatalf("expected status %d but got %d", http.StatusAccepted, rr.Code)
	}
}

func TestRootRouteExactMatch(t *testing.T) {
	t.Parallel()

	rt := router.New[http.HandlerFunc]()

	rt.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Root should match.
	if !rt.Has(http.MethodGet, "/") {
		t.Fatal("router should have GET /")
	}

	// Arbitrary nested path should not match root.
	req := httptest.NewRequest(http.MethodGet, "/anything", nil)
	res := rt.Record(req)

	if res.StatusCode == http.StatusOK {
		t.Fatal("root route should not catch /anything")
	}
}

func TestTrailingSlashRouteMatchesBoth(t *testing.T) {
	t.Parallel()

	rt := router.New[http.HandlerFunc]()

	rt.Get("/items/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	if !rt.Has(http.MethodGet, "/items") {
		t.Fatal("trailing slash route should match without slash")
	}

	if !rt.Has(http.MethodGet, "/items/") {
		t.Fatal("trailing slash route should match with slash")
	}
}

func TestNonTrailingSlashRouteMatchesBoth(t *testing.T) {
	t.Parallel()

	rt := router.New[http.HandlerFunc]()

	rt.Get("/items", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	if !rt.Has(http.MethodGet, "/items") {
		t.Fatal("should match without trailing slash")
	}

	if !rt.Has(http.MethodGet, "/items/") {
		t.Fatal("should match with trailing slash")
	}
}

func TestCatchAllWildcard(t *testing.T) {
	t.Parallel()

	rt := router.New[http.HandlerFunc]()

	rt.Get("/files/{path...}", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	if !rt.Has(http.MethodGet, "/files/a/b/c") {
		t.Fatal("catch-all should match nested path")
	}

	if !rt.Has(http.MethodGet, "/files/") {
		t.Fatal("catch-all should match base with trailing slash")
	}

	if !rt.Has(http.MethodGet, "/files") {
		t.Fatal("catch-all should match base without trailing slash")
	}
}

func TestGroupRootRoute(t *testing.T) {
	t.Parallel()

	rt := router.New[http.HandlerFunc]()

	rt.Group("/api", func(api *router.Router[http.HandlerFunc]) {
		api.Get("/", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
	})

	if !rt.Has(http.MethodGet, "/api") {
		t.Fatal("group root should match /api")
	}

	if !rt.Has(http.MethodGet, "/api/") {
		t.Fatal("group root should match /api/")
	}
}

func TestGroupWithCatchAll(t *testing.T) {
	t.Parallel()

	rt := router.New[http.HandlerFunc]()

	rt.Group("/static", func(s *router.Router[http.HandlerFunc]) {
		s.Get("/{file...}", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
	})

	if !rt.Has(http.MethodGet, "/static/css/main.css") {
		t.Fatal("group catch-all should match nested paths")
	}
}

func TestMiddlewareWrapsInReverseOrder(t *testing.T) {
	t.Parallel()

	rt := router.New[http.HandlerFunc]()

	var order []string

	rt.Use(func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			order = append(order, "A-before")
			next(w, r)
			order = append(order, "A-after")
		}
	})

	rt.Use(func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			order = append(order, "B-before")
			next(w, r)
			order = append(order, "B-after")
		}
	})

	rt.Get("/mw", func(w http.ResponseWriter, r *http.Request) {
		order = append(order, "handler")
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/mw", nil)
	rt.Record(req)

	expected := []string{
		"A-before", "B-before", "handler", "B-after", "A-after",
	}

	if len(order) != len(expected) {
		t.Fatalf("expected %d calls but got %d", len(expected), len(order))
	}

	for i, v := range expected {
		if order[i] != v {
			t.Fatalf("position %d: expected %s but got %s", i, v, order[i])
		}
	}
}

func TestMethodRegistration(t *testing.T) {
	t.Parallel()

	rt := router.New[http.HandlerFunc]()

	rt.Method(http.MethodPut, "/resource", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	if !rt.Has(http.MethodPut, "/resource") {
		t.Fatal("Method() should register PUT /resource")
	}

	req := httptest.NewRequest(http.MethodPut, "/resource", nil)
	res := rt.Record(req)

	if res.StatusCode != http.StatusNoContent {
		t.Fatalf(
			"expected status %d but got %d",
			http.StatusNoContent,
			res.StatusCode,
		)
	}
}

func TestHandlerExecutesCorrectHandler(t *testing.T) {
	t.Parallel()

	rt := router.New[http.HandlerFunc]()

	rt.Get("/a", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	rt.Get("/b", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	})

	reqA := httptest.NewRequest(http.MethodGet, "/a", nil)
	resA := rt.Record(reqA)

	if resA.StatusCode != http.StatusOK {
		t.Fatalf("expected %d for /a but got %d", http.StatusOK, resA.StatusCode)
	}

	reqB := httptest.NewRequest(http.MethodGet, "/b", nil)
	resB := rt.Record(reqB)

	if resB.StatusCode != http.StatusAccepted {
		t.Fatalf(
			"expected %d for /b but got %d",
			http.StatusAccepted,
			resB.StatusCode,
		)
	}
}

func TestGroupMiddlewareDoesNotAffectParent(t *testing.T) {
	t.Parallel()

	rt := router.New[http.HandlerFunc]()

	rt.Get("/outside", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	rt.Group("/scoped", func(s *router.Router[http.HandlerFunc]) {
		s.Use(func(next http.HandlerFunc) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("X-Scoped", "yes")
				next(w, r)
			}
		})

		s.Get("/inner", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
	})

	req := httptest.NewRequest(http.MethodGet, "/outside", nil)
	res := rt.Record(req)

	if res.Header.Get("X-Scoped") != "" {
		t.Fatal("group middleware should not affect routes outside the group")
	}

	req = httptest.NewRequest(http.MethodGet, "/scoped/inner", nil)
	res = rt.Record(req)

	if res.Header.Get("X-Scoped") != "yes" {
		t.Fatal("group middleware should apply to routes inside the group")
	}
}

func TestPathParameter(t *testing.T) {
	t.Parallel()

	rt := router.New[http.HandlerFunc]()

	rt.Get("/users/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(id))
	})

	req := httptest.NewRequest(http.MethodGet, "/users/123", nil)
	rr := httptest.NewRecorder()

	rt.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d but got %d", http.StatusOK, rr.Code)
	}

	if body := rr.Body.String(); body != "123" {
		t.Fatalf("expected body 123 but got %s", body)
	}
}

func TestDeeplyNestedGroups(t *testing.T) {
	t.Parallel()

	rt := router.New[http.HandlerFunc]()

	rt.Group("/a", func(a *router.Router[http.HandlerFunc]) {
		a.Group("/b", func(b *router.Router[http.HandlerFunc]) {
			b.Group("/c", func(c *router.Router[http.HandlerFunc]) {
				c.Get("/d", func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
				})
			})
		})
	})

	if !rt.Has(http.MethodGet, "/a/b/c/d") {
		t.Fatal("deeply nested group should register /a/b/c/d")
	}
}

func TestMultipleMiddlewareInWith(t *testing.T) {
	t.Parallel()

	rt := router.New[http.HandlerFunc]()

	mw1 := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-MW1", "true")
			next(w, r)
		}
	}

	mw2 := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-MW2", "true")
			next(w, r)
		}
	}

	sub := rt.With(mw1, mw2)

	sub.Get("/multi-mw", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/multi-mw", nil)
	res := rt.Record(req)

	if res.Header.Get("X-MW1") != "true" {
		t.Fatal("first middleware should be applied")
	}

	if res.Header.Get("X-MW2") != "true" {
		t.Fatal("second middleware should be applied")
	}
}

func TestWithEmptyMiddleware(t *testing.T) {
	t.Parallel()

	rt := router.New[http.HandlerFunc]()

	sub := rt.With()

	sub.Get("/with-empty", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/with-empty", nil)
	res := rt.Record(req)

	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected status %d but got %d", http.StatusOK, res.StatusCode)
	}
}

func TestAnyAllMethodsRespond(t *testing.T) {
	t.Parallel()

	rt := router.New[http.HandlerFunc]()

	rt.Any("/all", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	methods := []string{
		http.MethodGet,
		http.MethodHead,
		http.MethodPost,
		http.MethodPut,
		http.MethodPatch,
		http.MethodDelete,
		http.MethodOptions,
	}

	for _, method := range methods {
		req := httptest.NewRequest(method, "/all", nil)
		res := rt.Record(req)

		// HEAD responses don't have status written to body, but
		// the handler sets 200 which should be reflected.
		if res.StatusCode != http.StatusOK {
			t.Fatalf(
				"expected status %d for %s but got %d",
				http.StatusOK,
				method,
				res.StatusCode,
			)
		}
	}
}

func TestGroupedWithMiddlewareChain(t *testing.T) {
	t.Parallel()

	rt := router.New[http.HandlerFunc]()

	rt.Use(func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Parent", "true")
			next(w, r)
		}
	})

	rt.Grouped(func(sub *router.Router[http.HandlerFunc]) {
		sub.Use(func(next http.HandlerFunc) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("X-Sub", "true")
				next(w, r)
			}
		})

		sub.Get("/chained", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
	})

	req := httptest.NewRequest(http.MethodGet, "/chained", nil)
	res := rt.Record(req)

	if res.Header.Get("X-Parent") != "true" {
		t.Fatal("grouped should inherit parent middleware")
	}

	if res.Header.Get("X-Sub") != "true" {
		t.Fatal("grouped should apply its own middleware")
	}
}

func TestHasReturnsFalseForWrongMethod(t *testing.T) {
	t.Parallel()

	rt := router.New[http.HandlerFunc]()

	rt.Get("/only-get", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	if rt.Has(http.MethodPost, "/only-get") {
		t.Fatal("Has should return false for POST on a GET-only route")
	}
}

func TestHandlerReturnsCorrectHandlerForMethod(t *testing.T) {
	t.Parallel()

	rt := router.New[http.HandlerFunc]()

	rt.Get("/endpoint", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	rt.Post("/endpoint", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	})

	getHandler, ok := rt.Handler(http.MethodGet, "/endpoint")

	if !ok {
		t.Fatal("Handler should find GET /endpoint")
	}

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/endpoint", nil)
	getHandler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("GET handler should return %d but got %d", http.StatusOK, rr.Code)
	}

	postHandler, ok := rt.Handler(http.MethodPost, "/endpoint")

	if !ok {
		t.Fatal("Handler should find POST /endpoint")
	}

	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/endpoint", nil)
	postHandler.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf(
			"POST handler should return %d but got %d",
			http.StatusCreated,
			rr.Code,
		)
	}
}

func TestGroupNestedWithMiddleware(t *testing.T) {
	t.Parallel()

	rt := router.New[http.HandlerFunc]()

	rt.Group("/api", func(api *router.Router[http.HandlerFunc]) {
		api.Use(func(next http.HandlerFunc) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("X-API", "true")
				next(w, r)
			}
		})

		api.Group("/v2", func(v2 *router.Router[http.HandlerFunc]) {
			v2.Use(func(next http.HandlerFunc) http.HandlerFunc {
				return func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("X-V2", "true")
					next(w, r)
				}
			})

			v2.Get("/items", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})
		})
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v2/items", nil)
	res := rt.Record(req)

	if res.Header.Get("X-API") != "true" {
		t.Fatal("nested group should inherit parent group middleware")
	}

	if res.Header.Get("X-V2") != "true" {
		t.Fatal("nested group should apply its own middleware")
	}
}

func TestRecordNotFoundRoute(t *testing.T) {
	t.Parallel()

	rt := router.New[http.HandlerFunc]()

	rt.Get("/exists", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/does-not-exist", nil)
	res := rt.Record(req)

	if res.StatusCode == http.StatusOK {
		t.Fatal("unregistered route should not return 200")
	}
}

func TestMultipleRoutesOnSameRouter(t *testing.T) {
	t.Parallel()

	rt := router.New[http.HandlerFunc]()

	rt.Get("/one", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	rt.Post("/two", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	})

	rt.Delete("/three", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	if !rt.Has(http.MethodGet, "/one") {
		t.Fatal("should have GET /one")
	}

	if !rt.Has(http.MethodPost, "/two") {
		t.Fatal("should have POST /two")
	}

	if !rt.Has(http.MethodDelete, "/three") {
		t.Fatal("should have DELETE /three")
	}
}

func TestWithChaining(t *testing.T) {
	t.Parallel()

	rt := router.New[http.HandlerFunc]()

	mwA := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-A", "true")
			next(w, r)
		}
	}

	mwB := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-B", "true")
			next(w, r)
		}
	}

	sub := rt.With(mwA).With(mwB)

	sub.Get("/chained-with", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/chained-with", nil)
	res := rt.Record(req)

	if res.Header.Get("X-A") != "true" {
		t.Fatal("first With middleware should apply")
	}

	if res.Header.Get("X-B") != "true" {
		t.Fatal("second With middleware should apply")
	}
}

func TestCatchAllWildcardPathValue(t *testing.T) {
	t.Parallel()

	rt := router.New[http.HandlerFunc]()

	rt.Get("/assets/{filepath...}", func(w http.ResponseWriter, r *http.Request) {
		fp := r.PathValue("filepath")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fp))
	})

	req := httptest.NewRequest(http.MethodGet, "/assets/css/style.css", nil)
	rr := httptest.NewRecorder()

	rt.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d but got %d", http.StatusOK, rr.Code)
	}

	if body := rr.Body.String(); body != "css/style.css" {
		t.Fatalf("expected body css/style.css but got %s", body)
	}
}

func TestGroupWithWith(t *testing.T) {
	t.Parallel()

	rt := router.New[http.HandlerFunc]()

	rt.Group("/api", func(api *router.Router[http.HandlerFunc]) {
		auth := api.With(func(next http.HandlerFunc) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("X-Auth-Required", "true")
				next(w, r)
			}
		})

		auth.Get("/secret", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		api.Get("/public", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
	})

	req := httptest.NewRequest(http.MethodGet, "/api/secret", nil)
	res := rt.Record(req)

	if res.Header.Get("X-Auth-Required") != "true" {
		t.Fatal("With inside Group should apply middleware")
	}

	req = httptest.NewRequest(http.MethodGet, "/api/public", nil)
	res = rt.Record(req)

	if res.Header.Get("X-Auth-Required") != "" {
		t.Fatal("non-With route in Group should not have middleware")
	}
}

func TestHasReturnsFalseForInvalidMethod(t *testing.T) {
	t.Parallel()

	rt := router.New[http.HandlerFunc]()

	rt.Get("/test", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// A method with invalid characters causes http.NewRequest to fail,
	// triggering the error branch in Has().
	if rt.Has("INVALID METHOD", "/test") {
		t.Fatal("Has should return false when http.NewRequest fails")
	}
}

func TestHandlerReturnsFalseForInvalidMethod(t *testing.T) {
	t.Parallel()

	rt := router.New[http.HandlerFunc]()

	rt.Get("/test", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// A method with invalid characters causes http.NewRequest to fail,
	// triggering the error branch in Handler().
	_, ok := rt.Handler("INVALID METHOD", "/test")

	if ok {
		t.Fatal("Handler should return false when http.NewRequest fails")
	}
}

func TestCatchAllWithFewSegments(t *testing.T) {
	t.Parallel()

	rt := router.New[http.HandlerFunc]()

	// Register a catch-all at root level: "/{rest...}" which splits
	// to ["", "{rest...}"] — only 2 segments, so the len > 2 branch
	// in registerPair is skipped.
	rt.Get("/{rest...}", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	if !rt.Has(http.MethodGet, "/anything/at/all") {
		t.Fatal("root catch-all should match nested paths")
	}

	if !rt.Has(http.MethodGet, "/single") {
		t.Fatal("root catch-all should match single segment")
	}
}

func TestMethodPanicsOnDotDotPattern(t *testing.T) {
	t.Parallel()

	rt := router.New[http.HandlerFunc]()

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for pattern containing '..' but got none")
		}
	}()

	rt.Get("/../admin", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

func TestMethodPanicsOnEmptyMethod(t *testing.T) {
	t.Parallel()

	rt := router.New[http.HandlerFunc]()

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for empty method but got none")
		}
	}()

	rt.Method("", "/test", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}
