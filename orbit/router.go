package orbit

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"path"
	"slices"
	"strings"
)

// Middleware is a func type that can be used to
// apply middleware logic between request and responses.
type Middleware[H http.Handler] = func(H) H

// Router is the structure that handles
// http routing in an application.
//
// This router is completly optional and
// uses [http.ServeMux] under the hood
// to register all the routes.
//
// It also handles some patterns automatically,
// such as {$}, that is appended on each route
// automatically, regardless of the pattern.
type Router[H http.Handler] struct {

	// native stores the actual [http.ServeMux]
	// that's used internally  to register the routes.
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

// allMethods stores all the http methods
// available to use. This is interesting when
// dealing with the [Router.Any] method.
var allMethods = []string{
	http.MethodGet,
	http.MethodHead,
	http.MethodPost,
	http.MethodPut,
	http.MethodPatch,
	http.MethodDelete,
	http.MethodConnect,
	http.MethodOptions,
	http.MethodTrace,
}

// NewRouter creates a new [Router] instance and
// automatically creates all the needed components
// such as the middleware list or the native
// [http.ServeMux] that's used under the hood.
func New[H http.Handler]() *Router[H] {
	return &Router[H]{
		native:      http.NewServeMux(),
		pattern:     "",
		parent:      nil,
		middlewares: make([]Middleware[H], 0),
	}
}

// Group uses the given pattern to automatically
// mount a sub-router that has that pattern as a
// prefix.
//
// This means that any route registered with the
// sub-router will also have the given pattern suffixed.
//
// Keep in mind this can be nested as well, meaning that
// many sub-routers may be grouped, creating complex patterns.
func (router *Router[H]) Group(pattern string, subrouter func(*Router[H])) {
	subrouter(&Router[H]{
		native:      nil, // parent's native will be used
		pattern:     path.Join(router.pattern, pattern),
		parent:      router,
		middlewares: slices.Clone(router.middlewares),
	})
}

// With does create a new sub-router that automatically applies
// the given middlewares.
//
// This is very usefull when used to inline some middlewares to
// specific routes.
//
// In constrast to [Router.Use] method, it does create a new
// sub-router instead of modifying the current router.
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
// some recursivity involved to get the same [http.ServeMux].
func (router *Router[H]) mux() *http.ServeMux {
	if router.parent != nil {
		return router.parent.mux()
	}

	return router.native
}

// wrap makes an handler wrapped by the current routers'
// middlewares. This means that the resulting handler is
// the same as first calling the router middlewares and then the
// provided handler.
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
// In constrats with the [Router.With] method, this one does modify
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
// is suposed to be a root route ("/").
func (router *Router[H]) registerRoot(method string, handler H) {
	router.register(method, "/{$}", handler)
}

// registerTrailing registers the given pattern when the route
// is suposed to end up in a slash ("/").
func (router *Router[H]) registerTrailing(method string, pattern string, handler H) {
	pattern = fmt.Sprintf("%s/{$}", pattern)
	router.register(method, pattern, handler)
}

// There's two possible scenarios here, either the route
// in question is a trailing slash or not. We must make sure
// to register both either way.
//
// Possible scenarios are:
//  1. The path ends in anonymous wildcard: '/'
//     - Must register the pattern
//     - Must register the same without the last '/' (sometimes)
//  2. The path ends in wildcard match all: '...':
//     - Must register the pattern
//     - Must register the same without wildcard and slash: '/{abc...}'
//  3. The path does not end in '/' nor '...':
//     - Must register the pattern
//     - Must register the pattern + '/'
func (router *Router[H]) registerPair(method string, pattern string, handler H) {
	// In all cases we always register the pattern, so let's do this first.
	router.register(method, pattern, handler)

	// Assest the first scenario, check if we
	// are in a pattern ended with a slash
	if strings.HasSuffix(pattern, "/") {
		// We must register the pattern without the last
		// slash, so we simply remove it:
		pattern = strings.TrimSuffix(pattern, "/")
		router.register(method, pattern, handler)

		return
	}

	// Assest the second scenario, check if we
	// are in a wildcard "rest" parameter.
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

	// Given the path does not end in a slash nor a
	// rest wildcard, we can simply register the trailing
	// slash pattern to adhere to both.
	router.registerTrailing(method, pattern, handler)
}

// Method registers a new handler to the router with the given
// method and pattern. This is usefull if you need to dynamically
// register a route to the router using a string as the method.
//
// A notable difference is that the patterns's ending slash "/" is
// not treated as an annonymous catch-all "{...}" and is instead treated
// as if it finished with "/{$}", making a specific route only.
//
// If the route does not finish in "/", one will be added automatically and
// then the paragraph above will apply unless the route finishes in a catch-all
// parameter "...}"
//
// Typically, the method string should be one of the following:
//   - [http.MethodGet]
//   - [http.MethodHead]
//   - [http.MethodPost]
//   - [http.MethodPut]
//   - [http.MethodPatch]
//   - [http.MethodDelete]
//   - [http.MethodConnect]
//   - [http.MethodOptions]
//   - [http.MethodTrace]
func (router *Router[H]) Method(method string, pattern string, handler H) {
	pattern = path.Join(router.pattern, pattern)

	// When the pattern is simply a slash, we shall
	// only register itself with an ending wildcard.
	if pattern == "/" {
		router.registerRoot(method, handler)

		return
	}

	router.registerPair(method, pattern, handler)
}

// MethodFunc is a simple alias of [Router.Method] to register
// a [TFunc] without the manual typecast.
func (router *Router[H]) MethodFunc(method string, pattern string, handler H) {
	router.Method(method, pattern, handler)
}

// Methods allows binding multiple methods to the pattern and handler.
func (router *Router[H]) Methods(methods []string, pattern string, handler H) {
	for _, method := range methods {
		router.Method(method, pattern, handler)
	}
}

// MethodsFunc is a simple alias of [Router.Methods] to register
// a [TFunc] without the manual typecast.
func (router *Router[H]) MethodsFunc(methods []string, pattern string, handler H) {
	router.Methods(methods, pattern, handler)
}

// Any registers all methods to the given pattern and handler.
func (router *Router[H]) Any(pattern string, handler H) {
	router.Methods(allMethods, pattern, handler)
}

// Get registers a new handler to the router using [Router.Method]
// and using the [http.MethodGet] as the method parameter.
func (router *Router[H]) Get(pattern string, handler H) {
	router.Method(http.MethodGet, pattern, handler)
}

// Head registers a new handler to the router using [Router.Method]
// and using the [http.MethodHead] as the method parameter.
func (router *Router[H]) Head(pattern string, handler H) {
	router.Method(http.MethodHead, pattern, handler)
}

// Post registers a new handler to the router using [Router.Method]
// and using the [http.MethodPost] as the method parameter.
func (router *Router[H]) Post(pattern string, handler H) {
	router.Method(http.MethodPost, pattern, handler)
}

// Put registers a new handler to the router using [Router.Method]
// and using the [http.MethodPut] as the method parameter.
func (router *Router[H]) Put(pattern string, handler H) {
	router.Method(http.MethodPut, pattern, handler)
}

// Patch registers a new handler to the router using [Router.Method]
// and using the [http.MethodPatch] as the method parameter.
func (router *Router[H]) Patch(pattern string, handler H) {
	router.Method(http.MethodPatch, pattern, handler)
}

// Delete registers a new handler to the router using [Router.Method]
// and using the [http.MethodDelete] as the method parameter.
func (router *Router[H]) Delete(pattern string, handler H) {
	router.Method(http.MethodDelete, pattern, handler)
}

// Connect registers a new handler to the router using [Router.Method]
// and using the [http.MethodConnect] as the method parameter.
func (router *Router[H]) Connect(pattern string, handler H) {
	router.Method(http.MethodConnect, pattern, handler)
}

// Options registers a new handler to the router using [Router.Method]
// and using the [http.MethodOptions] as the method parameter.
func (router *Router[H]) Options(pattern string, handler H) {
	router.Method(http.MethodOptions, pattern, handler)
}

// Trace registers a new handler to the router using [Router.Method]
// and using the [http.MethodTrace] as the method parameter.
func (router *Router[H]) Trace(pattern string, handler H) {
	router.Method(http.MethodTrace, pattern, handler)
}

// ServeHTTP is the method that will make the router implement
// the handler interface, making it possible to be used
// as a handler in places like [http.Server].
func (router *Router[H]) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	router.native.ServeHTTP(w, r)
}

// Has reports whether the given pattern is registered in the router
// with the given method.
//
// Alternatively, check out the [Router.Matches] to use an [http.Request]
// as the parameter.
func (router *Router[H]) Has(method string, pattern string) bool {
	if req, err := http.NewRequest(method, pattern, nil); err == nil {
		return router.Matches(req)
	}

	return false
}

// Matches reports whether the given [http.Request] match any registered
// route in the router.
//
// This means that, given the request method and the
// URL, a handler can be resolved.
func (router *Router[H]) Matches(request *http.Request) bool {
	_, ok := router.HandlerMatch(request)

	return ok
}

// Handler returns the handler that matches the given method and pattern.
// The second return value determines if the handler was found or not.
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
	if handler, pattern := router.native.Handler(request); pattern != "" {
		return handler.(H), true
	}

	return h, false
}

// Record returns a [httptest.ResponseRecorder] that can be used to inspect what
// the given http request would have returned as a response.
//
// This method is a shortcut of calling [RecordHandler] with the router as the
// handler and the given request.
func (router *Router[H]) Record(r *http.Request) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, r)

	return recorder
}
