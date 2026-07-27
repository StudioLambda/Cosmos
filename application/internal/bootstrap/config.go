package bootstrap

import (
	"io/fs"

	"github.com/studiolambda/cosmos/contract"
	"github.com/studiolambda/cosmos/framework/configuration/koanf"
)

// NewConfig creates a configuration service from an embedded filesystem.
func NewConfig(configurationFS fs.FS) (*contract.Configuration, error) {
	config, err := koanf.New(koanf.Config{Filesystem: configurationFS})
	if err != nil {
		return nil, err
	}

	return contract.NewConfiguration(config), nil
}
