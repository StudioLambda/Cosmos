package aes_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/studiolambda/cosmos/framework/crypto/aes"
)

func TestItCanCreateEncrypter(t *testing.T) {
	key := []byte("12345678901234567890123456789012")
	_, err := aes.NewEncrypter(key)

	require.NoError(t, err)
}

func TestItCanEncrypt(t *testing.T) {
	key := []byte("12345678901234567890123456789012")
	e, err := aes.NewEncrypter(key)

	require.NoError(t, err)

	plain := []byte("Hello, World!")
	_, err = e.Encrypt(plain)

	require.NoError(t, err)
}

func TestItCanDecrypt(t *testing.T) {
	key := []byte("12345678901234567890123456789012")
	e, _ := aes.NewEncrypter(key)
	plain := []byte("Hello, World!")
	cypher, _ := e.Encrypt(plain)
	res, err := e.Decrypt(cypher)

	require.NoError(t, err)
	require.Equal(t, plain, res)
}
