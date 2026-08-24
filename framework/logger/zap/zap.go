package zap

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/studiolambda/cosmos/contract"
	uberzap "go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	// FormatJSON writes structured JSON records.
	FormatJSON = "json"

	// FormatText writes plain text records.
	FormatText = "text"

	// FormatColored writes ANSI-colored text records.
	FormatColored = "colored"
)

// Config configures a Zap logger driver.
type Config struct {
	// Level accepts debug, info, warn, or error. The default is info.
	Level string

	// Format accepts json, text, or colored. The default is json.
	Format string

	// Output receives log records. A nil value defaults to os.Stderr.
	Output io.Writer
}

// DefaultConfig returns the default Zap logger configuration.
func DefaultConfig() Config {
	return Config{
		Level:  "info",
		Format: FormatJSON,
		Output: os.Stderr,
	}
}

// FromConfiguration populates the Zap configuration from configuration.
// Output is intentionally retained because writers are runtime objects.
func (config *Config) FromConfiguration(configuration *contract.Configuration) {
	output := config.Output
	*config = DefaultConfig()
	config.Output = output
	config.Level = configuration.GetOr("level", config.Level)
	config.Format = configuration.GetOr("format", config.Format)
}

// Zap implements [contract.LoggerDriver] with [zap.SugaredLogger].
type Zap struct {
	logger *uberzap.SugaredLogger
}

// New creates a Zap logger driver from configuration.
func New(config Config) (*Zap, error) {
	config = config.withDefaults()

	level, err := parseLevel(config.Level)
	if err != nil {
		return nil, err
	}

	encoderConfig := uberzap.NewProductionEncoderConfig()
	var encoder zapcore.Encoder

	switch config.Format {
	case FormatJSON:
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	case FormatText:
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	case FormatColored:
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	default:
		return nil, fmt.Errorf("invalid zap format %q", config.Format)
	}

	core := zapcore.NewCore(encoder, zapcore.AddSync(config.Output), level)

	return NewZapFrom(uberzap.New(core)), nil
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

func parseLevel(value string) (zapcore.Level, error) {
	var level zapcore.Level

	if err := level.Set(strings.ToLower(value)); err != nil {
		return 0, fmt.Errorf("invalid zap level %q: %w", value, err)
	}

	return level, nil
}

// NewZapFrom creates a new Zap driver from an existing [zap.Logger].
// A nil logger produces a discard driver.
func NewZapFrom(logger *uberzap.Logger) *Zap {
	if logger == nil {
		logger = uberzap.NewNop()
	}

	return &Zap{logger: logger.Sugar()}
}

// DebugContext logs a debug-level message with the given context and attributes.
// Zap does not natively propagate contexts.
func (zap *Zap) DebugContext(_ context.Context, message string, args ...any) {
	zap.logger.Debugw(message, args...)
}

// InfoContext logs an info-level message with the given context and attributes.
// Zap does not natively propagate contexts.
func (zap *Zap) InfoContext(_ context.Context, message string, args ...any) {
	zap.logger.Infow(message, args...)
}

// WarnContext logs a warning-level message with the given context and attributes.
// Zap does not natively propagate contexts.
func (zap *Zap) WarnContext(_ context.Context, message string, args ...any) {
	zap.logger.Warnw(message, args...)
}

// ErrorContext logs an error-level message with the given context and attributes.
// Zap does not natively propagate contexts.
func (zap *Zap) ErrorContext(_ context.Context, message string, args ...any) {
	zap.logger.Errorw(message, args...)
}

// With returns a derived Zap driver with the given persistent attributes.
func (zap *Zap) With(args ...any) contract.LoggerDriver {
	return &Zap{logger: zap.logger.With(args...)}
}
