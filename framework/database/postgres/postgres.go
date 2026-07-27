// Package postgres provides a PostgreSQL [contract.DatabaseDriver].
package postgres

import (
	"time"

	"github.com/studiolambda/cosmos/contract"
	"github.com/studiolambda/cosmos/framework/database/internal/sqlx"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// Config configures a PostgreSQL connection and its pool.
type Config struct {
	DSN             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

// New connects to PostgreSQL and returns a contract-compatible database driver.
func New(config Config) (contract.DatabaseDriver, error) {
	return sqlx.NewSQL(sqlx.SQLConfig{
		Driver:          "pgx",
		DSN:             config.DSN,
		MaxOpenConns:    config.MaxOpenConns,
		MaxIdleConns:    config.MaxIdleConns,
		ConnMaxLifetime: config.ConnMaxLifetime,
		ConnMaxIdleTime: config.ConnMaxIdleTime,
	})
}
