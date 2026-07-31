package configuration

import (
	"context"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/studiolambda/cosmos/contract"
)

type secretProvider struct {
	ctx    context.Context
	driver contract.SecretDriver
	name   string
	format string
	key    string
}

// JSONSecret creates a provider that parses a JSON object secret as
// configuration values.
func JSONSecret(ctx context.Context, driver contract.SecretDriver, name string) Provider {
	return secretProvider{ctx: ctx, driver: driver, name: name, format: "json"}
}

// YAMLSecret creates a provider that parses a YAML object secret as
// configuration values.
func YAMLSecret(ctx context.Context, driver contract.SecretDriver, name string) Provider {
	return secretProvider{ctx: ctx, driver: driver, name: name, format: "yaml"}
}

// RawSecret creates a provider that stores a UTF-8 secret value at key.
func RawSecret(ctx context.Context, driver contract.SecretDriver, name, key string) Provider {
	return secretProvider{ctx: ctx, driver: driver, name: name, format: "raw", key: key}
}

func (provider secretProvider) Values() (map[string]any, error) {
	if provider.ctx == nil {
		return nil, fmt.Errorf("secret context cannot be nil")
	}

	if provider.driver == nil {
		return nil, fmt.Errorf("secret driver cannot be nil")
	}

	if provider.name == "" {
		return nil, fmt.Errorf("secret name cannot be empty")
	}

	contents, err := provider.driver.Get(provider.ctx, provider.name)
	if err != nil {
		return nil, fmt.Errorf("get secret %q: %w", provider.name, err)
	}

	if provider.format == "raw" {
		if strings.TrimSpace(provider.key) == "" {
			return nil, fmt.Errorf("secret configuration key cannot be empty")
		}

		if !utf8.Valid(contents) {
			return nil, fmt.Errorf("secret %q is not valid UTF-8", provider.name)
		}

		return normalize(map[string]any{provider.key: string(contents)}), nil
	}

	values, err := parse("secret."+provider.format, string(contents))
	if err != nil {
		return nil, fmt.Errorf("parse %s secret %q: %w", provider.format, provider.name, err)
	}

	if values == nil {
		return nil, fmt.Errorf("%s secret %q must contain an object", provider.format, provider.name)
	}

	return normalize(values), nil
}
