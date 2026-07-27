// Package logger provides logging drivers for Cosmos applications.
//
// The drivers adapt logging backends to [contract.LoggerDriver], allowing
// applications to depend on [contract.Logger] without coupling application
// code to a specific backend.
//
// # Drivers
//
// The package provides drivers for [log/slog], Zerolog, and Zap.
//
// Example
//
//	backend := slog.New(slog.NewJSONHandler(os.Stdout, nil))
//	log := contract.NewLogger(logger.NewSlogFrom(backend))
//	log.Info("started", "addr", ":8080")
package logger
