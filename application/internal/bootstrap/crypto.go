package bootstrap

import (
	"errors"
	"os"

	"github.com/studiolambda/cosmos/framework/crypto/aes"
)

const cryptoAESKeyEnvironment = "COSMOS_CRYPTO_AES_KEY"

// NewCrypto creates an AES-GCM encrypter from crypto.aes configuration.
func NewCrypto() (*aes.AES, error) {
	key, ok := os.LookupEnv(cryptoAESKeyEnvironment)
	if !ok || key == "" {
		return nil, errors.New("COSMOS_CRYPTO_AES_KEY must be configured")
	}

	return aes.NewAES(aes.AESConfig{Key: []byte(key)})
}
