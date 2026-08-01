package oidc

import (
	"context"
	"fmt"
	"slices"
	"strings"

	coreoidc "github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

// Client authenticates requests with one OIDC provider.
type Client struct {
	provider *coreoidc.Provider
	verifier *coreoidc.IDTokenVerifier
	oauth    oauth2.Config
	config   Config
}

// New discovers an OIDC provider and creates a client. Empty scopes default to
// openid; empty accepted audiences default to the configured client ID.
func New(ctx context.Context, config Config) (*Client, error) {
	if config.IssuerURL == "" {
		return nil, fmt.Errorf("OIDC issuer URL cannot be empty")
	}

	if config.ClientID == "" {
		return nil, fmt.Errorf("OIDC client ID cannot be empty")
	}

	provider, err := coreoidc.NewProvider(ctx, config.IssuerURL)
	if err != nil {
		return nil, fmt.Errorf("discover OIDC provider: %w", err)
	}

	if len(config.Scopes) == 0 {
		config.Scopes = []string{coreoidc.ScopeOpenID}
	}

	return &Client{
		provider: provider,
		verifier: provider.Verifier(&coreoidc.Config{SkipClientIDCheck: true}),
		oauth: oauth2.Config{
			ClientID:     config.ClientID,
			ClientSecret: config.ClientSecret,
			RedirectURL:  config.RedirectURL,
			Endpoint:     provider.Endpoint(),
			Scopes:       config.Scopes,
		},
		config: config,
	}, nil
}

func (client *Client) validateAudience(token *coreoidc.IDToken) error {
	audiences := client.config.Audiences
	if len(audiences) == 0 {
		audiences = []string{client.config.ClientID}
	}

	for _, expected := range audiences {
		if slices.Contains(token.Audience, expected) {
			return nil
		}
	}

	return fmt.Errorf("OIDC token audience is not accepted")
}

func (client *Client) verify(ctx context.Context, raw string) (*coreoidc.IDToken, error) {
	token, err := client.verifier.Verify(ctx, raw)
	if err != nil {
		return nil, fmt.Errorf("verify OIDC token: %w", err)
	}

	if err := client.validateAudience(token); err != nil {
		return nil, err
	}

	return token, nil
}

func scopes(value any) map[string]struct{} {
	result := make(map[string]struct{})
	text, ok := value.(string)
	if !ok {
		return result
	}

	for scope := range strings.FieldsSeq(text) {
		result[scope] = struct{}{}
	}

	return result
}
