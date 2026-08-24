package contract_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/studiolambda/cosmos/contract"
)

type hasherDriver struct {
	value []byte
}

func (driver *hasherDriver) Hash(value []byte) ([]byte, error) {
	driver.value = value

	return []byte("hash"), nil
}

func (driver hasherDriver) Check(value, hash []byte) (bool, error) {
	return string(value) == `{"name":"cosmos"}` && string(hash) == "hash", nil
}

func TestHasherSerializesValues(t *testing.T) {
	t.Parallel()

	driver := &hasherDriver{}
	hasher := contract.NewHasher(driver)
	hash, err := hasher.Hash(struct {
		Name string `json:"name"`
	}{Name: "cosmos"})
	require.NoError(t, err)

	ok, err := hasher.Check(struct {
		Name string `json:"name"`
	}{Name: "cosmos"}, hash)

	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, []byte(`{"name":"cosmos"}`), driver.value)
}
