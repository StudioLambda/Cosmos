package crypto_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/studiolambda/cosmos/framework/crypto"
)

func TestItCanCreateAESEncrypter(t *testing.T) {
	t.Parallel()

	key := []byte("12345678901234567890123456789012")
	_, err := crypto.NewAES(key)

	require.NoError(t, err)
}

func TestItCanEncryptAES(t *testing.T) {
	t.Parallel()

	key := []byte("12345678901234567890123456789012")
	e, err := crypto.NewAES(key)

	require.NoError(t, err)

	plain := []byte("Hello, World!")
	_, err = e.Encrypt(plain)

	require.NoError(t, err)
}

func TestItCanDecryptAES(t *testing.T) {
	t.Parallel()

	key := []byte("12345678901234567890123456789012")
	e, err := crypto.NewAES(key)

	require.NoError(t, err)

	plain := []byte("Hello, World!")
	cypher, err := e.Encrypt(plain)

	require.NoError(t, err)

	res, err := e.Decrypt(cypher)

	require.NoError(t, err)
	require.Equal(t, plain, res)
}
