package configuration_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/studiolambda/cosmos/contract"
	"github.com/studiolambda/cosmos/framework/configuration"
)

type secretDriver struct {
	value []byte
	err   error
}

func (driver secretDriver) Get(context.Context, string) ([]byte, error) {
	return driver.value, driver.err
}

func TestJSONSecretParsesConfigurationObject(t *testing.T) {
	t.Parallel()

	values, err := configuration.Resolve(configuration.JSONSecret(
		context.Background(),
		secretDriver{value: []byte(`{"database":{"postgres":{"dsn":"postgres://example"}}}`)},
		"database",
	))

	require.NoError(t, err)
	require.Equal(t, "postgres://example", values["database"].(map[string]any)["postgres"].(map[string]any)["dsn"])
}

func TestYAMLSecretParsesConfigurationObject(t *testing.T) {
	t.Parallel()

	values, err := configuration.Resolve(configuration.YAMLSecret(
		context.Background(),
		secretDriver{value: []byte("database:\n  postgres:\n    dsn: postgres://example\n")},
		"database",
	))

	require.NoError(t, err)
	require.Equal(t, "postgres://example", values["database"].(map[string]any)["postgres"].(map[string]any)["dsn"])
}

func TestRawSecretAssignsValueToKey(t *testing.T) {
	t.Parallel()

	values, err := configuration.Resolve(configuration.RawSecret(
		context.Background(),
		secretDriver{value: []byte("postgres://example")},
		"database-dsn",
		"database.postgres.dsn",
	))

	require.NoError(t, err)
	require.Equal(t, "postgres://example", values["database"].(map[string]any)["postgres"].(map[string]any)["dsn"])
}

func TestSecretReturnsRetrievalError(t *testing.T) {
	t.Parallel()

	_, err := configuration.Resolve(configuration.JSONSecret(
		context.Background(),
		secretDriver{err: contract.ErrSecretNotFound},
		"missing",
	))

	require.ErrorIs(t, err, contract.ErrSecretNotFound)
}

func TestRawSecretRejectsNonUTF8Value(t *testing.T) {
	t.Parallel()

	_, err := configuration.Resolve(configuration.RawSecret(
		context.Background(),
		secretDriver{value: []byte{0xff}},
		"binary",
		"database.postgres.dsn",
	))

	require.Error(t, err)
}

func TestSecretReturnsErrorForEmptyKey(t *testing.T) {
	t.Parallel()

	_, err := configuration.Resolve(configuration.RawSecret(
		context.Background(),
		secretDriver{value: []byte("value")},
		"secret",
		"",
	))

	require.Error(t, err)
}

func TestSecretDoesNotDiscardErrors(t *testing.T) {
	t.Parallel()

	expected := errors.New("unavailable")
	_, err := configuration.Resolve(configuration.YAMLSecret(
		context.Background(),
		secretDriver{err: expected},
		"secret",
	))

	require.ErrorIs(t, err, expected)
}
