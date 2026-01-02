package request

import (
	"net/http"

	"github.com/studiolambda/cosmos/contract"
	"github.com/studiolambda/cosmos/framework"
	"github.com/studiolambda/cosmos/problem"
)

var ErrNoHooksMiddleware = problem.Problem{
	Title:  "No hooks context",
	Detail: "Unable to resolve hooks as there's no context value",
	Status: http.StatusInternalServerError,
}

// Hooks return the hooks struct to attach callbacks
// to certain lifecycle events. It panics if no context
// value is found. Make sure you use the hooks middleware
// before using this method.
func Hooks(r *http.Request) *framework.Hooks {
	if hooks, ok := r.Context().Value(contract.HooksKey).(*framework.Hooks); ok {
		return hooks
	}

	panic(ErrNoHooksMiddleware)
}
