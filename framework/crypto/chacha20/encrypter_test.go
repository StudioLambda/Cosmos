package chacha20_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/studiolambda/cosmos/framework/crypto/chacha20"
)

func TestItCanCreateEncrypter(t *testing.T) {
	key := []byte("12345678901234567890123456789012")
	_, err := chacha20.NewEncrypter(key)

	require.NoError(t, err)
}

func TestItCanEncrypt(t *testing.T) {
	key := []byte("12345678901234567890123456789012")
	e, err := chacha20.NewEncrypter(key)

	require.NoError(t, err)

	plain := []byte("Hello, World!")
	_, err = e.Encrypt(plain)

	require.NoError(t, err)
}

func TestItCanDecrypt(t *testing.T) {
	key := []byte("12345678901234567890123456789012")
	e, _ := chacha20.NewEncrypter(key)
	plain := []byte("Hello, World!")
	cypher, _ := e.Encrypt(plain)
	res, err := e.Decrypt(cypher)

	require.NoError(t, err)
	require.Equal(t, plain, res)
}
