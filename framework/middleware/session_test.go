package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	tmock "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/studiolambda/cosmos/contract"
	"github.com/studiolambda/cosmos/contract/mock"
	"github.com/studiolambda/cosmos/framework"
	"github.com/studiolambda/cosmos/framework/middleware"
)

func TestSessionCookieExists(t *testing.T) {
	handler := framework.Handler(func(w http.ResponseWriter, r *http.Request) error {
		return nil
	})

	cache := mock.NewSessionDriverMock(t)

	cache.On("Save", tmock.Anything, tmock.Anything, tmock.Anything).Return(nil).Once()

	handlerWithSessions := middleware.Session(cache)(handler)

	req := httptest.NewRequest("GET", "/", nil)
	res := handlerWithSessions.Record(req)

	cookies := res.Cookies()

	require.Len(t, cookies, 1)
	require.Len(t, cache.Calls, 1)
	require.Equal(t, cache.Calls[0].Arguments.Get(1).(contract.Session).SessionID(), cookies[0].Value)
	require.Equal(t, cookies[0].Name, middleware.DefaultSessionCookie)
}
