// Package oidc authenticates Cosmos applications with OpenID Connect providers.
//
// It supports Authorization Code login with PKCE and server-backed Cosmos
// sessions, plus JWT bearer access-token authentication. Session identities
// take precedence over bearer credentials. Login validates state, nonce, PKCE,
// issuer, audience, signature, and expiry; bearer tokens validate issuer,
// audience, signature, and expiry.
//
// The package supports local logout only. It does not yet support multiple
// providers, configurable role/group policies, token refresh, opaque-token
// introspection, provider logout, back-channel logout, JARM, PAR, or DPoP.
package oidc
