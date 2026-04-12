package request_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/studiolambda/cosmos/contract/request"
)

func TestParamReturnsPathValue(t *testing.T) {
	t.Parallel()

	r := httptest.NewRequest(http.MethodGet, "/users/42", nil)
	r.SetPathValue("id", "42")

	result := request.Param(r, "id")

	require.Equal(t, "42", result)
}

func TestParamReturnsEmptyWhenMissing(t *testing.T) {
	t.Parallel()

	r := httptest.NewRequest(http.MethodGet, "/", nil)

	result := request.Param(r, "id")

	require.Equal(t, "", result)
}

func TestParamOrReturnsValueWhenPresent(t *testing.T) {
	t.Parallel()

	r := httptest.NewRequest(http.MethodGet, "/users/42", nil)
	r.SetPathValue("id", "42")

	result := request.ParamOr(r, "id", "0")

	require.Equal(t, "42", result)
}

func TestParamOrReturnsFallbackWhenMissing(t *testing.T) {
	t.Parallel()

	r := httptest.NewRequest(http.MethodGet, "/", nil)

	result := request.ParamOr(r, "id", "default")

	require.Equal(t, "default", result)
}

func TestParamOrReturnsFallbackWhenEmpty(t *testing.T) {
	t.Parallel()

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.SetPathValue("id", "")

	result := request.ParamOr(r, "id", "fallback")

	require.Equal(t, "fallback", result)
}

func TestParamIntReturnsInteger(t *testing.T) {
	t.Parallel()

	r := httptest.NewRequest(http.MethodGet, "/users/42", nil)
	r.SetPathValue("id", "42")

	result, err := request.ParamInt(r, "id")

	require.NoError(t, err)
	require.Equal(t, 42, result)
}

func TestParamIntReturnsErrorWhenEmpty(t *testing.T) {
	t.Parallel()

	r := httptest.NewRequest(http.MethodGet, "/", nil)

	_, err := request.ParamInt(r, "id")

	require.Error(t, err)
	require.Contains(t, err.Error(), "is empty")
}

func TestParamIntReturnsErrorWhenNotInteger(t *testing.T) {
	t.Parallel()

	r := httptest.NewRequest(http.MethodGet, "/users/abc", nil)
	r.SetPathValue("id", "abc")

	_, err := request.ParamInt(r, "id")

	require.Error(t, err)
	require.Contains(t, err.Error(), "not a valid integer")
}

func TestParamIntReturnsNegativeInteger(t *testing.T) {
	t.Parallel()

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.SetPathValue("offset", "-5")

	result, err := request.ParamInt(r, "offset")

	require.NoError(t, err)
	require.Equal(t, -5, result)
}

func TestParamIntReturnsZero(t *testing.T) {
	t.Parallel()

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.SetPathValue("page", "0")

	result, err := request.ParamInt(r, "page")

	require.NoError(t, err)
	require.Equal(t, 0, result)
}

func TestParamIntOrReturnsIntegerWhenValid(t *testing.T) {
	t.Parallel()

	r := httptest.NewRequest(http.MethodGet, "/users/99", nil)
	r.SetPathValue("id", "99")

	result := request.ParamIntOr(r, "id", 0)

	require.Equal(t, 99, result)
}

func TestParamIntOrReturnsFallbackWhenEmpty(t *testing.T) {
	t.Parallel()

	r := httptest.NewRequest(http.MethodGet, "/", nil)

	result := request.ParamIntOr(r, "id", 10)

	require.Equal(t, 10, result)
}

func TestParamIntOrReturnsFallbackWhenNotInteger(t *testing.T) {
	t.Parallel()

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.SetPathValue("id", "abc")

	result := request.ParamIntOr(r, "id", 10)

	require.Equal(t, 10, result)
}
