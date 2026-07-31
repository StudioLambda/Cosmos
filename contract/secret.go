package contract

import (
	"context"
	"encoding/json/v2"
	"errors"
	"fmt"
)

var ErrSecretNotFound = errors.New("secret not found")

// SecretDriver retrieves raw secret payloads by name.
type SecretDriver interface {
	// Get retrieves the raw value of name. It returns [ErrSecretNotFound] when
	// the named secret does not exist.
	Get(ctx context.Context, name string) ([]byte, error)
}

// Secrets provides typed secret retrieval over a [SecretDriver].
type Secrets struct {
	driver SecretDriver
}

// NewSecrets creates a new Secrets that delegates to driver.
func NewSecrets(driver SecretDriver) *Secrets {
	return &Secrets{driver: driver}
}

// Driver returns the underlying [SecretDriver].
func (secrets *Secrets) Driver() SecretDriver {
	return secrets.driver
}

// Raw retrieves the unmodified bytes of name.
func (secrets *Secrets) Raw(ctx context.Context, name string) ([]byte, error) {
	return secrets.driver.Get(ctx, name)
}

// String retrieves name as a UTF-8 string.
func (secrets *Secrets) String(ctx context.Context, name string) (string, error) {
	value, err := secrets.Raw(ctx, name)
	if err != nil {
		return "", err
	}

	return string(value), nil
}

// Get retrieves name and JSON-decodes it into T.
func (secrets *Secrets) Get[T any](ctx context.Context, name string) (res T, err error) {
	value, err := secrets.Raw(ctx, name)
	if err != nil {
		return res, err
	}

	if err := json.Unmarshal(value, &res); err != nil {
		return res, fmt.Errorf("decode secret %q: %w", name, err)
	}

	return res, nil
}
