package sqlx_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/studiolambda/cosmos/contract"
	"github.com/studiolambda/cosmos/framework/configuration"
	"github.com/studiolambda/cosmos/framework/configuration/koanf"
	"github.com/studiolambda/cosmos/framework/database/internal/sqlx"
)

func TestSQLConfigFromConfigurationPreservesDriverAndDSN(t *testing.T) {
	t.Parallel()

	config := sqlx.SQLConfig{Driver: "pgx", DSN: "required"}
	config.FromConfiguration(newConfiguration(t, nil))

	require.Equal(t, "pgx", config.Driver)
	require.Equal(t, "required", config.DSN)
}

func TestSQLConfigFromConfigurationOverridesValues(t *testing.T) {
	t.Parallel()

	config := sqlx.SQLConfig{Driver: "pgx"}
	config.FromConfiguration(newConfiguration(t, map[string]any{
		"driver": "mysql", "dsn": "configured", "max_open_conns": 5,
		"max_idle_conns": 2, "conn_max_lifetime": "1m", "conn_max_idle_time": "2m",
	}))

	require.Equal(t, "mysql", config.Driver)
	require.Equal(t, "configured", config.DSN)
	require.Equal(t, 5, config.MaxOpenConns)
	require.Equal(t, 2, config.MaxIdleConns)
	require.Equal(t, time.Minute, config.ConnMaxLifetime)
	require.Equal(t, 2*time.Minute, config.ConnMaxIdleTime)
}

func newConfiguration(t *testing.T, values map[string]any) *contract.Configuration {
	t.Helper()

	driver, err := koanf.New(configuration.Map(values))
	require.NoError(t, err)

	return contract.NewConfiguration(driver)
}
