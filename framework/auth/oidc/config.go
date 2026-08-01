package oidc

import "github.com/studiolambda/cosmos/contract"

// Config configures an OIDC client.
type Config struct {
	// IssuerURL is the provider issuer URL used for OIDC discovery.
	IssuerURL string

	// ClientID is the registered OIDC client identifier.
	ClientID string

	// ClientSecret is the confidential client secret. Load it through a secret
	// provider rather than committing it to configuration.
	ClientSecret string

	// RedirectURL is the exact callback URL registered with the provider. It is
	// required when using [Client.Login] or [Client.Callback].
	RedirectURL string

	// Scopes are requested during browser login. Empty scopes default to openid.
	Scopes []string

	// Audiences are accepted JWT audiences. Empty audiences default to ClientID.
	Audiences []string
}

// FromConfiguration populates Config from issuer_url, client_id,
// client_secret, redirect_url, scopes, and audiences configuration values.
func (config *Config) FromConfiguration(configuration *contract.Configuration) {
	config.IssuerURL = configuration.GetOr("issuer_url", "")
	config.ClientID = configuration.GetOr("client_id", "")
	config.ClientSecret = configuration.GetOr("client_secret", "")
	config.RedirectURL = configuration.GetOr("redirect_url", "")
	config.Scopes = configuration.GetOr("scopes", []string(nil))
	config.Audiences = configuration.GetOr("audiences", []string(nil))
}
