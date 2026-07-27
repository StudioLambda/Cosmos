package middleware

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/studiolambda/cosmos/contract"
	"github.com/studiolambda/cosmos/framework"
	"github.com/studiolambda/cosmos/problem"
)

// ErrRateLimited is the default error returned when a request
// exceeds the configured rate limit. It uses HTTP 429 Too Many
// Requests per RFC 6585.
var ErrRateLimited = problem.Problem{
	Title:  "Too Many Requests",
	Detail: "Rate limit exceeded. Please slow down and retry later.",
	Status: http.StatusTooManyRequests,
}

// RateLimitConfigFrom returns a RateLimitConfig read from configuration below prefix.
func RateLimitConfigFrom(configuration *contract.Configuration, prefix string) RateLimitConfig {
	return RateLimitConfig{
		Name:   configuration.GetOr(prefix+".name", ""),
		Limit:  configuration.GetOr(prefix+".limit", 0),
		Window: configuration.GetOr(prefix+".window", time.Duration(0)),
	}
}

// RateLimitConfig configures the rate limiter middleware.
type RateLimitConfig struct {
	// Name identifies the logical policy and becomes part of the cache key.
	Name string

	// Limit is the maximum number of allowed requests inside Window.
	Limit int

	// Window is the fixed duration over which Limit applies.
	Window time.Duration

	// ErrorResponse is the problem returned when a request is
	// rate-limited. Defaults to [ErrRateLimited].
	ErrorResponse problem.Problem
}

// RateLimitKeyFunc resolves the caller key used for rate limiting.
// Returning ok=false skips rate limiting for the current request.
type RateLimitKeyFunc = func(r *http.Request) (key string, ok bool)

// DefaultRateLimitConfig holds sensible defaults for error response
// and the default policy name.
var DefaultRateLimitConfig = RateLimitConfig{
	Name:          "default",
	Limit:         15,
	Window:        time.Second,
	ErrorResponse: ErrRateLimited,
}

func DefaultRateLimitKeyFunc(r *http.Request) (string, bool) {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		if r.RemoteAddr == "" {
			return "", false
		}

		return r.RemoteAddr, true
	}

	return host, true
}

// withDefaults returns a copy of the config with zero values
// replaced by the corresponding [DefaultRateLimitConfig] fields.
func (config RateLimitConfig) withDefaults() RateLimitConfig {
	if config.Name == "" {
		config.Name = DefaultRateLimitConfig.Name
	}

	if config.Limit == 0 {
		config.Limit = DefaultRateLimitConfig.Limit
	}

	if config.Window == 0 {
		config.Window = DefaultRateLimitConfig.Window
	}

	if config.ErrorResponse.Status == 0 {
		config.ErrorResponse = DefaultRateLimitConfig.ErrorResponse
	}

	return config
}

type rateLimitDecision struct {
	Allowed    bool
	Count      int64
	Remaining  int
	RetryAfter time.Duration
}

// RateLimit returns middleware that limits requests using the given
// cache backend, configuration, and [DefaultRateLimitKeyFunc].
func RateLimit(cache *contract.Cache, config RateLimitConfig) framework.Middleware {
	return RateLimitWith(cache, config, nil)
}

// RateLimitWith returns middleware that limits requests using
// the provided configuration and key resolver. It uses fixed-window
// counters stored in the configured cache backend.
func RateLimitWith(cache *contract.Cache, config RateLimitConfig, keyFunc RateLimitKeyFunc) framework.Middleware {
	if cache == nil || cache.Driver() == nil {
		panic("rate limit middleware: cache must not be nil")
	}

	config = config.withDefaults()

	if keyFunc == nil {
		keyFunc = DefaultRateLimitKeyFunc
	}

	return func(next framework.Handler) framework.Handler {
		return func(w http.ResponseWriter, r *http.Request) error {
			key, ok := keyFunc(r)
			if !ok {
				return next(w, r)
			}

			cacheKey := fmt.Sprintf("cosmos:ratelimit:%s:%s", config.Name, key)
			decision, err := takeFixedWindow(r.Context(), cache, cacheKey, config.Limit, config.Window)
			if err != nil {
				return err
			}

			writeRateLimitHeaders(w, config.Limit, decision)

			if !decision.Allowed {
				w.Header().Set("Retry-After", strconv.FormatInt(retryAfterSeconds(decision.RetryAfter), 10))
				return config.ErrorResponse
			}

			return next(w, r)
		}
	}
}

func writeRateLimitHeaders(w http.ResponseWriter, limit int, current rateLimitDecision) {
	secondsUntilReset := int64(math.Ceil(current.RetryAfter.Seconds()))
	if secondsUntilReset < 0 {
		secondsUntilReset = 0
	}

	w.Header().Set("RateLimit-Limit", strconv.Itoa(limit))
	w.Header().Set("RateLimit-Remaining", strconv.Itoa(current.Remaining))
	w.Header().Set("RateLimit-Reset", strconv.FormatInt(secondsUntilReset, 10))
}

func retryAfterSeconds(d time.Duration) int64 {
	seconds := int64(math.Ceil(d.Seconds()))
	if seconds < 0 {
		return 0
	}

	return seconds
}

func takeFixedWindow(
	ctx context.Context,
	cache *contract.Cache,
	key string,
	limit int,
	window time.Duration,
) (rateLimitDecision, error) {
	added, err := cache.Add(ctx, key, int64(0), window)
	if err != nil {
		return rateLimitDecision{}, err
	}

	if added {
		count, err := cache.Increment(ctx, key, 1)
		if err == nil {
			ttl, err := cache.TTL(ctx, key)
			if err != nil {
				return rateLimitDecision{}, err
			}

			return buildDecision(count, limit, ttl), nil
		}

		if !errors.Is(err, contract.ErrCacheKeyNotFound) {
			return rateLimitDecision{}, err
		}
	}

	count, err := cache.Increment(ctx, key, 1)
	if err != nil {
		if errors.Is(err, contract.ErrCacheKeyNotFound) {
			return takeFixedWindow(ctx, cache, key, limit, window)
		}

		return rateLimitDecision{}, err
	}

	ttl, err := cache.TTL(ctx, key)
	if err != nil {
		return rateLimitDecision{}, err
	}

	return buildDecision(count, limit, ttl), nil
}

func buildDecision(count int64, limit int, ttl time.Duration) rateLimitDecision {
	remaining := limit - int(count)
	if remaining < 0 {
		remaining = 0
	}

	return rateLimitDecision{
		Allowed:    count <= int64(limit),
		Count:      count,
		Remaining:  remaining,
		RetryAfter: ttl,
	}
}
