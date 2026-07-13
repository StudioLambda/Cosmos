package contract

import (
	"context"
	"errors"

	"github.com/studiolambda/cosmos/collection"
)

var (
	// ErrDatabaseNoRows is the error that should be returned
	// when there are no rows found.
	ErrDatabaseNoRows = errors.New("no database rows were found")

	// ErrDatabaseNestedTransaction is the error that should be returned
	// when attempting to create a nested transaction, which is not supported.
	ErrDatabaseNestedTransaction = errors.New("nested transactions are not supported")
)

// DatabaseRows represents an open database cursor for streaming query results.
type DatabaseRows interface {
	// Next advances the cursor to the next row. It returns false when there are
	// no more rows or when iteration can no longer continue.
	Next() bool

	// Scan scans the current row into dest.
	Scan(dest any) error

	// Err returns any error encountered while iterating the cursor.
	Err() error

	// Close closes the cursor and releases associated resources.
	Close() error
}

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

	// Query executes a query and returns a cursor over the result rows.
	Query(ctx context.Context, query string, args ...any) (DatabaseRows, error)

	// QueryNamed executes a query using named parameters and returns a cursor
	// over the result rows.
	QueryNamed(ctx context.Context, query string, arg any) (DatabaseRows, error)

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
// methods.
type Database struct {
	driver DatabaseDriver
}

// NewDatabase creates a new [Database] that delegates operations to the given driver.
//
// Example:
//
//	db := contract.NewDatabase(driver)
//	if err := db.Ping(context.Background()); err != nil {
//		return err
//	}
func NewDatabase(driver DatabaseDriver) *Database {
	return &Database{driver: driver}
}

// Driver returns the underlying [DatabaseDriver].
//
// Example:
//
//	db := contract.NewDatabase(driver)
//	rawDriver := db.Driver()
//	_ = rawDriver
func (database *Database) Driver() DatabaseDriver {
	return database.driver
}

// Close closes the database connection and releases associated resources.
//
// Example:
//
//	db := contract.NewDatabase(driver)
//	if err := db.Close(); err != nil {
//		return err
//	}
func (database *Database) Close() error {
	return database.driver.Close()
}

// Ping verifies that the database connection is still alive.
//
// Example:
//
//	if err := db.Ping(context.Background()); err != nil {
//		return err
//	}
func (database *Database) Ping(ctx context.Context) error {
	return database.driver.Ping(ctx)
}

// Exec executes a SQL query that modifies data. Returns the number of rows affected.
//
// Example:
//
//	affected, err := db.Exec(ctx, "DELETE FROM sessions WHERE expires_at < ?", now)
//	if err != nil {
//		return err
//	}
//	_ = affected
func (database *Database) Exec(ctx context.Context, query string, args ...any) (int64, error) {
	return database.driver.Exec(ctx, query, args...)
}

// ExecNamed executes a SQL query using named parameters. Returns the number of rows affected.
//
// Example:
//
//	affected, err := db.ExecNamed(ctx, "UPDATE users SET name=:name WHERE id=:id", map[string]any{"id": 1, "name": "Alice"})
//	if err != nil {
//		return err
//	}
//	_ = affected
func (database *Database) ExecNamed(ctx context.Context, query string, arg any) (int64, error) {
	return database.driver.ExecNamed(ctx, query, arg)
}

// Select executes a query and scans the results into a [collection.Slice].
//
// Example:
//
//	users, err := db.Select[User](ctx, "SELECT id, name FROM users WHERE active = ?", true)
//	if err != nil {
//		return err
//	}
//	_ = users.Items()
func (database *Database) Select[T any](ctx context.Context, query string, args ...any) (res collection.Slice[T], err error) {
	items := make([]T, 0)

	if err := database.driver.Select(ctx, query, &items, args...); err != nil {
		return res, err
	}

	return collection.NewSlice(items), nil
}

// SelectNamed executes a named query and scans the results into a [collection.Slice].
//
// Example:
//
//	users, err := db.SelectNamed[User](ctx, "SELECT id, name FROM users WHERE account_id=:account_id", map[string]any{"account_id": 42})
//	if err != nil {
//		return err
//	}
//	_ = users.Items()
func (database *Database) SelectNamed[T any](ctx context.Context, query string, arg any) (res collection.Slice[T], err error) {
	items := make([]T, 0)

	if err := database.driver.SelectNamed(ctx, query, &items, arg); err != nil {
		return res, err
	}

	return collection.NewSlice(items), nil
}

// Cursor executes a query and returns a lazy cursor over its results.
//
// Example:
//
//	users := db.Cursor[User](ctx, "SELECT id, name FROM users WHERE active = ?", true)
//	err := users.Each(func(user User) error {
//		return nil
//	})
//	if err != nil {
//		return err
//	}
func (database *Database) Cursor[T any](ctx context.Context, query string, args ...any) collection.TryLazySlice[T] {
	return collection.NewTryLazySlice(func(yield func(T, error) bool) {
		rows, err := database.driver.Query(ctx, query, args...)
		if err != nil {
			var zero T

			yield(zero, err)

			return
		}

		defer rows.Close()

		for rows.Next() {
			var item T

			if err := rows.Scan(&item); err != nil {
				if !yield(item, err) {
					return
				}

				return
			}

			if !yield(item, nil) {
				return
			}
		}

		if err := rows.Err(); err != nil {
			var zero T

			yield(zero, err)
		}
	})
}

// CursorNamed executes a named query and returns a lazy cursor over its results.
//
// Example:
//
//	users := db.CursorNamed[User](ctx, "SELECT id, name FROM users WHERE account_id=:account_id", map[string]any{"account_id": 42})
//	err := users.Each(func(user User) error {
//		return nil
//	})
//	if err != nil {
//		return err
//	}
func (database *Database) CursorNamed[T any](ctx context.Context, query string, arg any) collection.TryLazySlice[T] {
	return collection.NewTryLazySlice(func(yield func(T, error) bool) {
		rows, err := database.driver.QueryNamed(ctx, query, arg)
		if err != nil {
			var zero T

			yield(zero, err)

			return
		}

		defer rows.Close()

		for rows.Next() {
			var item T

			if err := rows.Scan(&item); err != nil {
				if !yield(item, err) {
					return
				}

				return
			}

			if !yield(item, nil) {
				return
			}
		}

		if err := rows.Err(); err != nil {
			var zero T

			yield(zero, err)
		}
	})
}

// Find executes a query expected to return a single row and scans the result into dest.
//
// Example:
//
//	user, err := db.Find[User](ctx, "SELECT id, name FROM users WHERE id = ?", 1)
//	if err != nil {
//		return err
//	}
//	_ = user
func (database *Database) Find[T any](ctx context.Context, query string, args ...any) (res T, err error) {
	if err := database.driver.Find(ctx, query, &res, args...); err != nil {
		return res, err
	}

	return res, nil
}

// FindNamed executes a query using named parameters expected to return a single row.
//
// Example:
//
//	user, err := db.FindNamed[User](ctx, "SELECT id, name FROM users WHERE id=:id", map[string]any{"id": 1})
//	if err != nil {
//		return err
//	}
//	_ = user
func (database *Database) FindNamed[T any](ctx context.Context, query string, arg any) (res T, err error) {
	if err := database.driver.FindNamed(ctx, query, &res, arg); err != nil {
		return res, err
	}

	return res, nil
}

// WithTransaction executes fn within a database transaction.
//
// Example:
//
//	if err := db.WithTransaction(ctx, func(tx *contract.Database) error {
//		if _, err := tx.Exec(ctx, "UPDATE accounts SET balance = balance - ? WHERE id = ?", 100, fromID); err != nil {
//			return err
//		}
//		if _, err := tx.Exec(ctx, "UPDATE accounts SET balance = balance + ? WHERE id = ?", 100, toID); err != nil {
//			return err
//		}
//		return nil
//	}); err != nil {
//		return err
//	}
func (database *Database) WithTransaction(ctx context.Context, fn func(tx *Database) error) error {
	return database.driver.WithTransaction(ctx, func(tx DatabaseDriver) error {
		return fn(NewDatabase(tx))
	})
}
