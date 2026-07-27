package bootstrap

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewCryptoRequiresEnvironmentKey(t *testing.T) {
	t.Setenv(cryptoAESKeyEnvironment, "")

	_, err := NewCrypto()

	require.Error(t, err)
}

func TestNewCryptoUsesEnvironmentKey(t *testing.T) {
	t.Setenv(cryptoAESKeyEnvironment, "12345678901234567890123456789012")

	encrypter, err := NewCrypto()

	require.NoError(t, err)
	require.NoError(t, encrypter.Close())
}
