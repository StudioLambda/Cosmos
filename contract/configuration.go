package contract

import (
	"errors"
	"strings"
)

// ErrConfigurationKeyNotFound indicates that a configuration value is absent.
var ErrConfigurationKeyNotFound = errors.New("configuration key not found")

// ConfigurationDriver defines the contract for configuration backends.
// Implementations populate caller-provided destinations from the underlying
// configuration source.
type ConfigurationDriver interface {
	// Unmarshal decodes the configuration value at key into dest.
	// It returns [ErrConfigurationKeyNotFound] when the key does not exist.
	Unmarshal(key string, dest any) error

	// Has reports whether the given configuration key exists.
	Has(key string) bool

	// Delimiter returns the backend delimiter used to translate Cosmos dotted keys.
	Delimiter() string

	// Extend loads additional providers. Later providers override values loaded
	// earlier by this driver.
	Extend(providers ...ConfigurationProvider) error
}

// ConfigurationProvider supplies values that can be merged into a
// [ConfigurationDriver].
type ConfigurationProvider interface {
	Values() (map[string]any, error)
}

// Configurable populates itself from a prefixed configuration view.
type Configurable interface {
	FromConfiguration(configuration *Configuration)
}

// ConfigurablePointer constrains a pointer to a [Configurable] value.
type ConfigurablePointer[T any] interface {
	*T
	Configurable
}

// Configuration provides a type-safe wrapper over a [ConfigurationDriver].
type Configuration struct {
	config ConfigurationConfig
	driver ConfigurationDriver
}

// ConfigurationConfig configures a [Configuration] view.
type ConfigurationConfig struct {
	prefix string
}

// DefaultConfigurationConfig holds the default unprefixed view settings.
var DefaultConfigurationConfig = ConfigurationConfig{
	prefix: "",
}

// NewConfigurationWith creates a [Configuration] with the supplied view settings.
func NewConfiguration(driver ConfigurationDriver) *Configuration {
	return NewConfigurationWith(driver, DefaultConfigurationConfig)
}

// NewConfiguration creates a new [Configuration] that delegates to the given driver.
func NewConfigurationWith(driver ConfigurationDriver, config ConfigurationConfig) *Configuration {
	return &Configuration{
		config: config,
		driver: driver,
	}
}

// Prefix returns the dotted key prefix applied by this view.
func (configuration *Configuration) Prefix() string {
	return configuration.config.prefix
}

// Driver returns the underlying [ConfigurationDriver].
func (configuration *Configuration) Driver() ConfigurationDriver {
	return configuration.driver
}

// Extend loads additional configuration providers. Later providers override
// values loaded earlier by this configuration.
func (configuration *Configuration) Extend(providers ...ConfigurationProvider) error {
	return configuration.driver.Extend(providers...)
}

// Prefixed returns a view that shares the driver and prepends a normalized
// dotted prefix to subsequent keys.
func (configuration *Configuration) Prefixed(prefix string) *Configuration {
	return NewConfigurationWith(configuration.driver, ConfigurationConfig{
		prefix: configuration.Key(configuration.Prefix(), prefix),
	})
}

func (configuration *Configuration) driverKey(key string) string {
	return strings.ReplaceAll(key, ".", configuration.Driver().Delimiter())
}

// Key joins nonempty key parts into a normalized dotted key.
func (configuration *Configuration) Key(parts ...string) string {
	out := make([]string, 0, len(parts))

	for _, part := range parts {
		part = strings.Trim(part, ".")
		if part != "" {
			out = append(out, part)
		}
	}

	return strings.Join(out, ".")
}

// Get decodes the configuration value at key into T.
func (configuration *Configuration) Get[T any](key string) (res T, err error) {
	key = configuration.driverKey(configuration.Key(configuration.Prefix(), key))
	err = configuration.driver.Unmarshal(key, &res)

	return res, err
}

// GetOr returns the decoded value at key or the fallback when decoding fails.
func (configuration *Configuration) GetOr[T any](key string, fallback T) T {
	res, err := configuration.Get[T](key)

	if err != nil {
		return fallback
	}

	return res
}

// Has reports whether the given configuration key exists.
func (configuration *Configuration) Has(key string) bool {
	return configuration.driver.Has(configuration.driverKey(configuration.Key(configuration.Prefix(), key)))
}

// From allocates T, populates it from prefix through [Configurable], and
// returns the populated value.
func (configuration *Configuration) From[T any, P ConfigurablePointer[T]](prefix string) T {
	t := P(new(T))

	t.FromConfiguration(configuration.Prefixed(prefix))

	return *t
}
