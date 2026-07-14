// Package router provides a generic HTTP router built on net/http semantics.
//
// Router instances are parameterized by handler type and compose middleware,
// route groups, and method-specific registration while preserving compatibility
// with standard HTTP serving.
//
// # Routing model
//
// Use [New] to create a root router and register handlers with Get, Query,
// Post, Put, Patch, Delete, and related helpers. Groups inherit middleware and
// prefixes.
//
// # Thread safety
//
// Configure routes and middleware during startup. Request serving is
// concurrent. Avoid mutating router structure after the server begins handling
// traffic.
//
// Example
//
//	r := router.New[http.Handler]()
//	r.Use(logging)
//	r.Get("/users/{id}", http.HandlerFunc(showUser))
//	r.Group("/admin", func(admin *router.Router[http.Handler]) {
//		admin.Use(requireAdmin)
//		admin.Get("/stats", http.HandlerFunc(stats))
//	})
package router
