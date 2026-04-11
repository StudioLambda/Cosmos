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

// DefaultRateLimitOptions holds sensible defaults: 10 req/s
// sustained with a burst of 20, keyed by remote address.
var DefaultRateLimitOptions = RateLimitOptions{
	RequestsPerSecond: 10,
	Burst:             20,
	KeyFunc: func(r *http.Request) string {
		return r.RemoteAddr
	},
	ErrorResponse: ErrRateLimited,
}

// withDefaults returns a copy of the options with zero values
// replaced by the corresponding [DefaultRateLimitOptions] fields.
func (options RateLimitOptions) withDefaults() RateLimitOptions {
	if options.RequestsPerSecond == 0 {
		options.RequestsPerSecond = DefaultRateLimitOptions.RequestsPerSecond
	}

	if options.Burst == 0 {
		options.Burst = DefaultRateLimitOptions.Burst
	}

	if options.KeyFunc == nil {
		options.KeyFunc = DefaultRateLimitOptions.KeyFunc
	}

	if options.ErrorResponse.Status == 0 {
		options.ErrorResponse = DefaultRateLimitOptions.ErrorResponse
	}

	return options
}

// rateLimitRegistry manages per-key token bucket rate limiters.
// Each unique key gets its own [rate.Limiter] instance, created
// on first access and reused for subsequent requests.
type rateLimitRegistry struct {
	// mu protects concurrent access to the limiters map.
	mu sync.Mutex

	// limiters maps rate-limit keys to their token bucket.
	limiters map[string]*rate.Limiter

	// rps is the sustained requests-per-second rate for
	// newly created limiters.
	rps float64

	// burst is the maximum burst size for newly created
	// limiters.
	burst int
}

// newRateLimitRegistry creates a registry that produces limiters
// with the given sustained rate and burst size.
func newRateLimitRegistry(rps float64, burst int) *rateLimitRegistry {
	return &rateLimitRegistry{
		limiters: make(map[string]*rate.Limiter),
		rps:      rps,
		burst:    burst,
	}
}

// get returns the limiter for key, creating one if it does not
// already exist.
func (registry *rateLimitRegistry) get(key string) *rate.Limiter {
	registry.mu.Lock()
	defer registry.mu.Unlock()

	if limiter, ok := registry.limiters[key]; ok {
		return limiter
	}

	limiter := rate.NewLimiter(rate.Limit(registry.rps), registry.burst)
	registry.limiters[key] = limiter

	return limiter
}

// RateLimit returns middleware that limits requests to 10 req/s
// per IP with a burst of 20 using [DefaultRateLimitOptions].
func RateLimit() framework.Middleware {
	return RateLimitWith(DefaultRateLimitOptions)
}

// RateLimitWith returns middleware that limits requests using
// the provided options. It uses a per-key token bucket algorithm
// backed by [golang.org/x/time/rate].
func RateLimitWith(opts RateLimitOptions) framework.Middleware {
	opts = opts.withDefaults()
	registry := newRateLimitRegistry(opts.RequestsPerSecond, opts.Burst)

	return func(next framework.Handler) framework.Handler {
		return func(w http.ResponseWriter, r *http.Request) error {
			key := opts.KeyFunc(r)
			limiter := registry.get(key)

			if !limiter.Allow() {
				return opts.ErrorResponse
			}

			return next(w, r)
		}
	}
}
