package oidc

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"

	requestpkg "github.com/studiolambda/cosmos/contract/request"
	"github.com/studiolambda/cosmos/contract/response"
	"github.com/studiolambda/cosmos/framework"
	"golang.org/x/oauth2"
)

const (
	identitySessionKey = "cosmos:oidc:identity"
	stateSessionKey    = "cosmos:oidc:state"
	nonceSessionKey    = "cosmos:oidc:nonce"
	verifierSessionKey = "cosmos:oidc:verifier"
	returnToSessionKey = "cosmos:oidc:return-to"
)

// Login starts an Authorization Code login flow with PKCE.
func (client *Client) Login() framework.Handler {
	return func(w http.ResponseWriter, request *http.Request) error {
		if client.config.RedirectURL == "" {
			return fmt.Errorf("OIDC redirect URL cannot be empty")
		}

		session, ok := requestpkg.Session(request)
		if !ok {
			return ErrUnauthenticated
		}

		state, err := random()
		if err != nil {
			return fmt.Errorf("generate OIDC state: %w", err)
		}

		nonce, err := random()
		if err != nil {
			return fmt.Errorf("generate OIDC nonce: %w", err)
		}

		verifier, err := random()
		if err != nil {
			return fmt.Errorf("generate OIDC verifier: %w", err)
		}

		returnTo := request.URL.Query().Get("return_to")
		if returnTo == "" {
			returnTo = "/"
		}

		if err := response.SafeRedirect(httplessWriter{}, http.StatusFound, returnTo); err != nil {
			if errors.Is(err, response.ErrUnsafeRedirect) {
				return invalidLoginRequest(err)
			}

			return fmt.Errorf("validate OIDC return URL: %w", err)
		}

		session.Put(stateSessionKey, state)
		session.Put(nonceSessionKey, nonce)
		session.Put(verifierSessionKey, verifier)
		session.Put(returnToSessionKey, returnTo)

		digest := sha256.Sum256([]byte(verifier))
		challenge := base64.RawURLEncoding.EncodeToString(digest[:])
		url := client.oauth.AuthCodeURL(state,
			oauth2.AccessTypeOffline,
			oauth2.SetAuthURLParam("nonce", nonce),
			oauth2.SetAuthURLParam("code_challenge", challenge),
			oauth2.SetAuthURLParam("code_challenge_method", "S256"),
		)

		return response.Redirect(w, http.StatusFound, url)
	}
}

// Callback completes an Authorization Code login flow.
func (client *Client) Callback() framework.Handler {
	return func(w http.ResponseWriter, request *http.Request) error {
		session, ok := requestpkg.Session(request)
		if !ok {
			return ErrUnauthenticated
		}

		state, _ := session.Get[string](stateSessionKey)
		nonce, _ := session.Get[string](nonceSessionKey)
		verifier, _ := session.Get[string](verifierSessionKey)
		returnTo, _ := session.Get[string](returnToSessionKey)
		session.Delete(stateSessionKey)
		session.Delete(nonceSessionKey)
		session.Delete(verifierSessionKey)
		session.Delete(returnToSessionKey)

		if state == "" || subtle.ConstantTimeCompare([]byte(state), []byte(request.URL.Query().Get("state"))) != 1 || verifier == "" {
			return invalidCallback(nil)
		}

		code := request.URL.Query().Get("code")
		if code == "" || request.URL.Query().Get("error") != "" {
			return invalidCallback(nil)
		}

		token, err := client.oauth.Exchange(request.Context(), code, oauth2.SetAuthURLParam("code_verifier", verifier))
		if err != nil {
			return providerUnavailable(err)
		}

		raw, ok := token.Extra("id_token").(string)
		if !ok {
			return invalidCallback(nil)
		}

		idToken, err := client.verify(request.Context(), raw)
		if err != nil || nonce == "" || subtle.ConstantTimeCompare([]byte(idToken.Nonce), []byte(nonce)) != 1 {
			return invalidCallback(err)
		}

		identity, err := identityFromToken(idToken)
		if err != nil {
			return err
		}

		session.Regenerate()

		session.Put(identitySessionKey, identity)

		if returnTo == "" {
			returnTo = "/"
		}

		if err := response.SafeRedirect(w, http.StatusFound, returnTo); err != nil {
			if errors.Is(err, response.ErrUnsafeRedirect) {
				return invalidCallback(err)
			}

			return fmt.Errorf("redirect after OIDC callback: %w", err)
		}

		return nil
	}
}

// Logout clears the local OIDC identity and regenerates the session.
func (client *Client) Logout() framework.Handler {
	return func(w http.ResponseWriter, request *http.Request) error {
		session, ok := requestpkg.Session(request)
		if !ok {
			return ErrUnauthenticated
		}

		session.Delete(identitySessionKey)

		session.Regenerate()

		return response.SafeRedirect(w, http.StatusFound, "/")
	}
}

func random() (string, error) {
	value := make([]byte, 32)
	if _, err := rand.Read(value); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(value), nil
}

// httplessWriter validates redirects without writing a response.
type httplessWriter struct{}

func (httplessWriter) Header() http.Header       { return http.Header{} }
func (httplessWriter) Write([]byte) (int, error) { return 0, nil }
func (httplessWriter) WriteHeader(int)           {}
