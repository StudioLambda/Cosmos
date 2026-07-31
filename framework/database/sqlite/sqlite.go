package sqlite

import (
	"time"

	"github.com/studiolambda/cosmos/contract"
	"github.com/studiolambda/cosmos/framework/database/internal/sqlx"

	_ "modernc.org/sqlite"
)

// Config configures a SQLite connection and its pool.
type Config struct {
	DSN             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

// FromConfiguration populates the SQLite configuration from configuration.
func (config *Config) FromConfiguration(configuration *contract.Configuration) {
	config.DSN = configuration.GetOr("dsn", "")
	config.MaxOpenConns = configuration.GetOr("max_open_conns", 0)
	config.MaxIdleConns = configuration.GetOr("max_idle_conns", 0)
	config.ConnMaxLifetime = configuration.GetOr("conn_max_lifetime", time.Duration(0))
	config.ConnMaxIdleTime = configuration.GetOr("conn_max_idle_time", time.Duration(0))
}

// New connects to SQLite and returns a contract-compatible database driver.
func New(config Config) (contract.DatabaseDriver, error) {
	return sqlx.NewSQL(sqlx.SQLConfig{
		Driver:          "sqlite",
		DSN:             config.DSN,
		MaxOpenConns:    config.MaxOpenConns,
		MaxIdleConns:    config.MaxIdleConns,
		ConnMaxLifetime: config.ConnMaxLifetime,
		ConnMaxIdleTime: config.ConnMaxIdleTime,
	})
}
