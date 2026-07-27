package contract

import "errors"

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
}

// Configuration provides a type-safe wrapper over a [ConfigurationDriver].
type Configuration struct {
	driver ConfigurationDriver
}

// NewConfiguration creates a new [Configuration] that delegates to the given driver.
func NewConfiguration(driver ConfigurationDriver) *Configuration {
	return &Configuration{
		driver: driver,
	}
}

// Driver returns the underlying [ConfigurationDriver].
func (configuration *Configuration) Driver() ConfigurationDriver {
	return configuration.driver
}

// Get decodes the configuration value at key into T.
func (configuration *Configuration) Get[T any](key string) (res T, err error) {
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
	return configuration.driver.Has(key)
}
