package awssm

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager/types"
	"github.com/stretchr/testify/require"
	"github.com/studiolambda/cosmos/contract"
)

type secretsManagerClient struct {
	output *secretsmanager.GetSecretValueOutput
	err    error
}

func (client secretsManagerClient) GetSecretValue(context.Context, *secretsmanager.GetSecretValueInput, ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error) {
	return client.output, client.err
}

func TestClientGetReturnsStringSecret(t *testing.T) {
	t.Parallel()

	client := NewFrom(secretsManagerClient{
		output: &secretsmanager.GetSecretValueOutput{SecretString: aws.String("value")},
	})

	value, err := client.Get(context.Background(), "example")

	require.NoError(t, err)
	require.Equal(t, []byte("value"), value)
}

func TestClientGetReturnsBinarySecret(t *testing.T) {
	t.Parallel()

	client := NewFrom(secretsManagerClient{
		output: &secretsmanager.GetSecretValueOutput{SecretBinary: []byte{1, 2, 3}},
	})

	value, err := client.Get(context.Background(), "example")

	require.NoError(t, err)
	require.Equal(t, []byte{1, 2, 3}, value)
}

func TestClientGetMapsNotFoundError(t *testing.T) {
	t.Parallel()

	client := NewFrom(secretsManagerClient{
		err: &types.ResourceNotFoundException{},
	})

	_, err := client.Get(context.Background(), "missing")

	require.ErrorIs(t, err, contract.ErrSecretNotFound)
}

func TestClientGetPreservesUnexpectedError(t *testing.T) {
	t.Parallel()

	expected := errors.New("unavailable")
	client := NewFrom(secretsManagerClient{err: expected})

	_, err := client.Get(context.Background(), "example")

	require.ErrorIs(t, err, expected)
}
