// Package database provides a sqlx-backed contract.Database implementation.
//
// The adapter centralizes common SQL execution patterns and exposes transaction
// boundaries via [contract.Database.WithTransaction].
//
// # Transaction behavior
//
// Nested transactions are rejected to avoid ambiguous commit/rollback behavior.
// Statement resources are closed promptly to avoid pool starvation.
//
// Example
//
//	err := db.WithTransaction(ctx, func(tx contract.Database) error {
//		_, err := tx.Exec(ctx, "INSERT INTO users(name) VALUES($1)", "alice")
//		return err
//	})
//	_ = err
package database
