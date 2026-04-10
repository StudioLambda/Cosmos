package request

import (
	"net/http"

	"github.com/studiolambda/cosmos/contract"
	"github.com/studiolambda/cosmos/problem"
)

// ErrNoHooksMiddleware is the problem returned when attempting to
// access hooks from a request that lacks the hooks middleware.
var ErrNoHooksMiddleware = problem.Problem{
	Title:  "No hooks context",
	Detail: "Unable to resolve hooks as there's no context value",
	Status: http.StatusInternalServerError,
}

// Hooks returns the hooks instance used to attach callbacks
// to lifecycle events. It panics if the hooks middleware has
// not been applied to the request's context.
func Hooks(r *http.Request) contract.Hooks {
	if hooks, ok := r.Context().Value(contract.HooksKey).(contract.Hooks); ok {
		return hooks
	}

	panic(ErrNoHooksMiddleware)
}
