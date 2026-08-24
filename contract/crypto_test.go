package contract_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/studiolambda/cosmos/contract"
)

type encrypterDriver struct{}

func (encrypterDriver) Encrypt(value []byte) ([]byte, error) { return value, nil }
func (encrypterDriver) Decrypt(value []byte) ([]byte, error) { return value, nil }
func (encrypterDriver) Close() error                         { return nil }

func TestEncrypterSerializesAndDeserializesValues(t *testing.T) {
	t.Parallel()

	encrypter := contract.NewEncrypter(encrypterDriver{})
	ciphertext, err := encrypter.Encrypt(struct {
		Name string `json:"name"`
	}{Name: "cosmos"})
	require.NoError(t, err)

	value, err := encrypter.Decrypt[struct {
		Name string `json:"name"`
	}](ciphertext)

	require.NoError(t, err)
	require.Equal(t, "cosmos", value.Name)
}
