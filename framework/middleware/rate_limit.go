package middleware

import (
	"net/http"
	"sync"

	"github.com/studiolambda/cosmos/framework"
	"github.com/studiolambda/cosmos/problem"
	"golang.org/x/time/rate"
)

// ErrRateLimited is the default error returned when a request
// exceeds the configured rate limit. It uses HTTP 429 Too Many
// Requests per RFC 6585.
var ErrRateLimited = problem.Problem{
	Title:  "Too Many Requests",
	Detail: "Rate limit exceeded. Please slow down and retry later.",
	Status: http.StatusTooManyRequests,
}

// RateLimitOptions configures the rate limiter middleware.
type RateLimitOptions struct {
	// RequestsPerSecond is the sustained request rate allowed
	// per key (typically per IP). Defaults to 10.
	RequestsPerSecond float64

	// Burst is the maximum number of requests allowed in a
	// single burst above the sustained rate. Defaults to 20.
	Burst int

	// KeyFunc extracts the rate-limit key from a request.
	// Defaults to the client's remote address.
	KeyFunc func(r *http.Request) string

	// ErrorResponse is the problem returned when a request is
	// rate-limited. Defaults to [ErrRateLimited].
	ErrorResponse problem.Problem
}

// DefaultRateLimitOptions returns sensible defaults: 10 req/s
// sustained with a burst of 20, keyed by remote address.
func DefaultRateLimitOptions() RateLimitOptions {
	return RateLimitOptions{
		RequestsPerSecond: 10,
		Burst:             20,
		KeyFunc: func(r *http.Request) string {
			return r.RemoteAddr
		},
		ErrorResponse: ErrRateLimited,
	}
}

// RateLimit returns middleware that limits requests to 10 req/s
// per IP with a burst of 20 using [DefaultRateLimitOptions].
func RateLimit() framework.Middleware {
	return RateLimitWith(DefaultRateLimitOptions())
}

// RateLimitWith returns middleware that limits requests using
// the provided options. It uses a per-key token bucket algorithm
// backed by [golang.org/x/time/rate].
func RateLimitWith(opts RateLimitOptions) framework.Middleware {
	defaults := DefaultRateLimitOptions()

	if opts.RequestsPerSecond == 0 {
		opts.RequestsPerSecond = defaults.RequestsPerSecond
	}

	if opts.Burst == 0 {
		opts.Burst = defaults.Burst
	}

	if opts.KeyFunc == nil {
		opts.KeyFunc = defaults.KeyFunc
	}

	if opts.ErrorResponse.Status == 0 {
		opts.ErrorResponse = defaults.ErrorResponse
	}

	var (
		mu       sync.Mutex
		limiters = make(map[string]*rate.Limiter)
	)

	getLimiter := func(key string) *rate.Limiter {
		mu.Lock()
		defer mu.Unlock()

		if limiter, ok := limiters[key]; ok {
			return limiter
		}

		limiter := rate.NewLimiter(rate.Limit(opts.RequestsPerSecond), opts.Burst)
		limiters[key] = limiter

		return limiter
	}

	return func(next framework.Handler) framework.Handler {
		return func(w http.ResponseWriter, r *http.Request) error {
			key := opts.KeyFunc(r)
			limiter := getLimiter(key)

			if !limiter.Allow() {
				return opts.ErrorResponse
			}

			return next(w, r)
		}
	}
}
