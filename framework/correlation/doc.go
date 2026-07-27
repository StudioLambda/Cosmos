// Package correlation provides correlation ID logging helpers.
//
// It establishes a request-scoped correlation identifier, stores it in
// context and exposes a slog handler decorator that injects that value into
// logs. Use [middleware.Correlation] to establish the correlation ID.
//
// Example
//
//	logger := slog.New(correlation.Handler(slog.NewJSONHandler(os.Stdout, nil)))
package correlation
