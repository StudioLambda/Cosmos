package request

import (
	"net/http"

	"github.com/studiolambda/cosmos/contract"
	"github.com/studiolambda/cosmos/problem"
)

var ErrSessionNotFound = problem.Problem{
	Title:  "Session not found",
	Detail: "Unable to find the session in the request",
	Status: http.StatusInternalServerError,
}

func Session(r *http.Request, key any) (contract.Session, bool) {
	s, ok := r.Context().Value(key).(contract.Session)

	return s, ok
}

func MustSession(r *http.Request, key any) contract.Session {
	if s, ok := Session(r, key); ok {
		return s
	}

	panic(ErrSessionNotFound)
}
