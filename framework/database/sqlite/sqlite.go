// Package sqlite provides a SQLite [contract.DatabaseDriver].
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

// ConfigFrom returns a Config read from configuration below prefix.
func ConfigFrom(configuration *contract.Configuration, prefix string) Config {
	return Config{
		DSN:             configuration.GetOr(prefix+".dsn", ""),
		MaxOpenConns:    configuration.GetOr(prefix+".max_open_conns", 0),
		MaxIdleConns:    configuration.GetOr(prefix+".max_idle_conns", 0),
		ConnMaxLifetime: configuration.GetOr(prefix+".conn_max_lifetime", time.Duration(0)),
		ConnMaxIdleTime: configuration.GetOr(prefix+".conn_max_idle_time", time.Duration(0)),
	}
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
