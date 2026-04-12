package middleware

import (
	"net/http"
	"sync"
	"time"

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
	// per key (typically per IP). Defaults to 15.
	RequestsPerSecond float64

	// Burst is the maximum number of requests allowed in a
	// single burst above the sustained rate. Defaults to 30.
	Burst int

	// KeyFunc extracts the rate-limit key from a request.
	// Defaults to the client's remote address.
	KeyFunc func(r *http.Request) string

	// ErrorResponse is the problem returned when a request is
	// rate-limited. Defaults to [ErrRateLimited].
	ErrorResponse problem.Problem

	// CleanupInterval is how often the registry sweeps for
	// idle entries. Defaults to 1 minute.
	CleanupInterval time.Duration

	// MaxIdleTime is how long an entry can be idle before
	// being evicted. Defaults to 5 minutes.
	MaxIdleTime time.Duration
}

// DefaultRateLimitOptions holds sensible defaults: 15 req/s
// sustained with a burst of 30, keyed by remote address.
// Idle entries are evicted after 5 minutes of inactivity.
var DefaultRateLimitOptions = RateLimitOptions{
	RequestsPerSecond: 15,
	Burst:             30,
	KeyFunc: func(r *http.Request) string {
		return r.RemoteAddr
	},
	ErrorResponse:   ErrRateLimited,
	CleanupInterval: 1 * time.Minute,
	MaxIdleTime:     5 * time.Minute,
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

	if options.CleanupInterval == 0 {
		options.CleanupInterval = DefaultRateLimitOptions.CleanupInterval
	}

	if options.MaxIdleTime == 0 {
		options.MaxIdleTime = DefaultRateLimitOptions.MaxIdleTime
	}

	return options
}

// rateLimitEntry pairs a token bucket limiter with the time it
// was last accessed. Entries idle longer than [RateLimitOptions.MaxIdleTime]
// are evicted by the cleanup goroutine.
type rateLimitEntry struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// rateLimitRegistry manages per-key token bucket rate limiters
// with automatic eviction of idle entries. Each unique key gets
// its own [rate.Limiter] instance, created on first access and
// reused for subsequent requests. A background goroutine
// periodically removes entries that have been idle longer than
// the configured maximum idle time.
type rateLimitRegistry struct {
	// mu protects concurrent access to the entries map.
	mu sync.Mutex

	// entries maps rate-limit keys to their limiter and
	// last-seen timestamp.
	entries map[string]*rateLimitEntry

	// rps is the sustained requests-per-second rate for
	// newly created limiters.
	rps float64

	// burst is the maximum burst size for newly created
	// limiters.
	burst int

	// stop signals the cleanup goroutine to exit.
	stop chan struct{}
}

// newRateLimitRegistry creates a registry that produces limiters
// with the given sustained rate and burst size. It starts a
// background goroutine that evicts entries idle longer than
// maxIdle at the given interval. Call [rateLimitRegistry.close]
// to stop the goroutine.
func newRateLimitRegistry(
	rps float64,
	burst int,
	cleanupInterval time.Duration,
	maxIdle time.Duration,
) *rateLimitRegistry {
	registry := &rateLimitRegistry{
		entries: make(map[string]*rateLimitEntry),
		rps:     rps,
		burst:   burst,
		stop:    make(chan struct{}),
	}

	go registry.cleanup(cleanupInterval, maxIdle)

	return registry
}

// get returns the limiter for key, creating one if it does not
// already exist. It updates the entry's last-seen timestamp on
// every access.
func (registry *rateLimitRegistry) get(key string) *rate.Limiter {
	registry.mu.Lock()
	defer registry.mu.Unlock()

	if entry, ok := registry.entries[key]; ok {
		entry.lastSeen = time.Now()

		return entry.limiter
	}

	limiter := rate.NewLimiter(rate.Limit(registry.rps), registry.burst)

	registry.entries[key] = &rateLimitEntry{
		limiter:  limiter,
		lastSeen: time.Now(),
	}

	return limiter
}

// size returns the number of entries in the registry.
func (registry *rateLimitRegistry) size() int {
	registry.mu.Lock()
	defer registry.mu.Unlock()

	return len(registry.entries)
}

// cleanup periodically removes entries that have been idle
// longer than maxIdle. It runs until [rateLimitRegistry.close]
// is called.
func (registry *rateLimitRegistry) cleanup(
	interval time.Duration,
	maxIdle time.Duration,
) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-registry.stop:
			return
		case now := <-ticker.C:
			registry.mu.Lock()

			for key, entry := range registry.entries {
				if now.Sub(entry.lastSeen) > maxIdle {
					delete(registry.entries, key)
				}
			}

			registry.mu.Unlock()
		}
	}
}

// close stops the background cleanup goroutine.
func (registry *rateLimitRegistry) close() {
	close(registry.stop)
}

// RateLimit returns middleware that limits requests to 15 req/s
// per IP with a burst of 30 using [DefaultRateLimitOptions].
// Idle entries are automatically evicted after 5 minutes.
func RateLimit() framework.Middleware {
	return RateLimitWith(DefaultRateLimitOptions)
}

// RateLimitWith returns middleware that limits requests using
// the provided options. It uses a per-key token bucket algorithm
// backed by [golang.org/x/time/rate]. A background goroutine
// periodically evicts entries that have been idle longer than
// [RateLimitOptions.MaxIdleTime] to prevent unbounded memory
// growth.
func RateLimitWith(opts RateLimitOptions) framework.Middleware {
	opts = opts.withDefaults()

	registry := newRateLimitRegistry(
		opts.RequestsPerSecond,
		opts.Burst,
		opts.CleanupInterval,
		opts.MaxIdleTime,
	)

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
