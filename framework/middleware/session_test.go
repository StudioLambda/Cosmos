package middleware_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/studiolambda/cosmos/contract"
	"github.com/studiolambda/cosmos/contract/mock"
	"github.com/studiolambda/cosmos/contract/request"
	"github.com/studiolambda/cosmos/framework"
	sessionmiddleware "github.com/studiolambda/cosmos/framework/middleware"

	tmock "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestMiddlewareCookieExists(t *testing.T) {
	t.Parallel()

	handler := framework.Handler(func(w http.ResponseWriter, r *http.Request) error {
		return nil
	})

	cache := mock.NewSessionDriverMock(t)

	cache.On("Save", tmock.Anything, tmock.Anything, tmock.Anything).Return(nil).Once()

	handlerWithSessions := sessionmiddleware.Session(cache, sessionmiddleware.DefaultSessionConfig)(handler)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	res := handlerWithSessions.Record(req)

	cookies := res.Cookies()

	require.Len(t, cookies, 1)
	require.Len(t, cache.Calls, 1)
	require.Equal(t, cache.Calls[0].Arguments.Get(1).(*contract.Session).SessionID(), cookies[0].Value)
	require.Equal(t, sessionmiddleware.DefaultSessionCookie, cookies[0].Name)
}

func TestMiddlewareLoadsExistingSession(t *testing.T) {
	t.Parallel()

	existingSession, err := contract.NewSession(
		time.Now().Add(2*time.Hour),
		map[string]any{"user_id": 42},
	)

	require.NoError(t, err)

	sessionID := existingSession.SessionID()

	driver := mock.NewSessionDriverMock(t)
	driver.On(
		"Get", tmock.Anything, sessionID,
	).Return(existingSession, nil).Once()
	driver.On(
		"Save", tmock.Anything, tmock.Anything, tmock.Anything,
	).Return(nil).Maybe()

	var found bool
	var userID any

	handler := framework.Handler(
		func(w http.ResponseWriter, r *http.Request) error {
			sess, ok := request.Session(r)
			found = ok

			if ok {
				value, err := sess.Get[int]("user_id")
				require.NoError(t, err)
				userID = value
			}

			return nil
		},
	)

	handlerWithSessions := sessionmiddleware.Session(driver, sessionmiddleware.DefaultSessionConfig)(handler)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{
		Name:  sessionmiddleware.DefaultSessionCookie,
		Value: sessionID,
	})

	_ = handlerWithSessions.Record(req)

	require.True(t, found)
	require.Equal(t, 42, userID)
}

func TestMiddlewareCreatesNewSessionForInvalidCookieID(t *testing.T) {
	t.Parallel()

	driver := mock.NewSessionDriverMock(t)
	driver.On(
		"Save", tmock.Anything, tmock.Anything, tmock.Anything,
	).Return(nil).Once()

	var sessionFound bool

	handler := framework.Handler(
		func(w http.ResponseWriter, r *http.Request) error {
			_, sessionFound = request.Session(r)
			return nil
		},
	)

	handlerWithSessions := sessionmiddleware.Session(driver, sessionmiddleware.DefaultSessionConfig)(handler)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{
		Name:  sessionmiddleware.DefaultSessionCookie,
		Value: "invalid!@#$",
	})

	_ = handlerWithSessions.Record(req)

	require.True(t, sessionFound)
}

func TestMiddlewareCreatesNewSessionWhenDriverFails(t *testing.T) {
	t.Parallel()

	validID := "ABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890abcdefg"

	driver := mock.NewSessionDriverMock(t)
	driver.On(
		"Get", tmock.Anything, validID,
	).Return(nil, errors.New("driver error")).Once()
	driver.On(
		"Save", tmock.Anything, tmock.Anything, tmock.Anything,
	).Return(nil).Once()

	var sessionFound bool

	handler := framework.Handler(
		func(w http.ResponseWriter, r *http.Request) error {
			_, sessionFound = request.Session(r)
			return nil
		},
	)

	handlerWithSessions := sessionmiddleware.Session(driver, sessionmiddleware.DefaultSessionConfig)(handler)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{
		Name:  sessionmiddleware.DefaultSessionCookie,
		Value: validID,
	})

	_ = handlerWithSessions.Record(req)

	require.True(t, sessionFound)
}

func TestMiddlewareWithExpiredSessionRegenerates(t *testing.T) {
	t.Parallel()

	expiredSession, err := contract.NewSession(
		time.Now().Add(-1*time.Hour),
		map[string]any{},
	)

	require.NoError(t, err)

	sessionID := expiredSession.SessionID()

	driver := mock.NewSessionDriverMock(t)
	driver.On(
		"Get", tmock.Anything, sessionID,
	).Return(expiredSession, nil).Once()
	driver.On(
		"Delete", tmock.Anything, tmock.Anything,
	).Return(nil).Maybe()
	driver.On(
		"Save", tmock.Anything, tmock.Anything, tmock.Anything,
	).Return(nil).Maybe()

	handler := framework.Handler(
		func(w http.ResponseWriter, r *http.Request) error {
			return nil
		},
	)

	handlerWithSessions := sessionmiddleware.Session(driver, sessionmiddleware.DefaultSessionConfig)(handler)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{
		Name:  sessionmiddleware.DefaultSessionCookie,
		Value: sessionID,
	})

	res := handlerWithSessions.Record(req)
	cookies := res.Cookies()

	require.Len(t, cookies, 1)
	require.NotEqual(t, sessionID, cookies[0].Value)
}

func TestMiddlewareWithExpirationDeltaExtendsSession(t *testing.T) {
	t.Parallel()

	soonSession, err := contract.NewSession(
		time.Now().Add(5*time.Minute),
		map[string]any{},
	)

	require.NoError(t, err)

	sessionID := soonSession.SessionID()

	driver := mock.NewSessionDriverMock(t)
	driver.On(
		"Get", tmock.Anything, sessionID,
	).Return(soonSession, nil).Once()
	driver.On(
		"Save", tmock.Anything, tmock.Anything, tmock.Anything,
	).Return(nil).Maybe()

	handler := framework.Handler(
		func(w http.ResponseWriter, r *http.Request) error {
			return nil
		},
	)

	handlerWithSessions := sessionmiddleware.Session(
		driver, sessionmiddleware.SessionConfig{
			ExpirationDelta: 10 * time.Minute,
			TTL:             2 * time.Hour,
		},
	)(handler)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{
		Name:  sessionmiddleware.DefaultSessionCookie,
		Value: sessionID,
	})

	res := handlerWithSessions.Record(req)
	cookies := res.Cookies()

	require.Len(t, cookies, 1)
	require.True(t, cookies[0].Expires.After(time.Now().Add(1*time.Hour)))
}

func TestMiddlewareWithDefaultShortcutUsesDefaults(t *testing.T) {
	t.Parallel()

	driver := mock.NewSessionDriverMock(t)
	driver.On(
		"Save", tmock.Anything, tmock.Anything, tmock.Anything,
	).Return(nil).Once()

	handler := framework.Handler(
		func(w http.ResponseWriter, r *http.Request) error {
			return nil
		},
	)

	handlerWithSessions := sessionmiddleware.Session(driver, sessionmiddleware.DefaultSessionConfig)(handler)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	res := handlerWithSessions.Record(req)

	cookies := res.Cookies()

	require.Len(t, cookies, 1)
	require.Equal(t, sessionmiddleware.DefaultSessionCookie, cookies[0].Name)
	require.Equal(t, "/", cookies[0].Path)
	require.True(t, cookies[0].Secure)
	require.True(t, cookies[0].HttpOnly)
}

func TestMiddlewareErrorHandlerCalledOnSaveError(t *testing.T) {
	t.Parallel()

	driver := mock.NewSessionDriverMock(t)
	saveErr := errors.New("save failed")

	driver.On(
		"Save", tmock.Anything, tmock.Anything, tmock.Anything,
	).Return(saveErr).Once()

	var capturedErr error

	handler := framework.Handler(
		func(w http.ResponseWriter, r *http.Request) error {
			return nil
		},
	)

	handlerWithSessions := sessionmiddleware.SessionWith(
		driver, sessionmiddleware.DefaultSessionConfig, sessionmiddleware.SessionRuntime{
			ErrorHandler: func(err error) {
				capturedErr = err
			},
		},
	)(handler)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	_ = handlerWithSessions.Record(req)

	require.ErrorIs(t, capturedErr, saveErr)
}

func TestMiddlewareSessionNotSavedWhenUnchanged(t *testing.T) {
	t.Parallel()

	existingSession, err := contract.NewSession(
		time.Now().Add(2*time.Hour),
		map[string]any{},
	)

	require.NoError(t, err)

	sessionID := existingSession.SessionID()

	driver := mock.NewSessionDriverMock(t)
	driver.On(
		"Get", tmock.Anything, sessionID,
	).Return(existingSession, nil).Once()

	handler := framework.Handler(
		func(w http.ResponseWriter, r *http.Request) error {
			return nil
		},
	)

	handlerWithSessions := sessionmiddleware.Session(
		driver, sessionmiddleware.SessionConfig{
			ExpirationDelta: 1 * time.Minute,
		},
	)(handler)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{
		Name:  sessionmiddleware.DefaultSessionCookie,
		Value: sessionID,
	})

	res := handlerWithSessions.Record(req)
	cookies := res.Cookies()

	require.Empty(t, cookies)
}

func TestMiddlewareDoesNotSetCookieWhenSaveFails(t *testing.T) {
	t.Parallel()

	driver := mock.NewSessionDriverMock(t)
	saveErr := errors.New("save failed")

	driver.On(
		"Save", tmock.Anything, tmock.Anything, tmock.Anything,
	).Return(saveErr).Once()

	var capturedErr error

	handler := framework.Handler(
		func(w http.ResponseWriter, r *http.Request) error {
			return nil
		},
	)

	handlerWithSessions := sessionmiddleware.SessionWith(
		driver, sessionmiddleware.DefaultSessionConfig, sessionmiddleware.SessionRuntime{
			ErrorHandler: func(err error) {
				capturedErr = err
			},
		},
	)(handler)

	req := httptest.NewRequest("GET", "/", nil)
	res := handlerWithSessions.Record(req)

	require.ErrorIs(t, capturedErr, saveErr)
	require.Empty(t, res.Cookies(), "no Set-Cookie header should be set when Save fails")
}

func TestMiddlewareRegenerateDeletesOldSession(t *testing.T) {
	t.Parallel()

	existingSession, err := contract.NewSession(
		time.Now().Add(2*time.Hour),
		map[string]any{},
	)

	require.NoError(t, err)

	sessionID := existingSession.SessionID()

	driver := mock.NewSessionDriverMock(t)
	driver.On(
		"Get", tmock.Anything, sessionID,
	).Return(existingSession, nil).Once()
	driver.On(
		"Delete", tmock.Anything, sessionID,
	).Return(nil).Once()
	driver.On(
		"Save", tmock.Anything, tmock.Anything, tmock.Anything,
	).Return(nil).Once()

	handler := framework.Handler(
		func(w http.ResponseWriter, r *http.Request) error {
			sess, ok := request.Session(r)
			require.True(t, ok)

			sess.Regenerate()

			return nil
		},
	)

	handlerWithSessions := sessionmiddleware.Session(driver, sessionmiddleware.DefaultSessionConfig)(handler)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{
		Name:  sessionmiddleware.DefaultSessionCookie,
		Value: sessionID,
	})

	res := handlerWithSessions.Record(req)
	cookies := res.Cookies()

	require.Len(t, cookies, 1)
	require.NotEqual(t, sessionID, cookies[0].Value)
}

func TestMiddlewareCreatesNewSessionWhenExpired(t *testing.T) {
	t.Parallel()

	expiredSession, err := contract.NewSession(
		time.Now().Add(-1*time.Hour),
		map[string]any{"stale": true},
	)

	require.NoError(t, err)

	sessionID := expiredSession.SessionID()

	driver := mock.NewSessionDriverMock(t)
	driver.On(
		"Get", tmock.Anything, sessionID,
	).Return(expiredSession, nil).Once()
	driver.On(
		"Delete", tmock.Anything, sessionID,
	).Return(nil).Once()
	driver.On(
		"Save", tmock.Anything, tmock.Anything, tmock.Anything,
	).Return(nil).Once()

	var staleValue bool
	var staleErr error

	handler := framework.Handler(
		func(w http.ResponseWriter, r *http.Request) error {
			sess, ok := request.Session(r)
			require.True(t, ok)

			staleValue, staleErr = sess.Get[bool]("stale")

			return nil
		},
	)

	handlerWithSessions := sessionmiddleware.Session(driver, sessionmiddleware.DefaultSessionConfig)(handler)

	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{
		Name:  sessionmiddleware.DefaultSessionCookie,
		Value: sessionID,
	})

	res := handlerWithSessions.Record(req)
	cookies := res.Cookies()

	require.ErrorIs(t, staleErr, contract.ErrSessionKeyNotFound, "handler must not see expired session data")
	require.False(t, staleValue)
	require.Len(t, cookies, 1)
	require.NotEqual(t, sessionID, cookies[0].Value)
}

func TestMiddlewareWithDefaultsMaxLifetime(t *testing.T) {
	t.Parallel()

	oldSession, err := contract.NewSession(
		time.Now().Add(2*time.Hour),
		map[string]any{"user_id": 7},
	)

	require.NoError(t, err)

	sessionID := oldSession.SessionID()

	driver := mock.NewSessionDriverMock(t)
	driver.On(
		"Get", tmock.Anything, sessionID,
	).Return(oldSession, nil).Once()
	driver.On(
		"Save", tmock.Anything, tmock.Anything, tmock.Anything,
	).Return(nil).Maybe()

	var sessionFound bool

	handler := framework.Handler(
		func(w http.ResponseWriter, r *http.Request) error {
			_, sessionFound = request.Session(r)
			return nil
		},
	)

	// Middleware with zero MaxLifetime should get DefaultMaxLifetime.
	// The session was just created, so MaxLifetime (24h) should not trigger.
	handlerWithSessions := sessionmiddleware.Session(
		driver, sessionmiddleware.SessionConfig{},
	)(handler)

	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{
		Name:  sessionmiddleware.DefaultSessionCookie,
		Value: sessionID,
	})

	_ = handlerWithSessions.Record(req)

	require.True(t, sessionFound)
	require.Equal(t, sessionmiddleware.DefaultSessionMaxLifetime, 24*time.Hour)
}

func TestMiddlewareWithAllowInsecureDisablesSecureCookie(t *testing.T) {
	t.Parallel()

	driver := mock.NewSessionDriverMock(t)
	driver.On(
		"Save", tmock.Anything, tmock.Anything, tmock.Anything,
	).Return(nil).Once()

	handler := framework.Handler(
		func(w http.ResponseWriter, r *http.Request) error {
			return nil
		},
	)

	handlerWithSessions := sessionmiddleware.Session(
		driver, sessionmiddleware.SessionConfig{
			AllowInsecure: true,
		},
	)(handler)

	req := httptest.NewRequest("GET", "/", nil)
	res := handlerWithSessions.Record(req)

	cookies := res.Cookies()

	require.Len(t, cookies, 1)
	require.False(t, cookies[0].Secure)
}

func TestMiddlewarePartialConfigKeepsSecureCookie(t *testing.T) {
	t.Parallel()

	driver := mock.NewSessionDriverMock(t)
	driver.On("Save", tmock.Anything, tmock.Anything, tmock.Anything).Return(nil).Once()
	handler := sessionmiddleware.Session(driver, sessionmiddleware.SessionConfig{
		TTL: time.Hour,
	})(framework.Handler(func(w http.ResponseWriter, r *http.Request) error {
		return nil
	}))

	res := handler.Record(httptest.NewRequest(http.MethodGet, "/", nil))

	require.Len(t, res.Cookies(), 1)
	require.True(t, res.Cookies()[0].Secure)
}

func TestMiddlewareBoundsSessionPersistence(t *testing.T) {
	t.Parallel()

	driver := &blockingSessionDriver{}
	handler := sessionmiddleware.SessionWith(driver, sessionmiddleware.DefaultSessionConfig, sessionmiddleware.SessionRuntime{
		PersistenceTimeout: 10 * time.Millisecond,
	})(framework.Handler(func(w http.ResponseWriter, r *http.Request) error {
		return nil
	}))

	started := time.Now()
	_ = handler.Record(httptest.NewRequest(http.MethodGet, "/", nil))

	require.Less(t, time.Since(started), time.Second)
	require.True(t, driver.saveContextCancelled)
}

type blockingSessionDriver struct {
	saveContextCancelled bool
}

func (driver *blockingSessionDriver) Get(context.Context, string) (*contract.Session, error) {
	return nil, contract.ErrSessionKeyNotFound
}

func (driver *blockingSessionDriver) Save(ctx context.Context, session *contract.Session, ttl time.Duration) error {
	<-ctx.Done()
	driver.saveContextCancelled = true

	return ctx.Err()
}

func (driver *blockingSessionDriver) Delete(context.Context, string) error {
	return nil
}
