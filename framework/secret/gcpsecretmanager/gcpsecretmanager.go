// Package gcpsecretmanager provides a Google Cloud Secret Manager [contract.SecretDriver].
package gcpsecretmanager

import (
	"context"
	"errors"
	"fmt"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	secretmanagerpb "cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"github.com/googleapis/gax-go/v2"
	"github.com/studiolambda/cosmos/contract"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Config configures a Google Cloud Secret Manager client.
type Config struct{}

type client interface {
	AccessSecretVersion(ctx context.Context, request *secretmanagerpb.AccessSecretVersionRequest, options ...gax.CallOption) (*secretmanagerpb.AccessSecretVersionResponse, error)
}

// Client retrieves secrets from Google Cloud Secret Manager.
type Client struct {
	client client
}

// New creates a Google Cloud Secret Manager client using application default credentials.
func New(ctx context.Context, _ Config) (*Client, error) {
	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("create Google Cloud Secret Manager client: %w", err)
	}

	return NewFrom(client), nil
}

// NewFrom creates a client from an existing Google Cloud Secret Manager client.
func NewFrom(client client) *Client {
	return &Client{client: client}
}

// Get retrieves the latest value of a fully-qualified secret version name.
func (client *Client) Get(ctx context.Context, name string) ([]byte, error) {
	response, err := client.client.AccessSecretVersion(ctx, &secretmanagerpb.AccessSecretVersionRequest{
		Name: name,
	})
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, fmt.Errorf("%w: %s", contract.ErrSecretNotFound, name)
		}

		return nil, fmt.Errorf("get secret %q: %w", name, err)
	}

	if response.Payload == nil {
		return nil, errors.New("secret response has no payload")
	}

	return response.Payload.Data, nil
}
