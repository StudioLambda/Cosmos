// Package awssm provides an AWS Secrets Manager [contract.SecretDriver].
//
// Wrap [Client] with [contract.NewSecrets] for typed retrieval, or pass it to
// configuration.JSONSecret, configuration.YAMLSecret, or configuration.RawSecret.
// Missing secrets return [contract.ErrSecretNotFound].
package awssm
