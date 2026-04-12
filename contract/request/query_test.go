package request_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/studiolambda/cosmos/contract/request"
)

func TestQueryReturnsValue(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/?name=alice", nil)

	result := request.Query(r, "name")

	require.Equal(t, "alice", result)
}

func TestQueryReturnsEmptyWhenMissing(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)

	result := request.Query(r, "name")

	require.Equal(t, "", result)
}

func TestQueryReturnsEmptyValueWhenPresentButEmpty(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/?name=", nil)

	result := request.Query(r, "name")

	require.Equal(t, "", result)
}

func TestHasQueryReturnsTrueWhenPresent(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/?active=true", nil)

	result := request.HasQuery(r, "active")

	require.True(t, result)
}

func TestHasQueryReturnsTrueWhenPresentButEmpty(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/?active=", nil)

	result := request.HasQuery(r, "active")

	require.True(t, result)
}

func TestHasQueryReturnsTrueWhenPresentNoValue(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/?active", nil)

	result := request.HasQuery(r, "active")

	require.True(t, result)
}

func TestHasQueryReturnsFalseWhenMissing(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)

	result := request.HasQuery(r, "active")

	require.False(t, result)
}

func TestQueryOrReturnsValueWhenPresent(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/?page=2", nil)

	result := request.QueryOr(r, "page", "1")

	require.Equal(t, "2", result)
}

func TestQueryOrReturnsEmptyValueWhenParamExists(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/?page=", nil)

	result := request.QueryOr(r, "page", "1")

	require.Equal(t, "", result)
}

func TestQueryOrReturnsFallbackWhenMissing(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)

	result := request.QueryOr(r, "page", "1")

	require.Equal(t, "1", result)
}

func TestQueryIntReturnsInteger(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/?page=5", nil)

	result, err := request.QueryInt(r, "page")

	require.NoError(t, err)
	require.Equal(t, 5, result)
}

func TestQueryIntReturnsErrorWhenEmpty(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)

	_, err := request.QueryInt(r, "page")

	require.Error(t, err)
	require.Contains(t, err.Error(), "is empty")
}

func TestQueryIntReturnsErrorWhenNotInteger(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/?page=abc", nil)

	_, err := request.QueryInt(r, "page")

	require.Error(t, err)
	require.Contains(t, err.Error(), "not a valid integer")
}

func TestQueryIntReturnsNegativeInteger(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/?offset=-3", nil)

	result, err := request.QueryInt(r, "offset")

	require.NoError(t, err)
	require.Equal(t, -3, result)
}

func TestQueryIntReturnsZero(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/?page=0", nil)

	result, err := request.QueryInt(r, "page")

	require.NoError(t, err)
	require.Equal(t, 0, result)
}

func TestQueryIntOrReturnsIntegerWhenValid(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/?page=7", nil)

	result := request.QueryIntOr(r, "page", 1)

	require.Equal(t, 7, result)
}

func TestQueryIntOrReturnsFallbackWhenEmpty(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)

	result := request.QueryIntOr(r, "page", 1)

	require.Equal(t, 1, result)
}

func TestQueryIntOrReturnsFallbackWhenNotInteger(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/?page=abc", nil)

	result := request.QueryIntOr(r, "page", 1)

	require.Equal(t, 1, result)
}

func TestQueryOrParsesOnceReturnsValueWhenPresent(t *testing.T) {
	t.Parallel()

	r := httptest.NewRequest(http.MethodGet, "/?key=value", nil)

	result := request.QueryOr(r, "key", "default")

	require.Equal(t, "value", result)
}

func TestQueryOrParsesOnceReturnsFallbackWhenMissing(t *testing.T) {
	t.Parallel()

	r := httptest.NewRequest(http.MethodGet, "/", nil)

	result := request.QueryOr(r, "key", "default")

	require.Equal(t, "default", result)
}

func TestQueryOrParsesOnceReturnsEmptyWhenPresentButEmpty(t *testing.T) {
	t.Parallel()

	r := httptest.NewRequest(http.MethodGet, "/?key=", nil)

	result := request.QueryOr(r, "key", "default")

	require.Equal(t, "", result)
}
