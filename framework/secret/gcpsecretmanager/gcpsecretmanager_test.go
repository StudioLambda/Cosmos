package gcpsecretmanager

import (
	"context"
	"errors"
	"testing"

	secretmanagerpb "cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"github.com/googleapis/gax-go/v2"
	"github.com/stretchr/testify/require"
	"github.com/studiolambda/cosmos/contract"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type secretManagerClient struct {
	response *secretmanagerpb.AccessSecretVersionResponse
	err      error
}

func (client secretManagerClient) AccessSecretVersion(context.Context, *secretmanagerpb.AccessSecretVersionRequest, ...gax.CallOption) (*secretmanagerpb.AccessSecretVersionResponse, error) {
	return client.response, client.err
}

func TestClientGetReturnsPayload(t *testing.T) {
	t.Parallel()

	client := NewFrom(secretManagerClient{response: &secretmanagerpb.AccessSecretVersionResponse{
		Payload: &secretmanagerpb.SecretPayload{Data: []byte("secret")},
	}})

	contents, err := client.Get(context.Background(), "projects/example/secrets/database-dsn/versions/latest")

	require.NoError(t, err)
	require.Equal(t, []byte("secret"), contents)
}

func TestClientGetMapsNotFoundResponse(t *testing.T) {
	t.Parallel()

	client := NewFrom(secretManagerClient{err: status.Error(codes.NotFound, "missing")})

	_, err := client.Get(context.Background(), "projects/example/secrets/missing/versions/latest")

	require.ErrorIs(t, err, contract.ErrSecretNotFound)
}

func TestClientGetPreservesUnexpectedError(t *testing.T) {
	t.Parallel()

	expected := errors.New("unavailable")
	client := NewFrom(secretManagerClient{err: expected})

	_, err := client.Get(context.Background(), "projects/example/secrets/database-dsn/versions/latest")

	require.ErrorIs(t, err, expected)
}
