package bootstrap

import (
	"log/slog"

	"github.com/samber/do/v2"
)

func NewLogger(do do.Injector) (*slog.Logger, error) {
	return slog.Default(), nil
}
