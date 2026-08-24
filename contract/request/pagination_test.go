package request_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/studiolambda/cosmos/contract"
	"github.com/studiolambda/cosmos/contract/request"
)

type paginationConfigurationDriver struct {
	values map[string]int
}

func (driver paginationConfigurationDriver) Unmarshal(key string, dest any) error {
	value, ok := driver.values[key]
	if !ok {
		return contract.ErrConfigurationKeyNotFound
	}

	*dest.(*int) = value

	return nil
}

func (paginationConfigurationDriver) Has(string) bool {
	return false
}

func (paginationConfigurationDriver) Delimiter() string {
	return "."
}

func (paginationConfigurationDriver) Extend(...contract.ConfigurationProvider) error {
	return nil
}

func TestPaginationConfigFromConfigurationPreservesDefaults(t *testing.T) {
	t.Parallel()

	configuration := contract.NewConfiguration(paginationConfigurationDriver{})
	config := request.PaginationConfig{}
	config.FromConfiguration(configuration.Prefixed("pagination"))

	require.Equal(t, request.DefaultPaginationConfig(), config)
}

func TestPaginationConfigFromConfigurationOverridesValues(t *testing.T) {
	t.Parallel()

	configuration := contract.NewConfiguration(paginationConfigurationDriver{values: map[string]int{
		"pagination.default_page":     2,
		"pagination.default_per_page": 10,
		"pagination.max_per_page":     50,
	}})
	config := request.PaginationConfig{}
	config.FromConfiguration(configuration.Prefixed("pagination"))

	require.Equal(t, request.PaginationConfig{DefaultPage: 2, DefaultPerPage: 10, MaxPerPage: 50}, config)
}

func TestCursorPaginationConfigFromConfigurationPreservesDefaults(t *testing.T) {
	t.Parallel()

	configuration := contract.NewConfiguration(paginationConfigurationDriver{})
	config := request.CursorPaginationConfig{}
	config.FromConfiguration(configuration.Prefixed("pagination"))

	require.Equal(t, request.DefaultCursorPaginationConfig(), config)
}

func TestCursorPaginationConfigFromConfigurationOverridesValues(t *testing.T) {
	t.Parallel()

	configuration := contract.NewConfiguration(paginationConfigurationDriver{values: map[string]int{
		"pagination.default_per_page": 10,
		"pagination.max_per_page":     50,
	}})
	config := request.CursorPaginationConfig{}
	config.FromConfiguration(configuration.Prefixed("pagination"))

	require.Equal(t, request.CursorPaginationConfig{DefaultPerPage: 10, MaxPerPage: 50}, config)
}

func TestPaginationReturnsDefaults(t *testing.T) {
	t.Parallel()

	r := httptest.NewRequest(http.MethodGet, "/", nil)

	page, perPage := request.Pagination(r, request.DefaultPaginationConfig())

	require.Equal(t, 1, page)
	require.Equal(t, 25, perPage)
}

func TestPaginationParsesQueryParams(t *testing.T) {
	t.Parallel()

	r := httptest.NewRequest(http.MethodGet, "/?page=3&per_page=50", nil)

	page, perPage := request.Pagination(r, request.DefaultPaginationConfig())

	require.Equal(t, 3, page)
	require.Equal(t, 50, perPage)
}

func TestPaginationClampsPerPageToMax(t *testing.T) {
	t.Parallel()

	r := httptest.NewRequest(http.MethodGet, "/?per_page=999", nil)

	_, perPage := request.Pagination(r, request.DefaultPaginationConfig())

	require.Equal(t, 100, perPage)
}

func TestPaginationClampsPageBelowOne(t *testing.T) {
	t.Parallel()

	r := httptest.NewRequest(http.MethodGet, "/?page=0", nil)

	page, _ := request.Pagination(r, request.DefaultPaginationConfig())

	require.Equal(t, 1, page)
}

func TestPaginationClampsNegativePage(t *testing.T) {
	t.Parallel()

	r := httptest.NewRequest(http.MethodGet, "/?page=-5", nil)

	page, _ := request.Pagination(r, request.DefaultPaginationConfig())

	require.Equal(t, 1, page)
}

func TestPaginationClampsPerPageBelowOne(t *testing.T) {
	t.Parallel()

	r := httptest.NewRequest(http.MethodGet, "/?per_page=0", nil)

	_, perPage := request.Pagination(r, request.DefaultPaginationConfig())

	require.Equal(t, 1, perPage)
}

func TestPaginationWithCustomDefaults(t *testing.T) {
	t.Parallel()

	r := httptest.NewRequest(http.MethodGet, "/", nil)

	page, perPage := request.PaginationWith(r, 2, 10, 50)

	require.Equal(t, 2, page)
	require.Equal(t, 10, perPage)
}

func TestPaginationWithCustomMax(t *testing.T) {
	t.Parallel()

	r := httptest.NewRequest(http.MethodGet, "/?per_page=100", nil)

	_, perPage := request.PaginationWith(r, 1, 10, 50)

	require.Equal(t, 50, perPage)
}

func TestPaginationIgnoresInvalidPage(t *testing.T) {
	t.Parallel()

	r := httptest.NewRequest(http.MethodGet, "/?page=abc", nil)

	page, _ := request.Pagination(r, request.DefaultPaginationConfig())

	require.Equal(t, 1, page)
}

func TestPaginationIgnoresInvalidPerPage(t *testing.T) {
	t.Parallel()

	r := httptest.NewRequest(http.MethodGet, "/?per_page=abc", nil)

	_, perPage := request.Pagination(r, request.DefaultPaginationConfig())

	require.Equal(t, 25, perPage)
}

func TestCursorPaginationReturnsDefaults(t *testing.T) {
	t.Parallel()

	r := httptest.NewRequest(http.MethodGet, "/", nil)

	cursor, perPage := request.CursorPagination(r, request.DefaultCursorPaginationConfig())

	require.Empty(t, cursor)
	require.Equal(t, 25, perPage)
}

func TestCursorPaginationParsesCursor(t *testing.T) {
	t.Parallel()

	r := httptest.NewRequest(http.MethodGet, "/?cursor=abc123&per_page=50", nil)

	cursor, perPage := request.CursorPagination(r, request.DefaultCursorPaginationConfig())

	require.Equal(t, "abc123", cursor)
	require.Equal(t, 50, perPage)
}

func TestCursorPaginationClampsPerPageToMax(t *testing.T) {
	t.Parallel()

	r := httptest.NewRequest(http.MethodGet, "/?per_page=999", nil)

	_, perPage := request.CursorPagination(r, request.DefaultCursorPaginationConfig())

	require.Equal(t, 100, perPage)
}

func TestCursorPaginationClampsPerPageBelowOne(t *testing.T) {
	t.Parallel()

	r := httptest.NewRequest(http.MethodGet, "/?per_page=0", nil)

	_, perPage := request.CursorPagination(r, request.DefaultCursorPaginationConfig())

	require.Equal(t, 1, perPage)
}

func TestCursorPaginationWithCustomDefaults(t *testing.T) {
	t.Parallel()

	r := httptest.NewRequest(http.MethodGet, "/", nil)

	_, perPage := request.CursorPaginationWith(r, 10, 50)

	require.Equal(t, 10, perPage)
}

func TestCursorPaginationWithCustomMax(t *testing.T) {
	t.Parallel()

	r := httptest.NewRequest(http.MethodGet, "/?per_page=100", nil)

	_, perPage := request.CursorPaginationWith(r, 10, 50)

	require.Equal(t, 50, perPage)
}

func TestCursorPaginationIgnoresInvalidPerPage(t *testing.T) {
	t.Parallel()

	r := httptest.NewRequest(http.MethodGet, "/?per_page=abc", nil)

	_, perPage := request.CursorPagination(r, request.DefaultCursorPaginationConfig())

	require.Equal(t, 25, perPage)
}

func TestCursorPaginationNegativePerPage(t *testing.T) {
	t.Parallel()

	r := httptest.NewRequest(http.MethodGet, "/?per_page=-5", nil)

	_, perPage := request.CursorPagination(r, request.DefaultCursorPaginationConfig())

	require.Equal(t, 1, perPage)
}
