package contract

import (
	"context"
	"errors"
)

// ErrDatabaseNoRows is the error that should be returned
// when there's no rows found.
var ErrDatabaseNoRows = errors.New("no database rows were found")

// Database defines a generic interface for interacting with a SQL-based datastore.
// It provides methods for executing queries, retrieving multiple or single records,
// and performing operations within a transaction context.
type Database interface {
	// Exec executes a SQL query that modifies data (e.g., INSERT, UPDATE, DELETE).
	// It returns the number of rows affected.
	Exec(ctx context.Context, query string, args ...any) (int64, error)

	// Query executes a SQL SELECT query and scans the results into dest.
	// Dest should be a pointer to a slice of structs (e.g., *[]User).
	// Returns an error if the query fails or if scanning is unsuccessful.
	Query(ctx context.Context, dest any, query string, args ...any) error

	// QueryOne executes a SQL SELECT query expected to return a single row,
	// and scans the result into dest. Dest should be a pointer to a struct.
	// Returns an error if no row is found, multiple rows are returned, or scanning fails.
	QueryOne(ctx context.Context, dest any, query string, args ...any) error

	// WithTransaction executes the provided function fn within a database transaction.
	// If fn returns an error, the transaction is rolled back. Otherwise, it is committed.
	// The tx passed to fn implements the same Database interface and can be used
	// for nested operations within the transaction.
	WithTransaction(ctx context.Context, fn func(tx Database) error) error
}
