package framework_test

import (
	"errors"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/studiolambda/cosmos/contract"
	"github.com/studiolambda/cosmos/framework"
	cachememory "github.com/studiolambda/cosmos/framework/cache/memory"
	cryptoaes "github.com/studiolambda/cosmos/framework/crypto/aes"
	"github.com/studiolambda/cosmos/framework/database/sqlite"
	eventmemory "github.com/studiolambda/cosmos/framework/event/memory"
	loggerslog "github.com/studiolambda/cosmos/framework/logger/slog"
	"github.com/studiolambda/cosmos/framework/middleware"

	"github.com/stretchr/testify/require"
)

type configurationDriver map[string]any

func (driver configurationDriver) Unmarshal(key string, dest any) error {
	value, ok := driver[key]
	if !ok {
		return contract.ErrConfigurationKeyNotFound
	}

	destination := reflect.ValueOf(dest)
	if destination.Kind() != reflect.Pointer || destination.IsNil() {
		return errors.New("configuration destination must be a non-nil pointer")
	}

	valueOf := reflect.ValueOf(value)
	if !valueOf.Type().AssignableTo(destination.Elem().Type()) {
		return errors.New("configuration value has an incompatible type")
	}

	destination.Elem().Set(valueOf)

	return nil
}

func (driver configurationDriver) Has(key string) bool {
	_, ok := driver[key]

	return ok
}

func (configurationDriver) Delimiter() string {
	return "."
}

func (configurationDriver) Extend(...contract.ConfigurationProvider) error {
	return nil
}

func TestFromConfigurationUsesDefaultsWhenEmpty(t *testing.T) {
	t.Parallel()

	configuration := contract.NewConfiguration(configurationDriver{})

	require.Equal(t, middleware.DefaultCORSConfig(), configuration.From[middleware.CORSConfig]("cors"))
	require.Equal(t, middleware.DefaultSecureHeadersConfig(), configuration.From[middleware.SecureHeadersConfig]("secure_headers"))
	require.Equal(t, middleware.DefaultCorrelationConfig(), configuration.From[middleware.CorrelationConfig]("correlation"))
	require.Equal(t, middleware.DefaultRateLimitConfig(), configuration.From[middleware.RateLimitConfig]("rate_limit"))
	require.Equal(t, middleware.DefaultCSRFConfig(), configuration.From[middleware.CSRFConfig]("csrf"))
	require.Equal(t, framework.DefaultServerConfig(), configuration.From[framework.ServerConfig]("server"))
	require.Equal(t, cachememory.DefaultMemoryConfig(), configuration.From[cachememory.MemoryConfig]("cache"))
	require.Equal(t, eventmemory.DefaultMemoryBrokerConfig(), configuration.From[eventmemory.MemoryBrokerConfig]("event"))
	require.Equal(t, loggerslog.DefaultConfig(), configuration.From[loggerslog.Config]("logger"))
}

func TestFromConfigurationOverridesPrefixedValues(t *testing.T) {
	t.Parallel()

	configuration := contract.NewConfiguration(configurationDriver{
		"cors.allowed_origins":            []string{"https://example.com"},
		"secure_headers.frame_options":    "SAMEORIGIN",
		"correlation.problem_key":         "request_id",
		"rate_limit.limit":                42,
		"csrf.trusted_origins":            []string{"https://trusted.example.com"},
		"server.port":                     9090,
		"cache.cleanup":                   time.Minute,
		"aes.key":                         []byte("12345678901234567890123456789012"),
		"sqlite.dsn":                      "file:configured.db",
		"event.max_concurrent_deliveries": 8,
		"logger.level":                    "debug",
		"logger.output":                   "stdout",
	})

	require.Equal(t, []string{"https://example.com"}, configuration.From[middleware.CORSConfig]("cors").AllowedOrigins)
	require.Equal(t, "SAMEORIGIN", configuration.From[middleware.SecureHeadersConfig]("secure_headers").FrameOptions)
	require.Equal(t, "request_id", configuration.From[middleware.CorrelationConfig]("correlation").ProblemKey)
	require.Equal(t, 42, configuration.From[middleware.RateLimitConfig]("rate_limit").Limit)
	require.Equal(t, []string{"https://trusted.example.com"}, configuration.From[middleware.CSRFConfig]("csrf").TrustedOrigins)
	require.Equal(t, 9090, configuration.From[framework.ServerConfig]("server").Port)
	require.Equal(t, time.Minute, configuration.From[cachememory.MemoryConfig]("cache").Cleanup)
	require.Equal(t, []byte("12345678901234567890123456789012"), configuration.From[cryptoaes.AESConfig]("aes").Key)
	require.Equal(t, "file:configured.db", configuration.From[sqlite.Config]("sqlite").DSN)
	require.Equal(t, 8, configuration.From[eventmemory.MemoryBrokerConfig]("event").MaxConcurrentDeliveries)
	require.Equal(t, "debug", configuration.From[loggerslog.Config]("logger").Level)
	require.Same(t, os.Stdout, configuration.From[loggerslog.Config]("logger").Output)
}

func TestFromConfigurationPreservesMandatoryConfigurationFields(t *testing.T) {
	t.Parallel()

	configuration := contract.NewConfiguration(configurationDriver{
		"aes.key":               []byte("12345678901234567890123456789012"),
		"sqlite.max_open_conns": 10,
	})

	aesConfig := cryptoaes.AESConfig{Key: []byte("existing-key")}
	aesConfig.FromConfiguration(configuration.Prefixed("missing"))
	require.Equal(t, []byte("existing-key"), aesConfig.Key)

	aesConfig.FromConfiguration(configuration.Prefixed("aes"))
	require.Equal(t, []byte("12345678901234567890123456789012"), aesConfig.Key)

	sqliteConfig := sqlite.Config{
		DSN:             "file:existing.db",
		MaxOpenConns:    5,
		MaxIdleConns:    4,
		ConnMaxLifetime: time.Minute,
		ConnMaxIdleTime: 30 * time.Second,
	}
	sqliteConfig.FromConfiguration(configuration.Prefixed("sqlite"))
	require.Equal(t, "file:existing.db", sqliteConfig.DSN)
	require.Equal(t, 10, sqliteConfig.MaxOpenConns)
	require.Equal(t, 4, sqliteConfig.MaxIdleConns)
	require.Equal(t, time.Minute, sqliteConfig.ConnMaxLifetime)
	require.Equal(t, 30*time.Second, sqliteConfig.ConnMaxIdleTime)

	logger := contract.NewLogger(nil)
	brokerConfig := eventmemory.MemoryBrokerConfig{Logger: logger}
	brokerConfig.FromConfiguration(configuration.Prefixed("event"))
	require.Same(t, logger, brokerConfig.Logger)
}
