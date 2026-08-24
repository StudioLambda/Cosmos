package oidc

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/studiolambda/cosmos/framework"
)

func TestFromReturnsAttachedIdentity(t *testing.T) {
	t.Parallel()

	request := withIdentity(httptest.NewRequest(http.MethodGet, "/", nil), Identity{Subject: "user-1"})
	identity, ok := From(request)

	require.True(t, ok)
	require.Equal(t, "user-1", identity.Subject)
}

func TestRequireScopesAllowsIdentityWithRequiredScopes(t *testing.T) {
	t.Parallel()

	client := &Client{}
	handler := client.RequireScopes("read", "write")(func(w http.ResponseWriter, r *http.Request) error {
		w.WriteHeader(http.StatusNoContent)

		return nil
	})
	request := withIdentity(httptest.NewRequest(http.MethodGet, "/", nil), Identity{
		Claims: map[string]any{"scope": "read write"},
	})

	response := handler.Record(request)

	require.Equal(t, http.StatusNoContent, response.StatusCode)
}

func TestRequireScopesRejectsMissingScope(t *testing.T) {
	t.Parallel()

	client := &Client{}
	handler := framework.Handler(client.RequireScopes("write")(func(http.ResponseWriter, *http.Request) error {
		return nil
	}))
	request := withIdentity(httptest.NewRequest(http.MethodGet, "/", nil), Identity{
		Claims: map[string]any{"scope": "read"},
	})

	response := handler.Record(request)

	require.Equal(t, http.StatusForbidden, response.StatusCode)
}

func TestRandomReturnsUniqueValues(t *testing.T) {
	t.Parallel()

	first, err := random()
	require.NoError(t, err)
	second, err := random()

	require.NoError(t, err)
	require.NotEmpty(t, first)
	require.NotEqual(t, first, second)
}
