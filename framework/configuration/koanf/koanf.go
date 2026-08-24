// Package koanf provides a Koanf [contract.ConfigurationDriver].
package koanf

import (
	"fmt"
	"time"

	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/v2"
	"github.com/studiolambda/cosmos/contract"
	frameworkconfiguration "github.com/studiolambda/cosmos/framework/configuration"
)

// Config configures a Koanf-backed [contract.ConfigurationDriver].
type Config struct {
	// Delimiter is the nested key delimiter used when creating an internal Koanf
	// instance. Defaults to ".".
	Delimiter string

	// TimeLayout is the default layout used when decoding [time.Time] values.
	// Defaults to [time.RFC3339].
	TimeLayout string
}

// DefaultConfig returns the default Koanf configuration.
func DefaultConfig() Config {
	return Config{
		Delimiter:  ".",
		TimeLayout: time.RFC3339,
	}
}

func (config Config) withDefaults() Config {
	defaults := DefaultConfig()

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
	config Config
}

// New creates a Koanf-backed configuration driver from providers.
func New(providers ...frameworkconfiguration.Provider) (*Koanf, error) {
	return NewWith(DefaultConfig(), providers...)
}

// NewWith creates a Koanf-backed configuration driver from providers using config.
func NewWith(config Config, providers ...frameworkconfiguration.Provider) (*Koanf, error) {
	config = config.withDefaults()

	driver := &Koanf{
		koanf:  koanf.New(config.Delimiter),
		config: config,
	}

	if err := driver.Extend(providers...); err != nil {
		return nil, err
	}

	return driver, nil
}

// Extend loads additional providers. Later providers override existing values.
func (configuration *Koanf) Extend(providers ...contract.ConfigurationProvider) error {
	values, err := frameworkconfiguration.Resolve(providers...)
	if err != nil {
		return fmt.Errorf("resolve configuration: %w", err)
	}

	if err := configuration.koanf.Load(confmap.Provider(values, configuration.config.Delimiter), nil); err != nil {
		return fmt.Errorf("load configuration: %w", err)
	}

	return nil
}

// NewKoanfFrom creates a new Koanf-backed configuration driver from an
// existing [koanf.Koanf] instance using default settings.
func NewKoanfFrom(instance *koanf.Koanf) *Koanf {
	return &Koanf{
		koanf:  instance,
		config: DefaultConfig(),
	}
}

func (configuration *Koanf) Delimiter() string {
	return configuration.koanf.Delim()
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
