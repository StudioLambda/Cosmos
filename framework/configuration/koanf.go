package configuration

import (
	"fmt"
	"time"

	"github.com/knadh/koanf/v2"
	"github.com/studiolambda/cosmos/contract"
)

// KoanfConfig configures a Koanf-backed [contract.ConfigurationDriver].
type KoanfConfig struct {
	// Delimiter is the nested key delimiter used when creating an internal Koanf
	// instance. Defaults to ".".
	Delimiter string

	// TimeLayout is the default layout used when decoding [time.Time] values.
	// Defaults to [time.RFC3339].
	TimeLayout string
}

// DefaultKoanfConfig holds the default Koanf configuration.
var DefaultKoanfConfig = KoanfConfig{
	Delimiter:  ".",
	TimeLayout: time.RFC3339,
}

func (config KoanfConfig) withDefaults() KoanfConfig {
	defaults := DefaultKoanfConfig

	if config.TimeLayout == "" {
		config.TimeLayout = defaults.TimeLayout
	}

	if config.Delimiter == "" {
		config.Delimiter = defaults.Delimiter
	}

	return config
}

// Koanf wraps [koanf.Koanf] to implement [contract.ConfigurationDriver].
type Koanf struct {
	koanf  *koanf.Koanf
	config KoanfConfig
}

// NewKoanf creates a new Koanf-backed configuration driver with default
// settings.
func NewKoanf() *Koanf {
	return NewKoanfWith(DefaultKoanfConfig)
}

// NewKoanfFrom creates a new Koanf-backed configuration driver from an
// existing [koanf.Koanf] instance using default settings.
func NewKoanfFrom(instance *koanf.Koanf) *Koanf {
	return &Koanf{
		koanf:  instance,
		config: DefaultKoanfConfig,
	}
}

// NewKoanfWith creates a new Koanf-backed configuration driver with custom
// settings.
func NewKoanfWith(config KoanfConfig) *Koanf {
	config = config.withDefaults()

	return &Koanf{
		koanf:  koanf.New(config.Delimiter),
		config: config,
	}
}

// Instance returns the underlying [koanf.Koanf] instance.
func (configuration *Koanf) Instance() *koanf.Koanf {
	return configuration.koanf
}

// Has reports whether the given configuration key exists.
func (configuration *Koanf) Has(key string) bool {
	return configuration.koanf.Exists(key)
}

// Unmarshal decodes the configuration value at key into dest.
func (configuration *Koanf) Unmarshal(key string, dest any) error {
	if !configuration.Has(key) {
		return contract.ErrConfigurationKeyNotFound
	}

	switch target := dest.(type) {
	case *string:
		*target = configuration.koanf.String(key)
		return nil
	case *bool:
		*target = configuration.koanf.Bool(key)
		return nil
	case *int:
		*target = configuration.koanf.Int(key)
		return nil
	case *int64:
		*target = configuration.koanf.Int64(key)
		return nil
	case *float64:
		*target = configuration.koanf.Float64(key)
		return nil
	case *time.Duration:
		*target = configuration.koanf.Duration(key)
		return nil
	case *time.Time:
		*target = configuration.koanf.Time(key, configuration.config.TimeLayout)
		return nil
	case *[]string:
		*target = configuration.koanf.Strings(key)
		return nil
	}

	if err := configuration.koanf.Unmarshal(key, dest); err != nil {
		return fmt.Errorf("unmarshal configuration %q: %w", key, err)
	}

	return nil
}
