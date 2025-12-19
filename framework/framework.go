package framework

import (
	_ "github.com/studiolambda/cosmos/contract"
	_ "github.com/studiolambda/cosmos/problem"
	"github.com/studiolambda/cosmos/router"
)

// Router is the HTTP router type used by Cosmos applications.
// It provides routing functionality with support for path parameters,
// middleware, and handler composition. The router uses generics to
// work with Cosmos-specific Handler types.
//
// This is an alias for router.Router[Handler] from the router package.
type Router = router.Router[Handler]

// Middleware defines the signature for HTTP middleware functions in Cosmos.
// Middleware can intercept requests before they reach handlers, modify
// requests/responses, handle authentication, logging, CORS, and other
// cross-cutting concerns.
//
// This is an alias for router.Middleware[Handler] from the router package.
type Middleware = router.Middleware[Handler]

// New creates a new Cosmos application router instance.
// The returned router implements http.Handler and can be used directly
// with http.Server or other HTTP server implementations.
//
// The router supports:
//   - RESTful routing with HTTP methods (GET, POST, PUT, DELETE, etc.)
//   - Path parameters and wildcards
//   - Middleware composition
//   - Route groups for organizing related endpoints
//
// Example usage:
//
//	app := framework.New()
//	app.Get("/users/{id}", getUserHandler)
//	http.ListenAndServe(":8080", app)
func New() *Router {
	return router.New[Handler]()
}
