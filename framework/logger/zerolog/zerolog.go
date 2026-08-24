package zerolog

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	rszerolog "github.com/rs/zerolog"
	"github.com/studiolambda/cosmos/contract"
)

const (
	// FormatJSON writes structured JSON records.
	FormatJSON = "json"

	// FormatText writes plain text records.
	FormatText = "text"

	// FormatColored writes ANSI-colored text records.
	FormatColored = "colored"
)

// Config configures a Zerolog logger driver.
type Config struct {
	// Level accepts debug, info, warn, or error. The default is info.
	Level string

	// Format accepts json, text, or colored. The default is json.
	Format string

	// Output receives log records. A nil value defaults to os.Stderr.
	Output io.Writer
}

// DefaultConfig returns the default Zerolog logger configuration.
func DefaultConfig() Config {
	return Config{
		Level:  "info",
		Format: FormatJSON,
		Output: os.Stderr,
	}
}

// FromConfiguration populates the Zerolog configuration from configuration.
// Output is intentionally retained because writers are runtime objects.
func (config *Config) FromConfiguration(configuration *contract.Configuration) {
	output := config.Output
	*config = DefaultConfig()
	config.Output = output
	config.Level = configuration.GetOr("level", config.Level)
	config.Format = configuration.GetOr("format", config.Format)
}

// Zerolog implements [contract.LoggerDriver] with [zerolog.Logger].
type Zerolog struct {
	logger rszerolog.Logger
}

// New creates a Zerolog logger driver from configuration.
func New(config Config) (*Zerolog, error) {
	config = config.withDefaults()

	level, err := parseLevel(config.Level)
	if err != nil {
		return nil, err
	}

	var output io.Writer = config.Output

	switch config.Format {
	case FormatJSON:
	case FormatText:
		output = rszerolog.ConsoleWriter{Out: config.Output, NoColor: true}
	case FormatColored:
		output = rszerolog.ConsoleWriter{Out: config.Output}
	default:
		return nil, fmt.Errorf("invalid zerolog format %q", config.Format)
	}

	logger := rszerolog.New(output).Level(level).With().Timestamp().Logger()

	return NewZerologFrom(logger), nil
}

func (config Config) withDefaults() Config {
	defaults := DefaultConfig()

	if config.Level == "" {
		config.Level = defaults.Level
	}

	if config.Format == "" {
		config.Format = defaults.Format
	}

	if config.Output == nil {
		config.Output = defaults.Output
	}

	return config
}

func parseLevel(value string) (rszerolog.Level, error) {
	switch strings.ToLower(value) {
	case "debug":
		return rszerolog.DebugLevel, nil
	case "info":
		return rszerolog.InfoLevel, nil
	case "warn":
		return rszerolog.WarnLevel, nil
	case "error":
		return rszerolog.ErrorLevel, nil
	default:
		return rszerolog.NoLevel, fmt.Errorf("invalid zerolog level %q", value)
	}
}

// NewZerologFrom creates a new Zerolog driver from an existing [zerolog.Logger].
func NewZerologFrom(logger rszerolog.Logger) *Zerolog {
	return &Zerolog{logger: logger}
}

// DebugContext logs a debug-level message with the given context and attributes.
func (zerolog *Zerolog) DebugContext(ctx context.Context, message string, args ...any) {
	zerolog.logger.Debug().Ctx(ctx).Fields(args).Msg(message)
}

// InfoContext logs an info-level message with the given context and attributes.
func (zerolog *Zerolog) InfoContext(ctx context.Context, message string, args ...any) {
	zerolog.logger.Info().Ctx(ctx).Fields(args).Msg(message)
}

// WarnContext logs a warning-level message with the given context and attributes.
func (zerolog *Zerolog) WarnContext(ctx context.Context, message string, args ...any) {
	zerolog.logger.Warn().Ctx(ctx).Fields(args).Msg(message)
}

// ErrorContext logs an error-level message with the given context and attributes.
func (zerolog *Zerolog) ErrorContext(ctx context.Context, message string, args ...any) {
	zerolog.logger.Error().Ctx(ctx).Fields(args).Msg(message)
}

// With returns a derived Zerolog driver with the given persistent attributes.
func (zerolog *Zerolog) With(args ...any) contract.LoggerDriver {
	return NewZerologFrom(zerolog.logger.With().Fields(args).Logger())
}
