package contract

import (
	"context"
	"errors"
)

var (
	// ErrDatabaseNoRows is the error that should be returned
	// when there's no rows found.
	ErrDatabaseNoRows = errors.New("no database rows were found")

	// ErrDatabaseNestedTransaction is the error that should be returned
	// when attempting to create a nested transaction, which is not supported.
	ErrDatabaseNestedTransaction = errors.New("nested transactions are not supported")
)

// Database defines a generic interface for interacting with a SQL-based datastore.
// It provides methods for executing queries, retrieving multiple or single records,
// and performing operations within a transaction context.
type Database interface {
	// Close closes the database connection and releases any associated resources.
	// It should be called when the database is no longer needed to prevent
	// resource leaks. After calling Close, the Database instance should not
	// be used for further operations.
	Close() error

	// Ping verifies that the database connection is still alive and accessible.
	// It sends a simple query to the database to test connectivity and returns
	// an error if the connection is unavailable or the database is unreachable.
	// This method is typically used for health checks and connection validation.
	// Ping will also connect to the database if the connection was not established.
	Ping(ctx context.Context) error

	// Exec executes a SQL query that modifies data (e.g., INSERT, UPDATE, DELETE).
	// It returns the number of rows affected.
	Exec(ctx context.Context, query string, args ...any) (int64, error)

	// ExecNamed executes a SQL query that modifies data using named parameters.
	// The arg parameter should be a struct or map containing the named parameters.
	// Returns the number of rows affected.
	ExecNamed(ctx context.Context, query string, arg any) (int64, error)

	// Select executes a query and scans the results into dest.
	// Dest should be a pointer to a slice of structs (e.g., *[]User).
	// Returns an error if the query fails or if scanning is unsuccessful.
	Select(ctx context.Context, query string, dest any, args ...any) error

	// SelectNamed executes a query using named parameters and scans the results into dest.
	// Dest should be a pointer to a slice of structs, and arg should contain the named parameters.
	// Returns an error if the query fails or if scanning is unsuccessful.
	SelectNamed(ctx context.Context, query string, dest any, arg any) error

	// Find executes a query expected to return a single row,
	// and scans the result into dest. Dest should be a pointer to a struct.
	// Returns an error if no row is found, multiple rows are returned, or scanning fails.
	Find(ctx context.Context, query string, dest any, args ...any) error

	// FindNamed executes a query using named parameters expected to return a single row,
	// and scans the result into dest. Dest should be a pointer to a struct,
	// and arg should contain the named parameters.
	// Returns an error if no row is found, multiple rows are returned, or scanning fails.
	FindNamed(ctx context.Context, query string, dest any, arg any) error

	// WithTransaction executes the provided function fn within a database transaction.
	// If fn returns an error, the transaction is rolled back. Otherwise, it is committed.
	// The tx passed to fn implements the same Database interface and can be used
	// for nested operations within the transaction.
	WithTransaction(ctx context.Context, fn func(tx Database) error) error
}
