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
//
// WARNING: This function panics when hooks are missing. Use
// [TryHooks] for a non-panicking alternative, or ensure the
// [framework.Recover] middleware is in place.
func Hooks(r *http.Request) contract.Hooks {
	if hooks, ok := r.Context().Value(contract.HooksKey).(contract.Hooks); ok {
		return hooks
	}

	panic(ErrNoHooksMiddleware)
}

// TryHooks retrieves the hooks instance from the request
// context without panicking. The boolean return value indicates
// whether hooks were found. This is the safe alternative to
// [Hooks] for use outside the framework handler chain.
func TryHooks(r *http.Request) (contract.Hooks, bool) {
	hooks, ok := r.Context().Value(contract.HooksKey).(contract.Hooks)

	return hooks, ok
}
