package bootstrap

import (
	"github.com/studiolambda/cosmos/contract"
	"github.com/studiolambda/cosmos/framework/crypto/aes"
)

// NewCrypto creates an AES-GCM encrypter from crypto.aes configuration.
func NewCrypto(configuration *contract.Configuration) (contract.Encrypter, error) {
	config := aes.ConfigFrom(configuration, "crypto.aes")

	return aes.NewAES(config)
}
