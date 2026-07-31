package vault

import (
	"context"
	"errors"
	"testing"

	"github.com/hashicorp/vault/api"
	"github.com/stretchr/testify/require"
	"github.com/studiolambda/cosmos/contract"
)

type vaultClient struct {
	secret *api.KVSecret
	err    error
}

func (client vaultClient) Get(context.Context, string) (*api.KVSecret, error) {
	return client.secret, client.err
}

func TestClientGetReturnsJSONData(t *testing.T) {
	t.Parallel()

	client := NewFrom(vaultClient{secret: &api.KVSecret{Data: map[string]any{
		"database": map[string]any{"dsn": "postgres://example"},
	}}})

	contents, err := client.Get(context.Background(), "api/config")

	require.NoError(t, err)
	require.JSONEq(t, `{"database":{"dsn":"postgres://example"}}`, string(contents))
}

func TestClientGetMapsMissingSecret(t *testing.T) {
	t.Parallel()

	client := NewFrom(vaultClient{})

	_, err := client.Get(context.Background(), "missing")

	require.ErrorIs(t, err, contract.ErrSecretNotFound)
}

func TestClientGetMapsNotFoundResponse(t *testing.T) {
	t.Parallel()

	client := NewFrom(vaultClient{err: &api.ResponseError{StatusCode: 404}})

	_, err := client.Get(context.Background(), "missing")

	require.ErrorIs(t, err, contract.ErrSecretNotFound)
}

func TestClientGetPreservesUnexpectedError(t *testing.T) {
	t.Parallel()

	expected := errors.New("unavailable")
	client := NewFrom(vaultClient{err: expected})

	_, err := client.Get(context.Background(), "api/config")

	require.ErrorIs(t, err, expected)
}
