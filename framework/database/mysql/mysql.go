package mysql

import (
	"time"

	"github.com/studiolambda/cosmos/contract"
	"github.com/studiolambda/cosmos/framework/database/internal/sqlx"

	_ "github.com/go-sql-driver/mysql"
)

// Config configures a MySQL connection and its pool.
type Config struct {
	DSN             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

// DefaultConfig returns the default MySQL connection configuration.
func DefaultConfig() Config {
	return Config{}
}

// FromConfiguration populates the MySQL configuration from configuration.
func (config *Config) FromConfiguration(configuration *contract.Configuration) {
	dsn := config.DSN
	*config = DefaultConfig()
	config.DSN = configuration.GetOr("dsn", dsn)
	config.MaxOpenConns = configuration.GetOr("max_open_conns", config.MaxOpenConns)
	config.MaxIdleConns = configuration.GetOr("max_idle_conns", config.MaxIdleConns)
	config.ConnMaxLifetime = configuration.GetOr("conn_max_lifetime", config.ConnMaxLifetime)
	config.ConnMaxIdleTime = configuration.GetOr("conn_max_idle_time", config.ConnMaxIdleTime)
}

// New connects to MySQL and returns a contract-compatible database driver.
func New(config Config) (contract.DatabaseDriver, error) {
	return sqlx.NewSQL(sqlx.SQLConfig{
		Driver:          "mysql",
		DSN:             config.DSN,
		MaxOpenConns:    config.MaxOpenConns,
		MaxIdleConns:    config.MaxIdleConns,
		ConnMaxLifetime: config.ConnMaxLifetime,
		ConnMaxIdleTime: config.ConnMaxIdleTime,
	})
}
