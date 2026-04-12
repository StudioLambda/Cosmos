package middleware_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/studiolambda/cosmos/framework"
	"github.com/studiolambda/cosmos/framework/middleware"
	"github.com/studiolambda/cosmos/problem"

	"github.com/stretchr/testify/require"
)

func TestRateLimitAllowsWithinLimit(t *testing.T) {
	t.Parallel()

	handler := middleware.RateLimit()(framework.Handler(func(
		w http.ResponseWriter,
		r *http.Request,
	) error {
		w.WriteHeader(http.StatusOK)

		return nil
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	res := handler.Record(req)

	require.Equal(t, http.StatusOK, res.StatusCode)
}

func TestRateLimitBlocksExceedingBurst(t *testing.T) {
	t.Parallel()

	handler := middleware.RateLimitWith(middleware.RateLimitOptions{
		RequestsPerSecond: 1,
		Burst:             1,
	})(framework.Handler(func(
		w http.ResponseWriter,
		r *http.Request,
	) error {
		w.WriteHeader(http.StatusOK)

		return nil
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "192.168.1.1:12345"

	first := handler.Record(req)
	require.Equal(t, http.StatusOK, first.StatusCode)

	second := handler.Record(req)
	require.Equal(t, http.StatusTooManyRequests, second.StatusCode)
}

func TestRateLimitWithDefaultOptions(t *testing.T) {
	t.Parallel()

	called := false
	handler := middleware.RateLimit()(framework.Handler(func(
		w http.ResponseWriter,
		r *http.Request,
	) error {
		called = true
		w.WriteHeader(http.StatusOK)

		return nil
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	res := handler.Record(req)

	require.True(t, called)
	require.Equal(t, http.StatusOK, res.StatusCode)
}

func TestRateLimitWithCustomKeyFunc(t *testing.T) {
	t.Parallel()

	handler := middleware.RateLimitWith(middleware.RateLimitOptions{
		RequestsPerSecond: 1,
		Burst:             1,
		KeyFunc: func(r *http.Request) string {
			return r.Header.Get("X-API-Key")
		},
	})(framework.Handler(func(
		w http.ResponseWriter,
		r *http.Request,
	) error {
		w.WriteHeader(http.StatusOK)

		return nil
	}))

	reqA := httptest.NewRequest(http.MethodGet, "/", nil)
	reqA.Header.Set("X-API-Key", "key-a")
	resA := handler.Record(reqA)
	require.Equal(t, http.StatusOK, resA.StatusCode)

	reqB := httptest.NewRequest(http.MethodGet, "/", nil)
	reqB.Header.Set("X-API-Key", "key-b")
	resB := handler.Record(reqB)
	require.Equal(t, http.StatusOK, resB.StatusCode)

	reqA2 := httptest.NewRequest(http.MethodGet, "/", nil)
	reqA2.Header.Set("X-API-Key", "key-a")
	resA2 := handler.Record(reqA2)
	require.Equal(t, http.StatusTooManyRequests, resA2.StatusCode)
}

func TestRateLimitWithCustomErrorResponse(t *testing.T) {
	t.Parallel()

	customErr := problem.Problem{
		Title:  "Slow Down",
		Detail: "You are going too fast.",
		Status: http.StatusServiceUnavailable,
	}

	handler := middleware.RateLimitWith(middleware.RateLimitOptions{
		RequestsPerSecond: 1,
		Burst:             1,
		ErrorResponse:     customErr,
	})(framework.Handler(func(
		w http.ResponseWriter,
		r *http.Request,
	) error {
		w.WriteHeader(http.StatusOK)

		return nil
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "10.0.0.1:9999"

	first := handler.Record(req)
	require.Equal(t, http.StatusOK, first.StatusCode)

	second := handler.Record(req)
	require.Equal(t, http.StatusServiceUnavailable, second.StatusCode)
}

func TestRateLimitRegistryReusesExistingLimiter(t *testing.T) {
	t.Parallel()

	callCount := 0
	handler := middleware.RateLimitWith(middleware.RateLimitOptions{
		RequestsPerSecond: 100,
		Burst:             100,
		KeyFunc: func(r *http.Request) string {
			return "same-key"
		},
	})(framework.Handler(func(
		w http.ResponseWriter,
		r *http.Request,
	) error {
		callCount++
		w.WriteHeader(http.StatusOK)

		return nil
	}))

	for range 5 {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		res := handler.Record(req)
		require.Equal(t, http.StatusOK, res.StatusCode)
	}

	require.Equal(t, 5, callCount)
}

func TestRateLimitDifferentKeysGetSeparateLimiters(t *testing.T) {
	t.Parallel()

	handler := middleware.RateLimitWith(middleware.RateLimitOptions{
		RequestsPerSecond: 1,
		Burst:             1,
		KeyFunc: func(r *http.Request) string {
			return r.RemoteAddr
		},
	})(framework.Handler(func(
		w http.ResponseWriter,
		r *http.Request,
	) error {
		w.WriteHeader(http.StatusOK)

		return nil
	}))

	reqA := httptest.NewRequest(http.MethodGet, "/", nil)
	reqA.RemoteAddr = "10.0.0.1:1111"
	resA := handler.Record(reqA)
	require.Equal(t, http.StatusOK, resA.StatusCode)

	reqB := httptest.NewRequest(http.MethodGet, "/", nil)
	reqB.RemoteAddr = "10.0.0.2:2222"
	resB := handler.Record(reqB)
	require.Equal(t, http.StatusOK, resB.StatusCode)

	reqA2 := httptest.NewRequest(http.MethodGet, "/", nil)
	reqA2.RemoteAddr = "10.0.0.1:1111"
	resA2 := handler.Record(reqA2)
	require.Equal(t, http.StatusTooManyRequests, resA2.StatusCode)

	reqB2 := httptest.NewRequest(http.MethodGet, "/", nil)
	reqB2.RemoteAddr = "10.0.0.2:2222"
	resB2 := handler.Record(reqB2)
	require.Equal(t, http.StatusTooManyRequests, resB2.StatusCode)
}

func TestRateLimitWithZeroOptionsUsesDefaults(t *testing.T) {
	t.Parallel()

	handler := middleware.RateLimitWith(
		middleware.RateLimitOptions{},
	)(framework.Handler(func(
		w http.ResponseWriter,
		r *http.Request,
	) error {
		w.WriteHeader(http.StatusOK)

		return nil
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	res := handler.Record(req)

	require.Equal(t, http.StatusOK, res.StatusCode)
}

func TestRateLimitDefaultsAre15ReqPerSecBurst30(t *testing.T) {
	t.Parallel()

	require.Equal(t, float64(15), middleware.DefaultRateLimitOptions.RequestsPerSecond)
	require.Equal(t, 30, middleware.DefaultRateLimitOptions.Burst)
}

func TestRateLimitKeyFuncStripsPort(t *testing.T) {
	t.Parallel()

	handler := middleware.RateLimitWith(middleware.RateLimitOptions{
		RequestsPerSecond: 1,
		Burst:             1,
	})(framework.Handler(func(
		w http.ResponseWriter,
		r *http.Request,
	) error {
		w.WriteHeader(http.StatusOK)

		return nil
	}))

	first := httptest.NewRequest(http.MethodGet, "/", nil)
	first.RemoteAddr = "10.0.0.1:12345"
	res := handler.Record(first)
	require.Equal(t, http.StatusOK, res.StatusCode)

	second := httptest.NewRequest(http.MethodGet, "/", nil)
	second.RemoteAddr = "10.0.0.1:54321"
	res = handler.Record(second)
	require.Equal(t, http.StatusTooManyRequests, res.StatusCode)
}

func TestRateLimitKeyFuncFallsBackOnInvalidAddr(t *testing.T) {
	t.Parallel()

	handler := middleware.RateLimitWith(middleware.RateLimitOptions{
		RequestsPerSecond: 1,
		Burst:             1,
	})(framework.Handler(func(
		w http.ResponseWriter,
		r *http.Request,
	) error {
		w.WriteHeader(http.StatusOK)

		return nil
	}))

	first := httptest.NewRequest(http.MethodGet, "/", nil)
	first.RemoteAddr = "/var/run/app.sock"
	res := handler.Record(first)
	require.Equal(t, http.StatusOK, res.StatusCode)

	second := httptest.NewRequest(http.MethodGet, "/", nil)
	second.RemoteAddr = "/var/run/app.sock"
	res = handler.Record(second)
	require.Equal(t, http.StatusTooManyRequests, res.StatusCode)
}

func TestRateLimitMaxEntriesUsesOverflowLimiter(t *testing.T) {
	t.Parallel()

	handler := middleware.RateLimitWith(middleware.RateLimitOptions{
		RequestsPerSecond: 1,
		Burst:             2,
		MaxEntries:        2,
		KeyFunc: func(r *http.Request) string {
			return r.Header.Get("X-Key")
		},
	})(framework.Handler(func(
		w http.ResponseWriter,
		r *http.Request,
	) error {
		w.WriteHeader(http.StatusOK)

		return nil
	}))

	// Fill the registry to max capacity with 2 distinct keys.
	for _, key := range []string{"key-1", "key-2"} {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("X-Key", key)
		res := handler.Record(req)
		require.Equal(t, http.StatusOK, res.StatusCode)
	}

	// A third distinct key should use the shared overflow limiter.
	// The overflow limiter has burst=2, so the first two requests
	// succeed but the third is rate-limited.
	for i := range 2 {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("X-Key", fmt.Sprintf("overflow-%d", i))
		res := handler.Record(req)
		require.Equal(t, http.StatusOK, res.StatusCode)
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Key", "overflow-2")
	res := handler.Record(req)
	require.Equal(t, http.StatusTooManyRequests, res.StatusCode)
}

func TestRateLimitMaxEntriesDefaultIs10000(t *testing.T) {
	t.Parallel()

	require.Equal(t, 10000, middleware.DefaultRateLimitOptions.MaxEntries)
}
