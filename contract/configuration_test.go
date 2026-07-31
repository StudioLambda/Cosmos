package contract_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/studiolambda/cosmos/contract"
	contractmock "github.com/studiolambda/cosmos/contract/mock"
)

func TestConfigurationGetReturnsDecodedValue(t *testing.T) {
	t.Parallel()

	driver := contractmock.NewConfigurationDriverMock(t)
	driver.On("Delimiter").Return(".")
	driver.On("Unmarshal", "http.port", mock.Anything).Run(func(args mock.Arguments) {
		*args.Get(1).(*int) = 8080
	}).Return(nil)
	configuration := contract.NewConfiguration(driver)

	port, err := configuration.Get[int]("http.port")

	require.NoError(t, err)
	require.Equal(t, 8080, port)
}

func TestConfigurationGetReturnsDriverError(t *testing.T) {
	t.Parallel()

	driver := contractmock.NewConfigurationDriverMock(t)
	driver.On("Delimiter").Return(".")
	driver.On("Unmarshal", "http.port", mock.Anything).Return(errors.New("boom"))
	configuration := contract.NewConfiguration(driver)

	_, err := configuration.Get[int]("http.port")

	require.EqualError(t, err, "boom")
}

func TestConfigurationGetOrReturnsFallbackOnError(t *testing.T) {
	t.Parallel()

	driver := contractmock.NewConfigurationDriverMock(t)
	driver.On("Delimiter").Return(".")
	driver.On("Unmarshal", "http.port", mock.Anything).Return(contract.ErrConfigurationKeyNotFound)
	configuration := contract.NewConfiguration(driver)

	port := configuration.GetOr("http.port", 8080)

	require.Equal(t, 8080, port)
}

func TestConfigurationHasReportsPresence(t *testing.T) {
	t.Parallel()

	driver := contractmock.NewConfigurationDriverMock(t)
	driver.On("Has", "app.name").Return(true)
	driver.On("Has", "app.port").Return(false)
	driver.On("Delimiter").Return(".")
	configuration := contract.NewConfiguration(driver)

	require.True(t, configuration.Has("app.name"))
	require.False(t, configuration.Has("app.port"))
}
