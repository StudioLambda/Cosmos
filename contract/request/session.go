package request

import (
	"net/http"

	"github.com/studiolambda/cosmos/contract"
	"github.com/studiolambda/cosmos/problem"
)

// ErrSessionNotFound is the problem returned when attempting to
// retrieve a session that does not exist in the request context.
var ErrSessionNotFound = problem.Problem{
	Title:  "Session not found",
	Detail: "Unable to find the session in the request",
	Status: http.StatusInternalServerError,
}

// Session retrieves the session from the request context using the
// default [contract.SessionKey]. The boolean return value indicates
// whether a session was found.
func Session(r *http.Request) (contract.Session, bool) {
	return SessionKeyed(r, contract.SessionKey)
}

// SessionKeyed retrieves the session from the request context using
// the provided key. The boolean return value indicates whether a
// session was found for that key.
func SessionKeyed(r *http.Request, key any) (contract.Session, bool) {
	s, ok := r.Context().Value(key).(contract.Session)

	return s, ok
}

// MustSessionKeyed retrieves the session from the request context
// using the provided key. It panics with [ErrSessionNotFound] if
// no session is found.
//
// WARNING: This function panics when the session is missing. Use
// [SessionKeyed] for a non-panicking alternative, or ensure the
// [framework.Recover] middleware is in place.
func MustSessionKeyed(r *http.Request, key any) contract.Session {
	if s, ok := SessionKeyed(r, key); ok {
		return s
	}

	panic(ErrSessionNotFound)
}

// MustSession retrieves the session from the request context using
// the default [contract.SessionKey]. It panics with [ErrSessionNotFound]
// if no session is found.
//
// WARNING: This function panics when the session is missing. Use
// [Session] for a non-panicking alternative, or ensure the
// [framework.Recover] middleware is in place.
func MustSession(r *http.Request) contract.Session {
	return MustSessionKeyed(r, contract.SessionKey)
}
