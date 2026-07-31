// Package azurekeyvault provides an Azure Key Vault [contract.SecretDriver].
//
// Wrap [Client] with [contract.NewSecrets] for typed retrieval, or pass it to
// configuration.JSONSecret, configuration.YAMLSecret, or configuration.RawSecret.
// Missing secrets return [contract.ErrSecretNotFound].
package azurekeyvault
