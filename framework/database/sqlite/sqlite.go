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
	config.DSN = configuration.GetOr("dsn", config.DSN)
	config.MaxOpenConns = configuration.GetOr("max_open_conns", config.MaxOpenConns)
	config.MaxIdleConns = configuration.GetOr("max_idle_conns", config.MaxIdleConns)
	config.ConnMaxLifetime = configuration.GetOr("conn_max_lifetime", config.ConnMaxLifetime)
	config.ConnMaxIdleTime = configuration.GetOr("conn_max_idle_time", config.ConnMaxIdleTime)
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
