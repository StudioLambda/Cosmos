package request_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/studiolambda/cosmos/contract/request"
)

func TestHeaderReturnsValue(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("X-Custom", "value")

	result := request.Header(r, "X-Custom")

	require.Equal(t, "value", result)
}

func TestHeaderReturnsEmptyWhenMissing(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)

	result := request.Header(r, "X-Missing")

	require.Equal(t, "", result)
}

func TestHasHeaderReturnsTrueWhenPresent(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("Authorization", "Bearer token")

	result := request.HasHeader(r, "Authorization")

	require.True(t, result)
}

func TestHasHeaderReturnsFalseWhenMissing(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)

	result := request.HasHeader(r, "Authorization")

	require.False(t, result)
}

func TestHasHeaderReturnsFalseWhenEmpty(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("X-Empty", "")

	result := request.HasHeader(r, "X-Empty")

	require.False(t, result)
}

func TestHeaderOrReturnsValueWhenPresent(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("Accept", "application/json")

	result := request.HeaderOr(r, "Accept", "text/html")

	require.Equal(t, "application/json", result)
}

func TestHeaderOrReturnsFallbackWhenMissing(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)

	result := request.HeaderOr(r, "Accept", "text/html")

	require.Equal(t, "text/html", result)
}

func TestHeaderOrReturnsFallbackWhenEmpty(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("Accept", "")

	result := request.HeaderOr(r, "Accept", "text/html")

	require.Equal(t, "text/html", result)
}

func TestHeaderValuesReturnsAllValues(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Add("X-Multi", "first")
	r.Header.Add("X-Multi", "second")

	result := request.HeaderValues(r, "X-Multi")

	require.Equal(t, []string{"first", "second"}, result)
}

func TestHeaderValuesReturnsNilWhenMissing(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)

	result := request.HeaderValues(r, "X-Missing")

	require.Nil(t, result)
}
