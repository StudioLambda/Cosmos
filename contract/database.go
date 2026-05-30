package contract

import (
	"context"
	"errors"
)

var (
	// ErrDatabaseNoRows is the error that should be returned
	// when there are no rows found.
	ErrDatabaseNoRows = errors.New("no database rows were found")

	// ErrDatabaseNestedTransaction is the error that should be returned
	// when attempting to create a nested transaction, which is not supported.
	ErrDatabaseNestedTransaction = errors.New("nested transactions are not supported")
)

// DatabaseDriver defines the interface for interacting with a SQL-based
// datastore. Implementations handle query execution, scanning, and
// transaction management. The [Database] wrapper provides type-safe
// convenience methods on top of this driver.
type DatabaseDriver interface {
	// Close closes the database connection and releases associated resources.
	Close() error

	// Ping verifies that the database connection is still alive.
	Ping(ctx context.Context) error

	// Exec executes a SQL query that modifies data (e.g., INSERT, UPDATE, DELETE).
	// Returns the number of rows affected.
	Exec(ctx context.Context, query string, args ...any) (int64, error)

	// ExecNamed executes a SQL query using named parameters.
	// The arg parameter should be a struct or map containing the named parameters.
	// Returns the number of rows affected.
	ExecNamed(ctx context.Context, query string, arg any) (int64, error)

	// Select executes a query and scans the results into dest.
	// Dest should be a pointer to a slice of structs (e.g., *[]User).
	Select(ctx context.Context, query string, dest any, args ...any) error

	// SelectNamed executes a query using named parameters and scans results into dest.
	// Dest should be a pointer to a slice of structs.
	SelectNamed(ctx context.Context, query string, dest any, arg any) error

	// Find executes a query expected to return a single row and scans
	// the result into dest. Dest should be a pointer to a struct.
	Find(ctx context.Context, query string, dest any, args ...any) error

	// FindNamed executes a query using named parameters expected to return
	// a single row and scans the result into dest.
	FindNamed(ctx context.Context, query string, dest any, arg any) error

	// WithTransaction executes fn within a database transaction. If fn returns
	// an error, the transaction is rolled back. Otherwise, it is committed.
	WithTransaction(ctx context.Context, fn func(tx DatabaseDriver) error) error
}

// Database provides a wrapper over a [DatabaseDriver] with convenience
// methods. When generic methods become available in Go, Find and Select
// will be updated to return typed values directly.
type Database struct {
	driver DatabaseDriver
}

// NewDatabase creates a new [Database] that delegates operations to the given driver.
func NewDatabase(driver DatabaseDriver) *Database {
	return &Database{driver: driver}
}

// Driver returns the underlying [DatabaseDriver].
func (database *Database) Driver() DatabaseDriver {
	return database.driver
}

// Close closes the database connection and releases associated resources.
func (database *Database) Close() error {
	return database.driver.Close()
}

// Ping verifies that the database connection is still alive.
func (database *Database) Ping(ctx context.Context) error {
	return database.driver.Ping(ctx)
}

// Exec executes a SQL query that modifies data. Returns the number of rows affected.
func (database *Database) Exec(ctx context.Context, query string, args ...any) (int64, error) {
	return database.driver.Exec(ctx, query, args...)
}

// ExecNamed executes a SQL query using named parameters. Returns the number of rows affected.
func (database *Database) ExecNamed(ctx context.Context, query string, arg any) (int64, error) {
	return database.driver.ExecNamed(ctx, query, arg)
}

// Select executes a query and scans the results into dest.
// Dest should be a pointer to a slice of structs.
func (database *Database) Select(ctx context.Context, query string, dest any, args ...any) error {
	return database.driver.Select(ctx, query, dest, args...)
}

// SelectNamed executes a query using named parameters and scans results into dest.
func (database *Database) SelectNamed(ctx context.Context, query string, dest any, arg any) error {
	return database.driver.SelectNamed(ctx, query, dest, arg)
}

// Find executes a query expected to return a single row and scans the result into dest.
func (database *Database) Find(ctx context.Context, query string, dest any, args ...any) error {
	return database.driver.Find(ctx, query, dest, args...)
}

// FindNamed executes a query using named parameters expected to return a single row.
func (database *Database) FindNamed(ctx context.Context, query string, dest any, arg any) error {
	return database.driver.FindNamed(ctx, query, dest, arg)
}

// WithTransaction executes fn within a database transaction.
func (database *Database) WithTransaction(ctx context.Context, fn func(tx *Database) error) error {
	return database.driver.WithTransaction(ctx, func(tx DatabaseDriver) error {
		return fn(NewDatabase(tx))
	})
}
