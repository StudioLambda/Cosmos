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

// DatabaseWithTransaction defines an interface for database operations that support transactions.
// It uses generics to allow the transaction object to implement the same interface as the parent,
// enabling type-safe transaction handling across different database implementations.
type DatabaseWithTransaction[T any] interface {
	// WithTransaction executes the provided function fn within a database transaction.
	// If fn returns an error, the transaction is rolled back. Otherwise, it is committed.
	// The tx passed to fn implements the same Database interface and can be used
	// for nested operations within the transaction.
	WithTransaction(ctx context.Context, fn func(tx T) error) error
}

// Database defines a generic interface for interacting with a SQL-based datastore.
// It provides methods for executing queries, retrieving multiple or single records,
// and performing operations within a transaction context.
type Database interface {
	DatabaseWithTransaction[Database]

	// Exec executes a SQL query that modifies data (e.g., INSERT, UPDATE, DELETE).
	// It returns the number of rows affected.
	Exec(ctx context.Context, query string, args ...any) (int64, error)

	// Select executes a query and scans the results into dest.
	// Dest should be a pointer to a slice of structs (e.g., *[]User).
	// Returns an error if the query fails or if scanning is unsuccessful.
	Select(ctx context.Context, query string, dest any, args ...any) error

	// Find executes a query expected to return a single row,
	// and scans the result into dest. Dest should be a pointer to a struct.
	// Returns an error if no row is found, multiple rows are returned, or scanning fails.
	Find(ctx context.Context, query string, dest any, args ...any) error
}

// DatabaseNamed extends the basic database functionality with named parameter support.
// It provides methods for executing queries using named parameters instead of positional ones,
// which can improve code readability and reduce errors when dealing with complex queries.
type DatabaseNamed interface {
	DatabaseWithTransaction[DatabaseNamed]

	// ExecNamed executes a SQL query that modifies data using named parameters.
	// The arg parameter should be a struct or map containing the named parameters.
	// Returns the number of rows affected.
	ExecNamed(ctx context.Context, query string, arg any) (int64, error)

	// SelectNamed executes a query using named parameters and scans the results into dest.
	// Dest should be a pointer to a slice of structs, and arg should contain the named parameters.
	// Returns an error if the query fails or if scanning is unsuccessful.
	SelectNamed(ctx context.Context, query string, dest any, arg any) error

	// FindNamed executes a query using named parameters expected to return a single row,
	// and scans the result into dest. Dest should be a pointer to a struct,
	// and arg should contain the named parameters.
	// Returns an error if no row is found, multiple rows are returned, or scanning fails.
	FindNamed(ctx context.Context, query string, dest any, arg any) error
}
