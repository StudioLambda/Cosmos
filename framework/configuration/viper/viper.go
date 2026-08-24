// Package viper provides a Viper [contract.ConfigurationDriver].
package viper

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
	"github.com/studiolambda/cosmos/contract"
	frameworkconfiguration "github.com/studiolambda/cosmos/framework/configuration"
)

// Config configures a Viper-backed [contract.ConfigurationDriver].
type Config struct {
	// Delimiter is the nested key delimiter. Defaults to ".".
	Delimiter string

	// TimeLayout is the layout used when decoding [time.Time]. Defaults to
	// [time.RFC3339].
	TimeLayout string
}

// DefaultConfig returns the default Viper configuration.
func DefaultConfig() Config {
	return Config{
		Delimiter:  ".",
		TimeLayout: time.RFC3339,
	}
}

func (config Config) withDefaults() Config {
	if config.Delimiter == "" {
		config.Delimiter = DefaultConfig().Delimiter
	}

	if config.TimeLayout == "" {
		config.TimeLayout = DefaultConfig().TimeLayout
	}

	return config
}

// Viper wraps a dedicated [viper.Viper] instance.
type Viper struct {
	viper  *viper.Viper
	config Config
}

// New creates a Viper-backed configuration driver from providers.
func New(providers ...frameworkconfiguration.Provider) (*Viper, error) {
	return NewWith(DefaultConfig(), providers...)
}

// NewWith creates a Viper-backed configuration driver from providers using config.
func NewWith(config Config, providers ...frameworkconfiguration.Provider) (*Viper, error) {
	config = config.withDefaults()

	instance := viper.NewWithOptions(viper.KeyDelimiter(config.Delimiter))
	driver := &Viper{viper: instance, config: config}

	if err := driver.Extend(providers...); err != nil {
		return nil, err
	}

	return driver, nil
}

// Extend loads additional providers. Later providers override existing values.
func (configuration *Viper) Extend(providers ...contract.ConfigurationProvider) error {
	values, err := frameworkconfiguration.Resolve(providers...)
	if err != nil {
		return fmt.Errorf("resolve configuration: %w", err)
	}

	if err := configuration.viper.MergeConfigMap(values); err != nil {
		return fmt.Errorf("load configuration: %w", err)
	}

	return nil
}

// NewViperFrom wraps an existing [viper.Viper] instance.
// Due to a limitation in viper, you need to supply the delimiter
// used when creating the viper instance.
func NewViperFrom(instance *viper.Viper, delimiter string) *Viper {
	config := DefaultConfig()
	config.Delimiter = delimiter

	return &Viper{
		viper:  instance,
		config: config,
	}
}

func (configuration *Viper) Delimiter() string {
	return configuration.config.Delimiter
}

// Instance returns the underlying [viper.Viper] instance.
func (configuration *Viper) Instance() *viper.Viper {
	return configuration.viper
}

// Has reports whether key exists.
func (configuration *Viper) Has(key string) bool {
	return configuration.viper.IsSet(key)
}

// Unmarshal decodes key into dest.
func (configuration *Viper) Unmarshal(key string, dest any) error {
	if !configuration.Has(key) {
		return contract.ErrConfigurationKeyNotFound
	}

	switch target := dest.(type) {
	case *string:
		*target = configuration.viper.GetString(key)
		return nil
	case *bool:
		*target = configuration.viper.GetBool(key)
		return nil
	case *int:
		*target = configuration.viper.GetInt(key)
		return nil
	case *int64:
		*target = configuration.viper.GetInt64(key)
		return nil
	case *float64:
		*target = configuration.viper.GetFloat64(key)
		return nil
	case *time.Duration:
		*target = configuration.viper.GetDuration(key)
		return nil
	case *time.Time:
		value, err := time.Parse(configuration.config.TimeLayout, configuration.viper.GetString(key))
		if err != nil {
			return fmt.Errorf("unmarshal configuration %q: %w", key, err)
		}

		*target = value
		return nil
	case *[]string:
		*target = configuration.viper.GetStringSlice(key)
		return nil
	}

	if err := configuration.viper.UnmarshalKey(key, dest); err != nil {
		return fmt.Errorf("unmarshal configuration %q: %w", key, err)
	}

	return nil
}
