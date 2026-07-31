// Package configuration provides reusable configuration sources.
package configuration

import (
	"encoding/json/v2"
	"fmt"
	"io/fs"
	"maps"
	"os"
	"path"
	"strings"

	"github.com/studiolambda/cosmos/contract"
	"github.com/studiolambda/cosmos/framework/configuration/internal/environment"

	"gopkg.in/yaml.v3"
)

// Provider supplies configuration values. Later providers take precedence over
// earlier providers.
type Provider = contract.ConfigurationProvider

type mapProvider map[string]any

// Map creates a provider from values. Keys may use dot notation or nested maps.
func Map(values map[string]any) Provider {
	return mapProvider(values)
}

func (provider mapProvider) Values() (map[string]any, error) {
	return normalize(provider), nil
}

type filesystemProvider struct {
	filesystem fs.FS
	directory  string
}

// FilesystemOption configures a filesystem provider.
type FilesystemOption func(*filesystemProvider)

// Directory sets the directory from which [Filesystem] loads YAML files.
func Directory(directory string) FilesystemOption {
	return func(provider *filesystemProvider) {
		provider.directory = directory
	}
}

// Filesystem creates a provider that loads YAML files in lexical filename order
// from config by default.
func Filesystem(filesystem fs.FS, options ...FilesystemOption) Provider {
	provider := filesystemProvider{
		filesystem: filesystem,
		directory:  "config",
	}

	for _, option := range options {
		option(&provider)
	}

	return provider
}

func (provider filesystemProvider) Values() (map[string]any, error) {
	if provider.filesystem == nil {
		return nil, fmt.Errorf("configuration filesystem cannot be nil")
	}

	entries, err := fs.ReadDir(provider.filesystem, provider.directory)
	if err != nil {
		return nil, fmt.Errorf("read configuration directory %q: %w", provider.directory, err)
	}

	values := make(map[string]any)

	for _, entry := range entries {
		if entry.IsDir() || (!isYAML(entry.Name()) && !isJSON(entry.Name())) {
			continue
		}

		filename := path.Join(provider.directory, entry.Name())
		contents, err := fs.ReadFile(provider.filesystem, filename)
		if err != nil {
			return nil, fmt.Errorf("read configuration %q: %w", filename, err)
		}

		parsed, err := parse(entry.Name(), os.ExpandEnv(string(contents)))
		if err != nil {
			return nil, fmt.Errorf("parse configuration %q: %w", filename, err)
		}

		merge(values, normalize(parsed))
	}

	return values, nil
}

type environmentProvider string

// Environment creates a provider for PREFIX__SECTION__KEY environment variables.
// An empty prefix provides no values.
func Environment(prefix string) Provider {
	return environmentProvider(prefix)
}

func (provider environmentProvider) Values() (map[string]any, error) {
	return normalize(environment.Values(string(provider))), nil
}

// Resolve combines values from providers. Later providers take precedence.
func Resolve(providers ...Provider) (map[string]any, error) {
	values := make(map[string]any)

	for _, provider := range providers {
		if provider == nil {
			continue
		}

		resolved, err := provider.Values()
		if err != nil {
			return nil, err
		}

		merge(values, normalize(resolved))
	}

	return values, nil
}

func isYAML(filename string) bool {
	extension := strings.ToLower(path.Ext(filename))

	return extension == ".yaml" || extension == ".yml"
}

func isJSON(filename string) bool {
	return strings.EqualFold(path.Ext(filename), ".json")
}

func parse(filename, contents string) (map[string]any, error) {
	var values map[string]any

	if isJSON(filename) {
		if err := json.Unmarshal([]byte(contents), &values); err != nil {
			return nil, err
		}

		return values, nil
	}

	if err := yaml.Unmarshal([]byte(contents), &values); err != nil {
		return nil, err
	}

	return values, nil
}

func normalize(values map[string]any) map[string]any {
	result := make(map[string]any)

	for key, value := range values {
		parts := strings.Split(key, ".")
		current := result

		for _, part := range parts[:len(parts)-1] {
			if part == "" {
				continue
			}

			next, ok := current[part].(map[string]any)
			if !ok {
				next = make(map[string]any)
				current[part] = next
			}

			current = next
		}

		if nested, ok := value.(map[string]any); ok {
			value = normalize(nested)
		}

		last := parts[len(parts)-1]
		valueMap, valueIsMap := value.(map[string]any)
		existingMap, existingIsMap := current[last].(map[string]any)
		if valueIsMap && existingIsMap {
			merge(existingMap, valueMap)

			continue
		}

		current[last] = value
	}

	return result
}

func merge(destination, source map[string]any) {
	for key, sourceValue := range source {
		sourceMap, sourceIsMap := sourceValue.(map[string]any)
		destinationMap, destinationIsMap := destination[key].(map[string]any)
		if sourceIsMap && destinationIsMap {
			merge(destinationMap, sourceMap)

			continue
		}

		if sourceIsMap {
			destination[key] = maps.Clone(sourceMap)

			continue
		}

		destination[key] = sourceValue
	}
}
