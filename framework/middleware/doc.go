// Package middleware provides reusable framework middleware constructors.
//
// Middleware in this package composes around framework handlers and returns
// transformed handlers without global state.
//
// # Ordering
//
// Middleware ordering affects observable behavior. Recovery and logging should
// be registered early; policy middleware (CORS/CSRF/rate limiting) should be
// arranged according to endpoint requirements.
//
// Example
//
//	app.Use(middleware.Recover())
//	app.Use(middleware.Logger(*slog.Default()))
//	app.Use(middleware.SecureHeaders())
//	app.Use(middleware.RateLimit())
package middleware
