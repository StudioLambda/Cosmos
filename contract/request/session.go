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

func Session(r *http.Request) (contract.Session, bool) {
	return SessionKeyed(r, contract.SessionKey)
}

func SessionKeyed(r *http.Request, key any) (contract.Session, bool) {
	s, ok := r.Context().Value(key).(contract.Session)

	return s, ok
}

func MustSessionKeyed(r *http.Request, key any) contract.Session {
	if s, ok := SessionKeyed(r, key); ok {
		return s
	}

	panic(ErrSessionNotFound)
}

func MustSession(r *http.Request) contract.Session {
	return MustSessionKeyed(r, contract.SessionKey)
}
