package crypto_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/studiolambda/cosmos/framework/crypto"
)

func TestItCanCreateChaCha20Encrypter(t *testing.T) {
	t.Parallel()

	key := []byte("12345678901234567890123456789012")
	_, err := crypto.NewChaCha20(key)

	require.NoError(t, err)
}

func TestItCanEncryptChaCha20(t *testing.T) {
	t.Parallel()

	key := []byte("12345678901234567890123456789012")
	e, err := crypto.NewChaCha20(key)

	require.NoError(t, err)

	plain := []byte("Hello, World!")
	_, err = e.Encrypt(plain)

	require.NoError(t, err)
}

func TestItCanDecryptChaCha20(t *testing.T) {
	t.Parallel()

	key := []byte("12345678901234567890123456789012")
	e, err := crypto.NewChaCha20(key)

	require.NoError(t, err)

	plain := []byte("Hello, World!")
	cypher, err := e.Encrypt(plain)

	require.NoError(t, err)

	res, err := e.Decrypt(cypher)

	require.NoError(t, err)
	require.Equal(t, plain, res)
}
