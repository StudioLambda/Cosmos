package crypto_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/studiolambda/cosmos/framework/crypto"
)

func TestItCanCreateAESEncrypter(t *testing.T) {
	key := []byte("12345678901234567890123456789012")
	_, err := crypto.NewAES(key)

	require.NoError(t, err)
}

func TestItCanEncryptAES(t *testing.T) {
	key := []byte("12345678901234567890123456789012")
	e, err := crypto.NewAES(key)

	require.NoError(t, err)

	plain := []byte("Hello, World!")
	_, err = e.Encrypt(plain)

	require.NoError(t, err)
}

func TestItCanDecryptAES(t *testing.T) {
	key := []byte("12345678901234567890123456789012")
	e, _ := crypto.NewAES(key)
	plain := []byte("Hello, World!")
	cypher, _ := e.Encrypt(plain)
	res, err := e.Decrypt(cypher)

	require.NoError(t, err)
	require.Equal(t, plain, res)
}
