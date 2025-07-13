package framework

import (
	"github.com/studiolambda/cosmos/router"
)

// Router is the http router type that is
// used by an application.
//
// It is an alias for a orbit router.
type Router = router.Router[Handler]

// Middleware determines how a http middleware
// should be defined.
//
// It is an alias for a orbit middleware.
type Middleware = router.Middleware[Handler]

// New creates a new application.
//
// This can be passed directly as an http handler.
// Usually, this will be the handler of an [http.Server].
func New() *Router {
	return router.New[Handler]()
}
