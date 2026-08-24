# Router

`router` is a generic HTTP router built on `http.ServeMux`.

```bash
go get github.com/studiolambda/cosmos/router
```

## Constructor

```go
r := router.New[http.Handler]()
```

Type parameter is required (`H` must satisfy `http.Handler`).

---

## Registering routes

```go
r.Get("/users/{id}", getUser)
r.Post("/users", createUser)
r.Put("/users/{id}", updateUser)
r.Delete("/users/{id}", deleteUser)

r.Method(http.MethodPatch, "/users/{id}", patchUser)
r.Methods([]string{http.MethodGet, http.MethodHead}, "/assets/{path...}", assets)

r.Any("/health", health)
```

`Any` registers: GET, HEAD, POST, PUT, PATCH, DELETE, OPTIONS.
It intentionally excludes TRACE and CONNECT.

Patterns follow `http.ServeMux` syntax: `{id}`, `{path...}`.

Trailing slash handling is auto-registered for route pairs.

---

## Middleware

```go
type Middleware[H http.Handler] = func(H) H
```

### `Use` mutates current router

```go
r.Use(logging, recover)
```

### `With` returns a new sub-router

```go
authed := r.With(auth)
authed.Get("/me", me)
```

---

## Grouping

```go
r.Group("/api/v1", func(api *router.Router[http.Handler]) {
	api.Get("/users", listUsers)
	api.Get("/users/{id}", getUser)
})

r.Grouped(func(group *router.Router[http.Handler]) {
	group.Use(adminOnly)
	group.Delete("/users/{id}", deleteUser)
})
```

`Group` prefixes paths. `Grouped` scopes middleware without changing path prefix.

`Clone()` returns a sub-router sharing the same underlying mux with independent middleware slice.

---

## Inspection helpers

```go
exists := r.Has(http.MethodGet, "/users/{id}")
match := r.Matches(req)

h, ok := r.Handler(http.MethodGet, "/users/1")
h2, ok2 := r.HandlerMatch(req)

_ = exists
_ = match
_ = h
_ = ok
_ = h2
_ = ok2
```

---

## Testing with `Record`

```go
res := r.Record(httptest.NewRequest(http.MethodGet, "/health", nil))
defer res.Body.Close()
```

`Record` runs full router pipeline (middleware + handler) and returns `*http.Response`.

---

## Framework integration

```go
r := router.New[framework.Handler]()
r.Get("/users/{id}", func(w http.ResponseWriter, r *http.Request) error {
	id, err := request.ParamInt(r, "id")
	if err != nil {
		return err
	}
	return response.JSON(w, http.StatusOK, map[string]int{"id": id})
})
```

---

## Gotchas

- `router.New[...]` type parameter cannot be inferred; provide it explicitly.
- `Use()` mutates current router; `With()` creates a new one.
- `path.Join` normalization applies to grouped/joined patterns; `..` segments are rejected/purged.
- Catch-all routes (`{path...}`) also register base path without wildcard segment.
