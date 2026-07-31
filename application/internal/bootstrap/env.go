package bootstrap

import (
	"errors"
	"os"

	"github.com/joho/godotenv"
)

// LoadEnv loads environment variables from filename without overriding values
// already supplied by the process. A missing file is allowed.
func LoadEnv(filename string) error {
	if err := godotenv.Load(filename); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}

	return nil
}
