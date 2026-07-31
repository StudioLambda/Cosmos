package oidc

import (
	"context"
	"fmt"
	"net/http"

	coreoidc "github.com/coreos/go-oidc/v3/oidc"
)

type identityKey struct{}

// Identity represents an authenticated OIDC subject.
type Identity struct {
	Subject string         `json:"subject"`
	Issuer  string         `json:"issuer"`
	Email   string         `json:"email"`
	Name    string         `json:"name"`
	Claims  map[string]any `json:"claims"`
}

// From returns the authenticated identity attached to request.
func From(request *http.Request) (Identity, bool) {
	identity, ok := request.Context().Value(identityKey{}).(Identity)

	return identity, ok
}

func identityFromToken(token *coreoidc.IDToken) (Identity, error) {
	claims := make(map[string]any)
	if err := token.Claims(&claims); err != nil {
		return Identity{}, fmt.Errorf("decode OIDC claims: %w", err)
	}

	identity := Identity{
		Subject: token.Subject,
		Issuer:  token.Issuer,
		Claims:  claims,
	}

	if email, ok := claims["email"].(string); ok {
		identity.Email = email
	}

	if name, ok := claims["name"].(string); ok {
		identity.Name = name
	}

	return identity, nil
}

func withIdentity(request *http.Request, identity Identity) *http.Request {
	return request.WithContext(context.WithValue(request.Context(), identityKey{}, identity))
}
