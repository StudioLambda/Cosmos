package config

import (
	"github.com/knadh/koanf/v2"
	"github.com/samber/do/v2"
	"github.com/studiolambda/cosmos/framework/correlation"
)

func NewObservabilityCorrelation(i do.Injector) (correlation.MiddlewareConfig, error) {
	k := do.MustInvoke[*koanf.Koanf](i)

	config := correlation.MiddlewareConfig{
		Header: k.String("observability.correlation.header"),
	}

	return config, nil
}
