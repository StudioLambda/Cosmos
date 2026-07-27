package bootstrap

import (
	"github.com/studiolambda/cosmos/contract"
	"github.com/studiolambda/cosmos/framework/database/sqlite"
)

// NewDatabase creates a SQLite database from database.sqlite configuration.
func NewDatabase(configuration *contract.Configuration) (*contract.Database, error) {
	config := sqlite.ConfigFrom(configuration, "database.sqlite")

	driver, err := sqlite.New(config)
	if err != nil {
		return nil, err
	}

	return contract.NewDatabase(driver), nil
}
