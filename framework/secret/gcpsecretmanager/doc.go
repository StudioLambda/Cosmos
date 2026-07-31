// Package gcpsecretmanager provides a Google Cloud Secret Manager [contract.SecretDriver].
//
// Wrap [Client] with [contract.NewSecrets] for typed retrieval, or pass it to
// configuration.JSONSecret, configuration.YAMLSecret, or configuration.RawSecret.
// Missing secrets return [contract.ErrSecretNotFound].
package gcpsecretmanager
