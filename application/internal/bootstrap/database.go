package bootstrap

import (
	"github.com/samber/do/v2"
	"github.com/studiolambda/cosmos/contract"
	"github.com/studiolambda/cosmos/framework/database"
)

func NewDatabase(i do.Injector) (*contract.Database, error) {
	driver := do.MustInvoke[contract.DatabaseDriver](i)

	return contract.NewDatabase(driver), nil
}

func NewDatabaseDriver(i do.Injector) (contract.DatabaseDriver, error) {
	config := do.MustInvoke[database.SQLConfig](i)

	return database.NewSQL(config)
}
