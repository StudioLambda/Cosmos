// Package awssm provides an AWS Secrets Manager [contract.SecretDriver].
package awssm

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager/types"
	"github.com/studiolambda/cosmos/contract"
)

// Config configures an AWS Secrets Manager client.
type Config struct {
	Region   string
	Endpoint string
	Name     string
}

type client interface {
	GetSecretValue(ctx context.Context, params *secretsmanager.GetSecretValueInput, options ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error)
}

// Client retrieves secrets from AWS Secrets Manager.
type Client struct {
	client client
}

// New creates an AWS Secrets Manager client using the AWS default credential chain.
func New(ctx context.Context, clientConfig Config) (*Client, error) {
	options := []func(*config.LoadOptions) error{}

	if clientConfig.Region != "" {
		options = append(options, config.WithRegion(clientConfig.Region))
	}

	if clientConfig.Endpoint != "" {
		options = append(options, config.WithBaseEndpoint(clientConfig.Endpoint))
	}

	awsConfig, err := config.LoadDefaultConfig(ctx, options...)
	if err != nil {
		return nil, fmt.Errorf("load AWS configuration: %w", err)
	}

	return NewFrom(secretsmanager.NewFromConfig(awsConfig)), nil
}

// NewFrom creates a client from an existing AWS Secrets Manager client.
func NewFrom(client client) *Client {
	return &Client{client: client}
}

// Get retrieves the raw secret value for name.
func (client *Client) Get(ctx context.Context, name string) ([]byte, error) {
	output, err := client.client.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(name),
	})

	if _, ok := errors.AsType[*types.ResourceNotFoundException](err); ok {
		return nil, fmt.Errorf("%w: %s", contract.ErrSecretNotFound, name)
	}

	if err != nil {
		return nil, fmt.Errorf("get secret %q: %w", name, err)
	}

	if output.SecretString != nil {
		return []byte(*output.SecretString), nil
	}

	return output.SecretBinary, nil
}
