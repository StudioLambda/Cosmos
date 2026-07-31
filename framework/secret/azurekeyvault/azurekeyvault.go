package azurekeyvault

import (
	"context"
	"errors"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets"
	"github.com/studiolambda/cosmos/contract"
)

// Config configures an Azure Key Vault client.
type Config struct {
	VaultURL string
}

type client interface {
	GetSecret(ctx context.Context, name, version string, options *azsecrets.GetSecretOptions) (azsecrets.GetSecretResponse, error)
}

// Client retrieves secrets from Azure Key Vault.
type Client struct {
	client client
}

// New creates an Azure Key Vault client using Azure's default credential chain.
func New(ctx context.Context, clientConfig Config) (*Client, error) {
	credential, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, fmt.Errorf("create Azure credential: %w", err)
	}

	client, err := azsecrets.NewClient(clientConfig.VaultURL, credential, nil)
	if err != nil {
		return nil, fmt.Errorf("create Azure Key Vault client: %w", err)
	}

	return NewFrom(client), nil
}

// NewFrom creates a client from an existing Azure Key Vault client.
func NewFrom(client client) *Client {
	return &Client{client: client}
}

// Get retrieves the current value of name.
func (client *Client) Get(ctx context.Context, name string) ([]byte, error) {
	secret, err := client.client.GetSecret(ctx, name, "", nil)
	if err != nil {
		var responseError *azcore.ResponseError
		if errors.As(err, &responseError) && responseError.StatusCode == 404 {
			return nil, fmt.Errorf("%w: %s", contract.ErrSecretNotFound, name)
		}

		return nil, fmt.Errorf("get secret %q: %w", name, err)
	}

	if secret.Value == nil {
		return nil, fmt.Errorf("secret %q has no value", name)
	}

	return []byte(*secret.Value), nil
}
