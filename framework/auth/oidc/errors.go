package oidc

import (
	"net/http"

	"github.com/studiolambda/cosmos/problem"
)

// ErrUnauthenticated is returned when a request has no valid identity.
var ErrUnauthenticated = problem.Problem{
	Title:  "Unauthenticated",
	Detail: "Authentication is required to access this resource.",
	Status: http.StatusUnauthorized,
}

// ErrUnauthorized is returned when an authenticated identity lacks permission.
var ErrUnauthorized = problem.Problem{
	Title:  "Unauthorized",
	Detail: "The authenticated identity is not authorized to access this resource.",
	Status: http.StatusForbidden,
}

// ErrInvalidLoginRequest is returned when a login request contains an unsafe
// or malformed local redirect target.
var ErrInvalidLoginRequest = problem.Problem{
	Title:  "Invalid login request",
	Detail: "The login request could not be processed.",
	Status: http.StatusBadRequest,
}

// ErrInvalidCallback is returned when an OIDC callback cannot be verified.
var ErrInvalidCallback = problem.Problem{
	Title:  "Invalid authentication callback",
	Detail: "The authentication callback could not be verified.",
	Status: http.StatusBadRequest,
}

// ErrProviderUnavailable is returned when the OIDC provider cannot complete
// an authentication operation.
var ErrProviderUnavailable = problem.Problem{
	Title:  "Authentication provider unavailable",
	Detail: "The authentication provider could not complete the request.",
	Status: http.StatusBadGateway,
}

func invalidCallback(err error) problem.Problem {
	problem := ErrInvalidCallback.With("reason", "invalid_callback")
	if err == nil {
		return problem
	}

	return problem.WithError(err)
}

func invalidLoginRequest(err error) problem.Problem {
	problem := ErrInvalidLoginRequest.With("reason", "invalid_login_request")
	if err == nil {
		return problem
	}

	return problem.WithError(err)
}

func providerUnavailable(err error) problem.Problem {
	return ErrProviderUnavailable.WithError(err)
}
