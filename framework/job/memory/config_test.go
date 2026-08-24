package memory

import (
	"testing"
	"time"

	"github.com/studiolambda/cosmos/contract"
	frameworkconfiguration "github.com/studiolambda/cosmos/framework/configuration"
	configuration "github.com/studiolambda/cosmos/framework/configuration/koanf"

	"github.com/stretchr/testify/require"
)

func TestMemoryDriverConfigFromConfiguration(t *testing.T) {
	t.Parallel()

	driver, err := configuration.New(frameworkconfiguration.Map(map[string]any{"jobs": map[string]any{"retry_delay": "3s"}}))
	require.NoError(t, err)

	config := MemoryDriverConfig{}
	config.FromConfiguration(contract.NewConfiguration(driver).Prefixed("jobs"))

	require.Equal(t, MemoryDriverConfig{RetryDelay: 3 * time.Second}, config)

	config = MemoryDriverConfig{}
	config.FromConfiguration(contract.NewConfiguration(driver).Prefixed("missing"))
	require.Equal(t, DefaultMemoryDriverConfig(), config)
}
