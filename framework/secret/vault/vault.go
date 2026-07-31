package vault

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/vault/api"
	"github.com/studiolambda/cosmos/contract"

	"encoding/json/v2"
)

// Config configures a HashiCorp Vault KV v2 client.
type Config struct {
	Address string
	Token   string
	Mount   string
}

type client interface {
	Get(ctx context.Context, secretPath string) (*api.KVSecret, error)
}

// Client retrieves secrets from HashiCorp Vault KV v2.
type Client struct {
	client client
}

// New creates a HashiCorp Vault KV v2 client.
func New(clientConfig Config) (*Client, error) {
	config := api.DefaultConfig()
	if clientConfig.Address != "" {
		config.Address = clientConfig.Address
	}

	client, err := api.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("create Vault client: %w", err)
	}

	if clientConfig.Token != "" {
		client.SetToken(clientConfig.Token)
	}

	mount := clientConfig.Mount
	if mount == "" {
		mount = "secret"
	}

	return NewFrom(client.KVv2(mount)), nil
}

// NewFrom creates a client from an existing Vault KV v2 client.
func NewFrom(client client) *Client {
	return &Client{client: client}
}

// Get retrieves a Vault KV v2 secret as a JSON object.
func (client *Client) Get(ctx context.Context, name string) ([]byte, error) {
	secret, err := client.client.Get(ctx, name)
	if err != nil {
		var responseError *api.ResponseError
		if errors.As(err, &responseError) && responseError.StatusCode == 404 {
			return nil, fmt.Errorf("%w: %s", contract.ErrSecretNotFound, name)
		}

		return nil, fmt.Errorf("get secret %q: %w", name, err)
	}

	if secret == nil {
		return nil, fmt.Errorf("%w: %s", contract.ErrSecretNotFound, name)
	}

	contents, err := json.Marshal(secret.Data)
	if err != nil {
		return nil, fmt.Errorf("marshal secret %q: %w", name, err)
	}

	return contents, nil
}
