// Package mock provides generated mocks for contract interfaces.
//
// The package is intended for unit tests. Mocks expose testify-compatible
// expectation APIs and should not be used in production code paths.
//
// Example
//
//	cacheMock := mock.NewCacheDriverMock(t)
//	cacheMock.On("Get", mock.Anything, "users:1").Return([]byte(`{"id":1}`), nil)
//
//go:generate go tool mockery
package mock
