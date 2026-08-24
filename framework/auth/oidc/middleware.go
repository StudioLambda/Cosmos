package oidc

import (
	"fmt"
	"net/http"
	"strings"

	requestpkg "github.com/studiolambda/cosmos/contract/request"
	"github.com/studiolambda/cosmos/framework"
)

// RequireAuthentication requires a session identity or a valid JWT bearer token.
func (client *Client) RequireAuthentication() framework.Middleware {
	return func(next framework.Handler) framework.Handler {
		return func(w http.ResponseWriter, request *http.Request) error {
			identity, ok := sessionIdentity(request)
			if !ok {
				var err error
				identity, err = client.bearerIdentity(request)
				if err != nil {
					w.Header().Set("WWW-Authenticate", "Bearer")

					return ErrUnauthenticated.With("reason", "invalid_bearer_token").WithError(err)
				}
			}

			return next(w, withIdentity(request, identity))
		}
	}
}

// RequireScopes requires all scopes on an authenticated identity.
func (client *Client) RequireScopes(required ...string) framework.Middleware {
	return func(next framework.Handler) framework.Handler {
		return func(w http.ResponseWriter, request *http.Request) error {
			identity, ok := From(request)
			if !ok {
				return ErrUnauthenticated
			}

			available := scopes(identity.Claims["scope"])
			for _, scope := range required {
				if _, ok := available[scope]; !ok {
					return ErrUnauthorized
				}
			}

			return next(w, request)
		}
	}
}

func sessionIdentity(request *http.Request) (Identity, bool) {
	session, ok := requestpkg.Session(request)
	if !ok {
		return Identity{}, false
	}

	identity, err := session.Get[Identity](identitySessionKey)

	return identity, err == nil
}

func (client *Client) bearerIdentity(request *http.Request) (Identity, error) {
	values := request.Header.Values("Authorization")
	if len(values) != 1 {
		return Identity{}, fmt.Errorf("expected one authorization header")
	}

	scheme, raw, found := strings.Cut(values[0], " ")
	if !found || !strings.EqualFold(scheme, "Bearer") || raw == "" || strings.ContainsAny(raw, " \t") {
		return Identity{}, fmt.Errorf("invalid bearer authorization header")
	}

	token, err := client.verify(request.Context(), raw)
	if err != nil {
		return Identity{}, err
	}

	return identityFromToken(token)
}
