package slog

import (
	"context"
	"fmt"
	"io"
	stdslog "log/slog"
	"os"
	"strings"

	"github.com/lmittmann/tint"
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

// Config configures a slog logger driver.
type Config struct {
	// Level accepts debug, info, warn, or error. The default is info.
	Level string

	// Format accepts json, text, or colored. The default is json.
	Format string

	// Output receives log records. A nil value defaults to os.Stderr.
	// FromConfiguration supports the stdout and stderr output values.
	Output io.Writer
}

// DefaultConfig returns the default slog logger configuration.
func DefaultConfig() Config {
	return Config{
		Level:  "info",
		Format: FormatJSON,
		Output: os.Stderr,
	}
}

// FromConfiguration populates the slog configuration from configuration.
func (config *Config) FromConfiguration(configuration *contract.Configuration) {
	config.Level = configuration.GetOr("level", "")
	config.Format = configuration.GetOr("format", "")
	config.Output = outputFrom(configuration.GetOr("output", ""))
}

func outputFrom(value string) io.Writer {
	switch strings.ToLower(value) {
	case "stdout":
		return os.Stdout
	case "stderr":
		return os.Stderr
	default:
		return nil
	}
}

// Slog implements [contract.LoggerDriver] with [slog.Logger].
type Slog struct {
	logger *stdslog.Logger
}

// New creates a slog logger driver from configuration.
func New(config Config) (*Slog, error) {
	config = config.withDefaults()

	level, err := parseLevel(config.Level)
	if err != nil {
		return nil, err
	}

	options := &stdslog.HandlerOptions{Level: level}

	switch config.Format {
	case FormatJSON:
		return NewSlogFrom(stdslog.New(stdslog.NewJSONHandler(config.Output, options))), nil
	case FormatText:
		return NewSlogFrom(stdslog.New(stdslog.NewTextHandler(config.Output, options))), nil
	case FormatColored:
		return NewSlogFrom(stdslog.New(tint.NewTextHandler(config.Output, &tint.Options{Level: level}))), nil
	default:
		return nil, fmt.Errorf("invalid slog format %q", config.Format)
	}
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

func parseLevel(value string) (stdslog.Level, error) {
	switch strings.ToLower(value) {
	case "debug":
		return stdslog.LevelDebug, nil
	case "info":
		return stdslog.LevelInfo, nil
	case "warn":
		return stdslog.LevelWarn, nil
	case "error":
		return stdslog.LevelError, nil
	default:
		return 0, fmt.Errorf("invalid slog level %q", value)
	}
}

// NewSlogFrom creates a new Slog driver from an existing [slog.Logger].
// A nil logger produces a discard driver.
func NewSlogFrom(logger *stdslog.Logger) *Slog {
	if logger == nil {
		logger = stdslog.New(stdslog.DiscardHandler)
	}

	return &Slog{logger: logger}
}

// DebugContext logs a debug-level message with the given context and attributes.
func (slog *Slog) DebugContext(ctx context.Context, message string, args ...any) {
	slog.logger.DebugContext(ctx, message, args...)
}

// InfoContext logs an info-level message with the given context and attributes.
func (slog *Slog) InfoContext(ctx context.Context, message string, args ...any) {
	slog.logger.InfoContext(ctx, message, args...)
}

// WarnContext logs a warning-level message with the given context and attributes.
func (slog *Slog) WarnContext(ctx context.Context, message string, args ...any) {
	slog.logger.WarnContext(ctx, message, args...)
}

// ErrorContext logs an error-level message with the given context and attributes.
func (slog *Slog) ErrorContext(ctx context.Context, message string, args ...any) {
	slog.logger.ErrorContext(ctx, message, args...)
}

// With returns a derived Slog driver with the given persistent attributes.
func (slog *Slog) With(args ...any) contract.LoggerDriver {
	return NewSlogFrom(slog.logger.With(args...))
}
