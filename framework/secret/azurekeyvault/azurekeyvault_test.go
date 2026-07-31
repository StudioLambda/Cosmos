package azurekeyvault

import (
	"context"
	"errors"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets"
	"github.com/stretchr/testify/require"
	"github.com/studiolambda/cosmos/contract"
)

type keyVaultClient struct {
	response azsecrets.GetSecretResponse
	err      error
}

func (client keyVaultClient) GetSecret(context.Context, string, string, *azsecrets.GetSecretOptions) (azsecrets.GetSecretResponse, error) {
	return client.response, client.err
}

func TestClientGetReturnsValue(t *testing.T) {
	t.Parallel()

	value := "secret"
	client := NewFrom(keyVaultClient{response: azsecrets.GetSecretResponse{
		Secret: azsecrets.Secret{Value: &value},
	}})

	contents, err := client.Get(context.Background(), "database-dsn")

	require.NoError(t, err)
	require.Equal(t, []byte("secret"), contents)
}

func TestClientGetMapsNotFoundResponse(t *testing.T) {
	t.Parallel()

	client := NewFrom(keyVaultClient{err: &azcore.ResponseError{StatusCode: 404}})

	_, err := client.Get(context.Background(), "missing")

	require.ErrorIs(t, err, contract.ErrSecretNotFound)
}

func TestClientGetPreservesUnexpectedError(t *testing.T) {
	t.Parallel()

	expected := errors.New("unavailable")
	client := NewFrom(keyVaultClient{err: expected})

	_, err := client.Get(context.Background(), "database-dsn")

	require.ErrorIs(t, err, expected)
}
