// Package middleware provides reusable framework middleware constructors.
//
// Middleware in this package composes around framework handlers and returns
// transformed handlers without global state.
//
// # Ordering
//
// Middleware ordering affects observable behavior. Recovery, correlation, and
// logging should be registered early; policy middleware (CORS/CSRF/rate
// limiting) should be arranged according to endpoint requirements.
//
// Example
//
//	app.Use(middleware.Recover())
//	app.Use(middleware.Logger(contract.NewLogger(logger.NewSlogFrom(slog.Default()))))
//	app.Use(middleware.SecureHeaders(middleware.DefaultSecureHeadersConfig()))
//	app.Use(middleware.RateLimit(
//		contract.NewCache(cache.NewMemory(cache.MemoryConfig{Expiration: time.Second, Cleanup: time.Minute})),
//		middleware.DefaultRateLimitConfig(),
//	))
package middleware
