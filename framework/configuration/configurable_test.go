package configuration_test

import (
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/studiolambda/cosmos/contract"
	"github.com/studiolambda/cosmos/framework/cache/redis"
	"github.com/studiolambda/cosmos/framework/configuration"
	"github.com/studiolambda/cosmos/framework/configuration/koanf"
	"github.com/studiolambda/cosmos/framework/crypto/chacha20"
	"github.com/studiolambda/cosmos/framework/database/mysql"
	"github.com/studiolambda/cosmos/framework/database/postgres"
	"github.com/studiolambda/cosmos/framework/hash/argon2"
	"github.com/studiolambda/cosmos/framework/hash/bcrypt"
	zaplogger "github.com/studiolambda/cosmos/framework/logger/zap"
	"github.com/studiolambda/cosmos/framework/logger/zerolog"
	"github.com/studiolambda/cosmos/framework/secret/awssm"
	"github.com/studiolambda/cosmos/framework/secret/azurekeyvault"
	"github.com/studiolambda/cosmos/framework/secret/vault"
	"github.com/studiolambda/cosmos/framework/session"
)

func TestConfigFromConfigurationPreservesDefaults(t *testing.T) {
	t.Parallel()

	configuration := newConfiguration(t, nil)

	var redisConfig redis.RedisConfig
	redisConfig.FromConfiguration(configuration.Prefixed("redis"))
	require.Equal(t, redis.DefaultRedisConfig(), redisConfig)

	key := []byte("required-key")
	chachaConfig := chacha20.ChaCha20Config{Key: key}
	chachaConfig.FromConfiguration(configuration.Prefixed("chacha"))
	require.Equal(t, key, chachaConfig.Key)

	argon2Config := argon2.Argon2Config{}
	argon2Config.FromConfiguration(configuration.Prefixed("argon2"))
	require.Equal(t, argon2.DefaultArgon2Config(), argon2Config)

	bcryptConfig := bcrypt.BcryptConfig{}
	bcryptConfig.FromConfiguration(configuration.Prefixed("bcrypt"))
	require.Equal(t, bcrypt.DefaultBcryptConfig(), bcryptConfig)

	var output bytes.Buffer
	zapConfig := zaplogger.Config{Output: &output}
	zapConfig.FromConfiguration(configuration.Prefixed("zap"))
	require.Equal(t, zaplogger.DefaultConfig().Level, zapConfig.Level)
	require.Same(t, &output, zapConfig.Output)

	zerologConfig := zerolog.Config{Output: &output}
	zerologConfig.FromConfiguration(configuration.Prefixed("zerolog"))
	require.Equal(t, zerolog.DefaultConfig().Format, zerologConfig.Format)
	require.Same(t, &output, zerologConfig.Output)

	cacheConfig := session.CacheDriverConfig{}
	cacheConfig.FromConfiguration(configuration.Prefixed("session"))
	require.Equal(t, session.DefaultCacheDriverConfig(), cacheConfig)

	mysqlConfig := mysql.Config{DSN: "required"}
	mysqlConfig.FromConfiguration(configuration.Prefixed("mysql"))
	require.Equal(t, "required", mysqlConfig.DSN)

	postgresConfig := postgres.Config{DSN: "required"}
	postgresConfig.FromConfiguration(configuration.Prefixed("postgres"))
	require.Equal(t, "required", postgresConfig.DSN)

	vaultConfig := vault.Config{}
	vaultConfig.FromConfiguration(configuration.Prefixed("vault"))
	require.Equal(t, vault.DefaultConfig(), vaultConfig)
}

func TestConfigFromConfigurationOverridesValues(t *testing.T) {
	t.Parallel()

	configuration := newConfiguration(t, map[string]any{
		"redis.network": "unix", "redis.addr": "/tmp/redis.sock", "redis.db": 2,
		"chacha.key":       []byte("configured-key"),
		"argon2.time_cost": 4, "bcrypt.cost": 13,
		"zap.level": "debug", "zap.format": "text", "zerolog.level": "error", "zerolog.format": "colored",
		"session.prefix": "app:sessions",
		"mysql.dsn":      "mysql://configured", "mysql.max_open_conns": 5, "mysql.conn_max_lifetime": "1m",
		"postgres.dsn": "postgres://configured", "postgres.max_idle_conns": 3, "postgres.conn_max_idle_time": "2m",
		"awssm.region": "eu-west-1", "awssm.endpoint": "http://localhost", "awssm.name": "app",
		"azure.vault_url": "https://example.vault.azure.net",
		"vault.address":   "http://vault:8200", "vault.token": "token", "vault.mount": "app",
	})

	var redisConfig redis.RedisConfig
	redisConfig.FromConfiguration(configuration.Prefixed("redis"))
	require.Equal(t, "unix", redisConfig.Network)
	require.Equal(t, 2, redisConfig.DB)

	var chachaConfig chacha20.ChaCha20Config
	chachaConfig.FromConfiguration(configuration.Prefixed("chacha"))
	require.Equal(t, []byte("configured-key"), chachaConfig.Key)

	var argon2Config argon2.Argon2Config
	argon2Config.FromConfiguration(configuration.Prefixed("argon2"))
	require.EqualValues(t, 4, argon2Config.TimeCost)

	var bcryptConfig bcrypt.BcryptConfig
	bcryptConfig.FromConfiguration(configuration.Prefixed("bcrypt"))
	require.Equal(t, 13, bcryptConfig.Cost)

	var output bytes.Buffer
	zapConfig := zaplogger.Config{Output: &output}
	zapConfig.FromConfiguration(configuration.Prefixed("zap"))
	require.Equal(t, "debug", zapConfig.Level)
	require.Equal(t, zaplogger.FormatText, zapConfig.Format)
	require.Same(t, &output, zapConfig.Output)

	zerologConfig := zerolog.Config{Output: &output}
	zerologConfig.FromConfiguration(configuration.Prefixed("zerolog"))
	require.Equal(t, "error", zerologConfig.Level)
	require.Equal(t, zerolog.FormatColored, zerologConfig.Format)
	require.Same(t, &output, zerologConfig.Output)

	var cacheConfig session.CacheDriverConfig
	cacheConfig.FromConfiguration(configuration.Prefixed("session"))
	require.Equal(t, "app:sessions", cacheConfig.Prefix)

	var mysqlConfig mysql.Config
	mysqlConfig.FromConfiguration(configuration.Prefixed("mysql"))
	require.Equal(t, "mysql://configured", mysqlConfig.DSN)
	require.Equal(t, 5, mysqlConfig.MaxOpenConns)
	require.Equal(t, time.Minute, mysqlConfig.ConnMaxLifetime)

	var postgresConfig postgres.Config
	postgresConfig.FromConfiguration(configuration.Prefixed("postgres"))
	require.Equal(t, "postgres://configured", postgresConfig.DSN)
	require.Equal(t, 3, postgresConfig.MaxIdleConns)
	require.Equal(t, 2*time.Minute, postgresConfig.ConnMaxIdleTime)

	var awsConfig awssm.Config
	awsConfig.FromConfiguration(configuration.Prefixed("awssm"))
	require.Equal(t, awssm.Config{Region: "eu-west-1", Endpoint: "http://localhost", Name: "app"}, awsConfig)

	var azureConfig azurekeyvault.Config
	azureConfig.FromConfiguration(configuration.Prefixed("azure"))
	require.Equal(t, "https://example.vault.azure.net", azureConfig.VaultURL)

	var vaultConfig vault.Config
	vaultConfig.FromConfiguration(configuration.Prefixed("vault"))
	require.Equal(t, vault.Config{Address: "http://vault:8200", Token: "token", Mount: "app"}, vaultConfig)
}

func newConfiguration(t *testing.T, values map[string]any) *contract.Configuration {
	t.Helper()

	driver, err := koanf.New(configuration.Map(values))
	require.NoError(t, err)

	return contract.NewConfiguration(driver)
}
