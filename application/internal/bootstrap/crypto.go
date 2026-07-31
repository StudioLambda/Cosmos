package bootstrap

import (
	"github.com/studiolambda/cosmos/contract"
	"github.com/studiolambda/cosmos/framework/crypto/aes"
)

// NewCrypto creates an AES-GCM encrypter from crypto.aes configuration.
func NewCrypto(configuration *contract.Configuration) (*contract.Encrypter, error) {
	config := configuration.From[aes.AESConfig]("crypto.aes")

	driver, err := aes.NewAES(config)
	if err != nil {
		return nil, err
	}

	return contract.NewEncrypter(driver), nil
}
