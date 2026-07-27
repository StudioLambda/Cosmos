package bootstrap

import (
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/rawbytes"
	"github.com/knadh/koanf/v2"
	"github.com/samber/do/v2"
)

func NewConfig(i do.Injector) (*koanf.Koanf, error) {
	files := do.MustInvokeNamed[[]byte](i, "configuration")

	k := koanf.New(".")

	if err := k.Load(rawbytes.Provider(files), yaml.Parser()); err != nil {
		return nil, err
	}

	return k, nil
}
