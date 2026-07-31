package bootstrap

import (
	"github.com/studiolambda/cosmos/contract"
	"github.com/studiolambda/cosmos/framework/logger/slog"
)

// NewLogger creates a slog logger from observability.logger configuration.
func NewLogger(configuration *contract.Configuration) (*contract.Logger, error) {
	config := configuration.From[slog.Config]("observability.logger")

	driver, err := slog.New(config)
	if err != nil {
		return nil, err
	}

	return contract.NewLogger(driver), nil
}
