package contract_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/studiolambda/cosmos/contract"
)

type secretDriver struct {
	value []byte
}

func (driver secretDriver) Get(context.Context, string) ([]byte, error) {
	return driver.value, nil
}

func TestSecretsGetDecodesJSON(t *testing.T) {
	t.Parallel()

	secrets := contract.NewSecrets(secretDriver{value: []byte(`{"name":"cosmos"}`)})
	value, err := secrets.Get[struct {
		Name string `json:"name"`
	}](context.Background(), "app")

	require.NoError(t, err)
	require.Equal(t, "cosmos", value.Name)
}

func TestSecretsStringReturnsRawString(t *testing.T) {
	t.Parallel()

	secrets := contract.NewSecrets(secretDriver{value: []byte("secret")})
	value, err := secrets.String(context.Background(), "password")

	require.NoError(t, err)
	require.Equal(t, "secret", value)
}
