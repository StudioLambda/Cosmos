package config

import (
	"github.com/knadh/koanf/v2"
	"github.com/samber/do/v2"
	"github.com/studiolambda/cosmos/framework/database"
)

func NewSQLDatabase(i do.Injector) (database.SQLConfig, error) {
	k := do.MustInvoke[*koanf.Koanf](i)

	config := database.SQLConfig{
		Driver:          k.String("database.driver"),
		DSN:             k.String("database.dsn"),
		MaxOpenConns:    k.Int("database.max_open_conns"),
		MaxIdleConns:    k.Int("database.max_idle_conns"),
		ConnMaxLifetime: k.Duration("database.conn_max_lifetime"),
		ConnMaxIdleTime: k.Duration("database.conn_max_idle_time"),
	}

	return config, nil
}
