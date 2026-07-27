package configuration_test

import (
	"testing"
	"time"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/rawbytes"
	"github.com/knadh/koanf/v2"
	"github.com/stretchr/testify/require"
	"github.com/studiolambda/cosmos/contract"
	frameworkconfiguration "github.com/studiolambda/cosmos/framework/configuration"
)

type httpServerConfig struct {
	Host string `koanf:"host"`
	Port int    `koanf:"port"`
}

func TestKoanfHasReportsPresence(t *testing.T) {
	t.Parallel()

	driver := newKoanfDriver(t)

	require.True(t, driver.Has("http.server.port"))
	require.False(t, driver.Has("http.server.missing"))
}

func TestKoanfUnmarshalDecodesString(t *testing.T) {
	t.Parallel()

	driver := newKoanfDriver(t)

	var host string
	err := driver.Unmarshal("http.server.host", &host)

	require.NoError(t, err)
	require.Equal(t, "0.0.0.0", host)
}

func TestKoanfUnmarshalDecodesDuration(t *testing.T) {
	t.Parallel()

	driver := newKoanfDriver(t)

	var timeout time.Duration
	err := driver.Unmarshal("http.server.read_timeout", &timeout)

	require.NoError(t, err)
	require.Equal(t, 30*time.Second, timeout)
}

func TestKoanfUnmarshalDecodesTimeWithDefaultLayout(t *testing.T) {
	t.Parallel()

	driver := newKoanfDriver(t)

	var startedAt time.Time
	err := driver.Unmarshal("http.server.started_at", &startedAt)

	require.NoError(t, err)
	require.Equal(t, time.Date(2026, time.July, 27, 10, 30, 0, 0, time.UTC), startedAt)
}

func TestKoanfUnmarshalDecodesTimeWithCustomLayout(t *testing.T) {
	t.Parallel()

	driver := frameworkconfiguration.NewKoanfWith(frameworkconfiguration.KoanfConfig{
		TimeLayout: time.DateOnly,
	})

	err := driver.Instance().Load(rawbytes.Provider([]byte(`
http:
  server:
    started_at: "2026-07-27"
`)), yaml.Parser())
	require.NoError(t, err)

	var startedAt time.Time
	err = driver.Unmarshal("http.server.started_at", &startedAt)

	require.NoError(t, err)
	require.Equal(t, time.Date(2026, time.July, 27, 0, 0, 0, 0, time.UTC), startedAt)
}

func TestNewKoanfFromUsesProvidedInstance(t *testing.T) {
	t.Parallel()

	instance := koanf.New("/")
	driver := frameworkconfiguration.NewKoanfFrom(instance)

	require.Same(t, instance, driver.Instance())
}

func TestKoanfUnmarshalDecodesStruct(t *testing.T) {
	t.Parallel()

	driver := newKoanfDriver(t)

	var config httpServerConfig
	err := driver.Unmarshal("http.server", &config)

	require.NoError(t, err)
	require.Equal(t, "0.0.0.0", config.Host)
	require.Equal(t, 8080, config.Port)
}

func TestKoanfUnmarshalReturnsNotFoundWhenMissing(t *testing.T) {
	t.Parallel()

	driver := newKoanfDriver(t)

	var value string
	err := driver.Unmarshal("http.server.missing", &value)

	require.ErrorIs(t, err, contract.ErrConfigurationKeyNotFound)
}

func TestContractConfigurationGetUsesKoanfDriver(t *testing.T) {
	t.Parallel()

	configuration := contract.NewConfiguration(newKoanfDriver(t))

	port, err := configuration.Get[int]("http.server.port")

	require.NoError(t, err)
	require.Equal(t, 8080, port)
}

func TestContractConfigurationGetOrUsesFallbackWithKoanfDriver(t *testing.T) {
	t.Parallel()

	configuration := contract.NewConfiguration(newKoanfDriver(t))

	port := configuration.GetOr("http.server.missing_port", 3000)

	require.Equal(t, 3000, port)
}

func newKoanfDriver(t *testing.T) *frameworkconfiguration.Koanf {
	t.Helper()

	driver := frameworkconfiguration.NewKoanf()
	err := driver.Instance().Load(rawbytes.Provider([]byte(`
http:
  server:
    host: 0.0.0.0
    port: 8080
    read_timeout: 30s
    started_at: "2026-07-27T10:30:00Z"
`)), yaml.Parser())
	require.NoError(t, err)

	return driver
}
