// Package correlation provides correlation ID middleware and logging helpers.
//
// It establishes a request-scoped correlation identifier, stores it in
// context, and exposes helpers to retrieve or inject that value into logs.
//
// # Ordering
//
// Correlation middleware should run near the start of the middleware chain so
// subsequent middleware and handlers can include IDs in logs and errors.
//
// Example
//
//	app.Use(correlation.Middleware())
//	logger := slog.New(correlation.Handler(slog.NewJSONHandler(os.Stdout, nil)))
//	app.Use(middleware.Logger(*logger))
package correlation
