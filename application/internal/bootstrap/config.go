package bootstrap

import (
	"github.com/studiolambda/cosmos/contract"
	"github.com/studiolambda/cosmos/framework/configuration"
	"github.com/studiolambda/cosmos/framework/configuration/koanf"
)

// NewConfig creates a configuration service from providers.
func NewConfig(providers ...configuration.Provider) (*contract.Configuration, error) {
	config, err := koanf.New(providers...)
	if err != nil {
		return nil, err
	}

	return contract.NewConfiguration(config), nil
}
