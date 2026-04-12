package router

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"path"
	"slices"
	"strings"
)

// Middleware is a func type that can be used to apply middleware logic between requests and responses.
type Middleware[H http.Handler] = func(H) H

// Router is a generic HTTP router that wraps [http.ServeMux] and supports
// middleware, route groups, and automatic trailing-slash handling.
type Router[H http.Handler] struct {
	// native stores the actual [http.ServeMux]
	// that's used internally to register the routes.
	native *http.ServeMux

	// pattern stores the current pattern that will be
	// used as a prefix to all the route registrations
	// on this router. This pattern is already joined with
	// the parent router's pattern if any.
	pattern string

	// parent stores the parent [Router] if any. This is
	// used to correctly resolve the [http.ServeMux] to
	// use by sub-routers so that they all register the
	// routes to the same [http.ServeMux].
	parent *Router[H]

	// middlewares stores the actual middlewares that will
	// be applied to any route registration on the current
	// router. It already contains all the middlewares of
	// the parent's [Router] if any.
	middlewares []Middleware[H]
}

// allMethods stores the HTTP methods registered by [Router.Any].
// TRACE and CONNECT are intentionally excluded: TRACE enables
// cross-site tracing (XST) attacks that can leak credentials,
// and CONNECT is intended for HTTP proxies only.
var allMethods = []string{
	http.MethodGet,
	http.MethodHead,
	http.MethodPost,
	http.MethodPut,
	http.MethodPatch,
	http.MethodDelete,
	http.MethodOptions,
}

// New creates a new [Router] with an empty [http.ServeMux].
func New[H http.Handler]() *Router[H] {
	return &Router[H]{
		native:      http.NewServeMux(),
		middlewares: make([]Middleware[H], 0),
	}
}

// Group uses the given pattern to automatically
// mount a sub-router that has that pattern as a
// prefix.
//
// This means that any route registered with the
// sub-router will also have the given pattern prefixed.
//
// Keep in mind this can be nested as well, meaning that
// many sub-routers may be grouped, creating complex
// patterns.
//
// WARNING: The pattern is joined with the parent
// pattern using [path.Join], which normalizes the
// resulting path. This means ".." segments are resolved
// and collapsed. Callers should not rely on literal
// ".." segments in route patterns, as they will be
// cleaned away during path joining.
func (router *Router[H]) Group(pattern string, subrouter func(*Router[H])) {
	subrouter(&Router[H]{
		native:      nil, // parent's native will be used
		pattern:     path.Join(router.pattern, pattern),
		parent:      router,
		middlewares: slices.Clone(router.middlewares),
	})
}

// Grouped creates a cloned sub-router for scoping middleware without a path prefix.
func (router *Router[H]) Grouped(subrouter func(*Router[H])) {
	subrouter(router.Clone())
}

// Clone creates a sub-router that shares the same [http.ServeMux] but has an
// independent middleware stack.
func (router *Router[H]) Clone() *Router[H] {
	return &Router[H]{
		native:      nil, // parent's native will be used
		pattern:     router.pattern,
		parent:      router,
		middlewares: slices.Clone(router.middlewares),
	}
}

// With creates a new sub-router that applies the given middlewares in addition
// to any inherited ones.
//
// This is useful for inlining middleware on specific routes.
//
// In contrast to [Router.Use], it creates a new sub-router instead of
// modifying the current router.
func (router *Router[H]) With(middlewares ...Middleware[H]) *Router[H] {
	return &Router[H]{
		native:      nil, // parent's native will be used
		pattern:     router.pattern,
		parent:      router,
		middlewares: append(slices.Clone(router.middlewares), middlewares...),
	}
}

// mux returns the native [http.ServeMux] that is used
// internally by the router. This exists because sub-routers
// must use the same [http.ServeMux] and therefore, there's
// some recursion involved to get the same [http.ServeMux].
func (router *Router[H]) mux() *http.ServeMux {
	if router.parent != nil {
		return router.parent.mux()
	}

	return router.native
}

// wrap makes a handler wrapped by the current router's middleware.
// This means that the resulting handler is the same as first calling
// the router's middleware and then the provided handler.
func (router *Router[H]) wrap(handler H) H {
	for i := len(router.middlewares) - 1; i >= 0; i-- {
		handler = router.middlewares[i](handler)
	}

	return handler
}

// Use appends to the current router the given middlewares.
//
// Subsequent route registrations will be wrapped with any previous
// middlewares that the router had defined, plus the new ones
// that are registered after this call.
//
// WARNING: Middleware execution order matters. Security-critical
// middleware such as Recover, Logger, Secure Headers, and CSRF
// should be registered first to ensure they wrap all subsequent
// handlers and middleware. Registering security middleware after
// application middleware may leave routes unprotected.
//
// In contrast with the [Router.With] method, this one does modify
// the current router instead of returning a new sub-router.
func (router *Router[H]) Use(middlewares ...Middleware[H]) {
	router.middlewares = append(router.middlewares, middlewares...)
}

// register adds the given pattern and handler to the actual native
// router [http.ServeMux].
func (router *Router[H]) register(method string, pattern string, handler H) {
	pattern = fmt.Sprintf("%s %s", method, pattern)
	router.mux().Handle(pattern, router.wrap(handler))
}

// registerRoot registers the given pattern when the route
// is supposed to be a root route ("/").
func (router *Router[H]) registerRoot(method string, handler H) {
	router.register(method, "/{$}", handler)
}

// registerTrailing registers the given pattern when the route
// is supposed to end up in a slash ("/").
func (router *Router[H]) registerTrailing(method string, pattern string, handler H) {
	pattern = fmt.Sprintf("%s/{$}", pattern)
	router.register(method, pattern, handler)
}

// registerPair registers both a pattern and its
// trailing-slash counterpart to ensure routes match with
// or without a trailing slash.
//
// Possible scenarios are:
//  1. The path ends in wildcard match all: '...':
//     - Must register the pattern
//     - Must register the same without wildcard and
//     slash: '/{abc...}'
//  2. The path does not end in '...':
//     - Must register the pattern
//     - Must register the pattern + '/'
//
// Note: patterns ending in '/' never reach this method
// because [path.Join] in [Router.Method] strips trailing
// slashes, and the bare "/" case is handled separately by
// [Router.registerRoot].
//
// Note that catch-all routes (e.g., "/files/{path...}")
// also match their base path (e.g., "/files/") because
// both the catch-all pattern and the base path without
// the wildcard segment are registered. Handlers should
// account for the catch-all value being empty when the
// base path is matched directly.
func (router *Router[H]) registerPair(method string, pattern string, handler H) {
	// In all cases we always register the pattern, so let's do this first.
	router.register(method, pattern, handler)

	// Check if we are in a wildcard "rest" parameter.
	if strings.HasSuffix(pattern, "...}") {
		// We have to register the pattern without
		// the left slash and the rest parameter, so
		// we can remove it and register it.
		if segments := strings.Split(pattern, "/"); len(segments) > 2 {
			pattern = strings.Join(segments[:len(segments)-1], "/")
			router.register(method, pattern, handler)
		}

		return
	}

	// Given the path does not end in a rest wildcard,
	// we can simply register the trailing slash pattern
	// to adhere to both.
	router.registerTrailing(method, pattern, handler)
}

// Method registers a new handler to the router with the
// given method and pattern. This is useful if you need to
// dynamically register a route to the router using a string
// as the method.
//
// A notable difference is that the pattern's ending slash
// "/" is not treated as an anonymous catch-all "{...}" and
// is instead treated as if it finished with "/{$}", making
// a specific route only.
//
// If the route does not finish in "/", one will be added
// automatically and then the paragraph above will apply
// unless the route finishes in a catch-all parameter "...}"
//
// WARNING: The pattern is joined with the router's base
// pattern using [path.Join], which normalizes the resulting
// path. This means ".." segments are resolved and collapsed.
// Callers should not rely on literal ".." segments in route
// patterns.
//
// Typically, the method string should be one of the
// following:
//   - [http.MethodGet]
//   - [http.MethodHead]
//   - [http.MethodPost]
//   - [http.MethodPut]
//   - [http.MethodPatch]
//   - [http.MethodDelete]
//   - [http.MethodOptions]
//
// WARNING: TRACE and CONNECT should not be used in general-purpose
// applications. TRACE enables cross-site tracing (XST) attacks and
// CONNECT is reserved for HTTP proxies. If needed, use the
// [Router.Trace] or [Router.Connect] methods explicitly.
func (router *Router[H]) Method(method string, pattern string, handler H) {
	if method == "" {
		panic("router: method must not be empty")
	}

	pattern = path.Join(router.pattern, pattern)

	// When the pattern is simply a slash, we shall
	// only register itself with an ending wildcard.
	if pattern == "/" {
		router.registerRoot(method, handler)

		return
	}

	router.registerPair(method, pattern, handler)
}

// Methods registers a handler for each method in the given slice by calling
// [Router.Method] for each entry.
func (router *Router[H]) Methods(methods []string, pattern string, handler H) {
	for _, method := range methods {
		router.Method(method, pattern, handler)
	}
}

// Any registers a handler for all standard HTTP methods (GET, HEAD, POST, PUT,
// PATCH, DELETE, OPTIONS) using [Router.Methods]. TRACE and CONNECT are
// intentionally excluded for security reasons.
func (router *Router[H]) Any(pattern string, handler H) {
	router.Methods(allMethods, pattern, handler)
}

// Get registers a handler for [http.MethodGet] using [Router.Method].
func (router *Router[H]) Get(pattern string, handler H) {
	router.Method(http.MethodGet, pattern, handler)
}

// Head registers a handler for [http.MethodHead] using [Router.Method].
func (router *Router[H]) Head(pattern string, handler H) {
	router.Method(http.MethodHead, pattern, handler)
}

// Post registers a handler for [http.MethodPost] using [Router.Method].
func (router *Router[H]) Post(pattern string, handler H) {
	router.Method(http.MethodPost, pattern, handler)
}

// Put registers a handler for [http.MethodPut] using [Router.Method].
func (router *Router[H]) Put(pattern string, handler H) {
	router.Method(http.MethodPut, pattern, handler)
}

// Patch registers a handler for [http.MethodPatch] using [Router.Method].
func (router *Router[H]) Patch(pattern string, handler H) {
	router.Method(http.MethodPatch, pattern, handler)
}

// Delete registers a handler for [http.MethodDelete] using [Router.Method].
func (router *Router[H]) Delete(pattern string, handler H) {
	router.Method(http.MethodDelete, pattern, handler)
}

// Connect registers a handler for [http.MethodConnect] using [Router.Method].
func (router *Router[H]) Connect(pattern string, handler H) {
	router.Method(http.MethodConnect, pattern, handler)
}

// Options registers a handler for [http.MethodOptions] using [Router.Method].
func (router *Router[H]) Options(pattern string, handler H) {
	router.Method(http.MethodOptions, pattern, handler)
}

// Trace registers a handler for [http.MethodTrace] using [Router.Method].
func (router *Router[H]) Trace(pattern string, handler H) {
	router.Method(http.MethodTrace, pattern, handler)
}

// ServeHTTP implements [http.Handler] by delegating to the underlying [http.ServeMux].
func (router *Router[H]) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	router.mux().ServeHTTP(w, r)
}

// Has reports whether the given pattern is registered in the router
// with the given method.
//
// Alternatively, use the [Router.Matches] method with an [http.Request].
func (router *Router[H]) Has(method string, pattern string) bool {
	if req, err := http.NewRequest(method, pattern, nil); err == nil {
		return router.Matches(req)
	}

	return false
}

// Matches reports whether the given [http.Request] matches any registered
// route in the router.
//
// This means that, given the request method and the
// URL, a handler can be resolved.
func (router *Router[H]) Matches(request *http.Request) bool {
	_, ok := router.HandlerMatch(request)

	return ok
}

// Handler returns the handler that matches the given method and pattern.
// The second return value reports whether the handler was found; when false,
// the first return value is the zero value of H.
//
// For matching against an [http.Request] use the [Router.HandlerMatch] method.
func (router *Router[H]) Handler(method string, pattern string) (h H, ok bool) {
	if req, err := http.NewRequest(method, pattern, nil); err == nil {
		return router.HandlerMatch(req)
	}

	return h, false
}

// HandlerMatch returns the handler that matches the given [http.Request].
// The second return value determines if the handler was found or not.
//
// For matching against a method and a pattern, use the [Router.Handler] method.
func (router *Router[H]) HandlerMatch(request *http.Request) (h H, ok bool) {
	// We can look for that specific handler in the
	// native [http.ServeMux] and return it if found.
	if handler, pattern := router.mux().Handler(request); pattern != "" {
		if typed, ok := handler.(H); ok {
			return typed, true
		}
	}

	return h, false
}

// Record returns an [http.Response] produced by dispatching the given HTTP
// request through the router's full middleware and handler pipeline.
func (router *Router[H]) Record(request *http.Request) *http.Response {
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, request)

	return rr.Result()
}
