package bootstrap

import (
	"errors"

	"github.com/studiolambda/cosmos/contract"
	"github.com/studiolambda/cosmos/framework/crypto/aes"
)

// NewCrypto creates an AES-GCM encrypter from crypto.aes configuration.
func NewCrypto(configuration *contract.Configuration) (*aes.AES, error) {
	key, err := configuration.Get[string]("crypto.aes.key")
	if err != nil || key == "" {
		return nil, errors.New("crypto.aes.key must be configured")
	}

	return aes.NewAES(aes.AESConfig{Key: []byte(key)})
}
