// Package koanf provides a Koanf [contract.ConfigurationDriver].
package koanf

import (
	"errors"
	"fmt"
	"io/fs"
	"path"
	"strings"
	"time"

	"github.com/knadh/koanf/parsers/yaml"
	fsprovider "github.com/knadh/koanf/providers/fs"
	"github.com/knadh/koanf/v2"
	"github.com/studiolambda/cosmos/contract"
)

// Config configures a Koanf-backed [contract.ConfigurationDriver].
type Config struct {
	// Filesystem provides the configuration files. It is required by [New].
	Filesystem fs.FS

	// Directory contains YAML configuration files. Defaults to config.
	Directory string

	// Delimiter is the nested key delimiter used when creating an internal Koanf
	// instance. Defaults to ".".
	Delimiter string

	// TimeLayout is the default layout used when decoding [time.Time] values.
	// Defaults to [time.RFC3339].
	TimeLayout string
}

// DefaultConfig holds the default Koanf configuration.
var DefaultConfig = Config{
	Directory:  "config",
	Delimiter:  ".",
	TimeLayout: time.RFC3339,
}

func (config Config) withDefaults() Config {
	defaults := DefaultConfig

	if config.Directory == "" {
		config.Directory = defaults.Directory
	}

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

// New creates a Koanf-backed configuration driver and loads regular YAML files
// from the configured directory in lexical filename order.
func New(config Config) (*Koanf, error) {
	config = config.withDefaults()

	if config.Filesystem == nil {
		return nil, errors.New("configuration filesystem cannot be nil")
	}

	entries, err := fs.ReadDir(config.Filesystem, config.Directory)
	if err != nil {
		return nil, fmt.Errorf("read configuration directory %q: %w", config.Directory, err)
	}

	driver := &Koanf{
		koanf:  koanf.New(config.Delimiter),
		config: config,
	}

	for _, entry := range entries {
		if entry.IsDir() || !isYAML(entry.Name()) {
			continue
		}

		filename := path.Join(config.Directory, entry.Name())

		if err := driver.koanf.Load(fsprovider.Provider(config.Filesystem, filename), yaml.Parser()); err != nil {
			return nil, fmt.Errorf("load configuration %q: %w", filename, err)
		}
	}

	return driver, nil
}

// NewKoanfFrom creates a new Koanf-backed configuration driver from an
// existing [koanf.Koanf] instance using default settings.
func NewKoanfFrom(instance *koanf.Koanf) *Koanf {
	return &Koanf{
		koanf:  instance,
		config: DefaultConfig,
	}
}

func isYAML(filename string) bool {
	extension := strings.ToLower(path.Ext(filename))

	return extension == ".yaml" || extension == ".yml"
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
