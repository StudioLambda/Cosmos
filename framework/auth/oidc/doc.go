// Package oidc authenticates Cosmos applications with OpenID Connect providers.
//
// It supports Authorization Code login with PKCE and server-backed Cosmos
// sessions, plus JWT bearer access-token authentication. Session identities
// take precedence over bearer credentials. Login validates state, nonce, PKCE,
// issuer, audience, signature, and expiry; bearer tokens validate issuer,
// audience, signature, and expiry.
//
// Authorization Code handlers require middleware.Session. Use
// [Client.RequireAuthentication] before [Client.RequireScopes], then retrieve
// the resulting identity with [From]. Production deployments must use HTTPS,
// secure session cookies, an exact provider-registered callback URL, and CSRF
// protection on authenticated state-changing routes.
//
// Example:
//
//	client, err := New(ctx, config)
//	if err != nil {
//		return err
//	}
//	app.Use(middleware.Session(driver, middleware.DefaultSessionConfig))
//	app.Get("/login", client.Login())
//	app.Get("/oidc/callback", client.Callback())
//	app.With(client.RequireAuthentication()).Get("/me", handler)
//	app.With(client.RequireAuthentication(), client.RequireScopes("profile")).Get("/profile", handler)
//
// The package supports local logout only. It does not yet support multiple
// providers, configurable role/group policies, token refresh, opaque-token
// introspection, provider logout, back-channel logout, JARM, PAR, or DPoP.
package oidc
