package request_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/studiolambda/cosmos/contract/request"
)

func TestCookieReturnsMatchingCookie(t *testing.T) {
	t.Parallel()

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.AddCookie(&http.Cookie{Name: "session", Value: "abc123"})

	cookie := request.Cookie(r, "session")

	require.NotNil(t, cookie)
	require.Equal(t, "session", cookie.Name)
	require.Equal(t, "abc123", cookie.Value)
}

func TestCookieReturnsNilWhenNotFound(t *testing.T) {
	t.Parallel()

	r := httptest.NewRequest(http.MethodGet, "/", nil)

	cookie := request.Cookie(r, "missing")

	require.Nil(t, cookie)
}

func TestCookieValueReturnsValue(t *testing.T) {
	t.Parallel()

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.AddCookie(&http.Cookie{Name: "token", Value: "xyz"})

	value := request.CookieValue(r, "token")

	require.Equal(t, "xyz", value)
}

func TestCookieValueReturnsEmptyWhenNotFound(t *testing.T) {
	t.Parallel()

	r := httptest.NewRequest(http.MethodGet, "/", nil)

	value := request.CookieValue(r, "missing")

	require.Equal(t, "", value)
}

func TestCookieValueOrReturnsValueWhenPresent(t *testing.T) {
	t.Parallel()

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.AddCookie(&http.Cookie{Name: "lang", Value: "en"})

	value := request.CookieValueOr(r, "lang", "fr")

	require.Equal(t, "en", value)
}

func TestCookieValueOrReturnsFallbackWhenMissing(t *testing.T) {
	t.Parallel()

	r := httptest.NewRequest(http.MethodGet, "/", nil)

	value := request.CookieValueOr(r, "lang", "fr")

	require.Equal(t, "fr", value)
}

func TestCookieValueOrReturnsFallbackWhenEmpty(t *testing.T) {
	t.Parallel()

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.AddCookie(&http.Cookie{Name: "lang", Value: ""})

	value := request.CookieValueOr(r, "lang", "fr")

	require.Equal(t, "fr", value)
}
