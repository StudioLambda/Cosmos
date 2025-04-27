package nova

import (
	"github.com/studiolambda/cosmos/orbit"
)

// Router is the http router type that is
// used by a nova application.
//
// It is an alias for a orbit router.
type Router = orbit.Router[Handler]

// Middleware determines how a http middleware
// should be defined.
//
// It is an alias for a orbit middleware.
type Middleware = orbit.Middleware[Handler]

// New creates a new nova application router.
//
// This can be passed directly as an http handler.
// Usually, this will be the handler of an [http.Server].
func New() *Router {
	return orbit.New[Handler]()
}
