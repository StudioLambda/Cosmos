package oidc

import (
	"context"
	"fmt"
	"strings"

	coreoidc "github.com/coreos/go-oidc/v3/oidc"
	"github.com/studiolambda/cosmos/contract"
	"golang.org/x/oauth2"
)

// Config configures an OIDC client.
type Config struct {
	IssuerURL    string
	ClientID     string
	ClientSecret string
	RedirectURL  string
	Scopes       []string
	Audiences    []string
}

// FromConfiguration populates Config from OIDC configuration.
func (config *Config) FromConfiguration(configuration *contract.Configuration) {
	config.IssuerURL = configuration.GetOr("issuer_url", "")
	config.ClientID = configuration.GetOr("client_id", "")
	config.ClientSecret = configuration.GetOr("client_secret", "")
	config.RedirectURL = configuration.GetOr("redirect_url", "")
	config.Scopes = configuration.GetOr("scopes", []string(nil))
	config.Audiences = configuration.GetOr("audiences", []string(nil))
}

// Client authenticates requests with one OIDC provider.
type Client struct {
	provider *coreoidc.Provider
	verifier *coreoidc.IDTokenVerifier
	oauth    oauth2.Config
	config   Config
}

// New discovers an OIDC provider and creates a client.
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
		for _, audience := range token.Audience {
			if audience == expected {
				return nil
			}
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
