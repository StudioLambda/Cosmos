package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/studiolambda/cosmos/contract"
	"github.com/studiolambda/cosmos/framework"
	cache "github.com/studiolambda/cosmos/framework/cache/memory"
	"github.com/studiolambda/cosmos/framework/middleware"
	"github.com/studiolambda/cosmos/problem"
)

func TestRateLimitAllowsRequestsWithinLimit(t *testing.T) {
	t.Parallel()

	store := contract.NewCache(cache.NewMemory(cache.MemoryConfig{Expiration: time.Minute, Cleanup: time.Minute}))
	handler := middleware.RateLimitWith(store, middleware.RateLimitConfig{
		Name:   "login",
		Limit:  2,
		Window: time.Minute,
	}, func(*http.Request) (string, bool) {
		return "caller-a", true
	})(framework.Handler(func(w http.ResponseWriter, r *http.Request) error {
		w.WriteHeader(http.StatusOK)

		return nil
	}))

	res := handler.Record(httptest.NewRequest(http.MethodGet, "/", nil))

	require.Equal(t, http.StatusOK, res.StatusCode)
	require.Equal(t, "2", res.Header.Get("RateLimit-Limit"))
	require.Equal(t, "1", res.Header.Get("RateLimit-Remaining"))
	require.NotEmpty(t, res.Header.Get("RateLimit-Reset"))
	require.Empty(t, res.Header.Get("Retry-After"))
}

func TestRateLimitRejectsRequestsBeyondLimit(t *testing.T) {
	t.Parallel()

	store := contract.NewCache(cache.NewMemory(cache.MemoryConfig{Expiration: time.Minute, Cleanup: time.Minute}))
	handler := middleware.RateLimitWith(store, middleware.RateLimitConfig{
		Name:   "login",
		Limit:  1,
		Window: time.Minute,
	}, func(*http.Request) (string, bool) {
		return "caller-a", true
	})(framework.Handler(func(w http.ResponseWriter, r *http.Request) error {
		w.WriteHeader(http.StatusOK)

		return nil
	}))

	first := handler.Record(httptest.NewRequest(http.MethodGet, "/", nil))
	second := handler.Record(httptest.NewRequest(http.MethodGet, "/", nil))

	require.Equal(t, http.StatusOK, first.StatusCode)
	require.Equal(t, http.StatusTooManyRequests, second.StatusCode)
	require.NotEmpty(t, second.Header.Get("Retry-After"))
}

func TestRateLimitUsesIndependentBucketsPerKey(t *testing.T) {
	t.Parallel()

	store := contract.NewCache(cache.NewMemory(cache.MemoryConfig{Expiration: time.Minute, Cleanup: time.Minute}))
	keys := []string{"caller-a", "caller-b"}
	index := 0
	handler := middleware.RateLimitWith(store, middleware.RateLimitConfig{
		Name:   "login",
		Limit:  1,
		Window: time.Minute,
	}, func(*http.Request) (string, bool) {
		key := keys[index]
		index++
		return key, true
	})(framework.Handler(func(w http.ResponseWriter, r *http.Request) error {
		w.WriteHeader(http.StatusOK)

		return nil
	}))

	first := handler.Record(httptest.NewRequest(http.MethodGet, "/", nil))
	second := handler.Record(httptest.NewRequest(http.MethodGet, "/", nil))

	require.Equal(t, http.StatusOK, first.StatusCode)
	require.Equal(t, http.StatusOK, second.StatusCode)
}

func TestRateLimitSkipsRequestWhenKeyResolutionFails(t *testing.T) {
	t.Parallel()

	store := contract.NewCache(cache.NewMemory(cache.MemoryConfig{Expiration: time.Minute, Cleanup: time.Minute}))
	handler := middleware.RateLimitWith(store, middleware.RateLimitConfig{
		Name:   "login",
		Limit:  1,
		Window: time.Minute,
	}, func(*http.Request) (string, bool) {
		return "", false
	})(framework.Handler(func(w http.ResponseWriter, r *http.Request) error {
		w.WriteHeader(http.StatusOK)

		return nil
	}))

	first := handler.Record(httptest.NewRequest(http.MethodGet, "/", nil))
	second := handler.Record(httptest.NewRequest(http.MethodGet, "/", nil))

	require.Equal(t, http.StatusOK, first.StatusCode)
	require.Equal(t, http.StatusOK, second.StatusCode)
	require.Empty(t, second.Header.Get("RateLimit-Limit"))
}

func TestRateLimitUsesCustomErrorResponse(t *testing.T) {
	t.Parallel()

	store := contract.NewCache(cache.NewMemory(cache.MemoryConfig{Expiration: time.Minute, Cleanup: time.Minute}))
	customErr := problem.Problem{Title: "Slow Down", Status: http.StatusServiceUnavailable}
	handler := middleware.RateLimitWith(store, middleware.RateLimitConfig{
		Name:          "login",
		Limit:         1,
		Window:        time.Minute,
		ErrorResponse: customErr,
	}, func(*http.Request) (string, bool) {
		return "caller-a", true
	})(framework.Handler(func(w http.ResponseWriter, r *http.Request) error {
		w.WriteHeader(http.StatusOK)

		return nil
	}))

	require.Equal(t, http.StatusOK, handler.Record(httptest.NewRequest(http.MethodGet, "/", nil)).StatusCode)
	require.Equal(t, http.StatusServiceUnavailable, handler.Record(httptest.NewRequest(http.MethodGet, "/", nil)).StatusCode)
}

func TestRateLimitPropagatesCacheErrors(t *testing.T) {
	t.Parallel()

	store := contract.NewCache(brokenCacheDriver{})
	handler := middleware.RateLimitWith(store, middleware.RateLimitConfig{
		Name:   "login",
		Limit:  1,
		Window: time.Minute,
	}, func(*http.Request) (string, bool) {
		return "caller-a", true
	})(framework.Handler(func(w http.ResponseWriter, r *http.Request) error {
		w.WriteHeader(http.StatusOK)

		return nil
	}))

	err := handler(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))

	require.Equal(t, errBrokenCache, err)
}

func TestRateLimitUsesAtomicCounterWithTTL(t *testing.T) {
	t.Parallel()

	driver := &atomicCounterCacheDriver{}
	store := contract.NewCache(driver)
	handler := middleware.RateLimitWith(store, middleware.RateLimitConfig{
		Limit:  1,
		Window: time.Minute,
	}, func(*http.Request) (string, bool) {
		return "caller-a", true
	})(framework.Handler(func(w http.ResponseWriter, r *http.Request) error {
		w.WriteHeader(http.StatusOK)

		return nil
	}))

	res := handler.Record(httptest.NewRequest(http.MethodGet, "/", nil))

	require.Equal(t, http.StatusOK, res.StatusCode)
	require.Equal(t, 1, driver.incrementWithTTLCalls)
	require.Zero(t, driver.addCalls)
	require.Equal(t, time.Minute, driver.ttl)
}

type brokenCacheDriver struct{}

var errBrokenCache = problem.Problem{Title: "broken cache", Status: http.StatusInternalServerError}

func (brokenCacheDriver) Get(_ context.Context, _ string) ([]byte, error) { return nil, errBrokenCache }
func (brokenCacheDriver) Put(_ context.Context, _ string, _ []byte, _ time.Duration) error {
	return errBrokenCache
}
func (brokenCacheDriver) Delete(_ context.Context, _ string) error      { return errBrokenCache }
func (brokenCacheDriver) Has(_ context.Context, _ string) (bool, error) { return false, errBrokenCache }
func (brokenCacheDriver) Add(_ context.Context, _ string, _ []byte, _ time.Duration) (bool, error) {
	return false, errBrokenCache
}
func (brokenCacheDriver) Increment(_ context.Context, _ string, _ int64) (int64, error) {
	return 0, errBrokenCache
}
func (brokenCacheDriver) Decrement(_ context.Context, _ string, _ int64) (int64, error) {
	return 0, errBrokenCache
}
func (brokenCacheDriver) TTL(_ context.Context, _ string) (time.Duration, error) {
	return 0, errBrokenCache
}
func (brokenCacheDriver) Ping(_ context.Context) error { return errBrokenCache }

type atomicCounterCacheDriver struct {
	brokenCacheDriver
	incrementWithTTLCalls int
	addCalls              int
	ttl                   time.Duration
}

func (driver *atomicCounterCacheDriver) Add(ctx context.Context, key string, value []byte, ttl time.Duration) (bool, error) {
	driver.addCalls++

	return driver.brokenCacheDriver.Add(ctx, key, value, ttl)
}

func (driver *atomicCounterCacheDriver) IncrementWithTTL(_ context.Context, _ string, _ int64, ttl time.Duration) (int64, time.Duration, error) {
	driver.incrementWithTTLCalls++
	driver.ttl = ttl

	return 1, ttl, nil
}
