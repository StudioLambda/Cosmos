package database

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
	"github.com/studiolambda/cosmos/contract"
)

// prepare is the shared interface between *sqlx.DB and *sqlx.Tx
// that allows SQL to operate transparently on either a raw
// connection or an active transaction.
type prepare interface {
	PreparexContext(ctx context.Context, query string) (*sqlx.Stmt, error)
	PrepareNamedContext(ctx context.Context, query string) (*sqlx.NamedStmt, error)
}

// SQL implements contract.Database using sqlx for query preparation,
// named parameters, and struct scanning. It supports both direct
// connections and transactions through the shared prepare interface.
type SQL struct {
	db  prepare  // can be *sqlx.DB or *sqlx.Tx
	raw *sqlx.DB // needed for transactions
}

// NewSQL connects to the database using the given driver name and DSN,
// returning a ready-to-use SQL instance or an error if the connection
// cannot be established.
func NewSQL(driver string, dsn string) (*SQL, error) {
	db, err := sqlx.Connect(driver, dsn)

	if err != nil {
		return nil, err
	}

	return NewSQLFrom(db), nil
}

// NewSQLFrom wraps an existing sqlx.DB connection in a SQL instance.
// This is useful when you need to configure the connection pool or
// driver options before handing it to the framework.
func NewSQLFrom(db *sqlx.DB) *SQL {
	return &SQL{db: db, raw: db}
}

// Ping verifies that the database connection is still alive,
// returning an error if the check fails.
func (database *SQL) Ping(ctx context.Context) error {
	return database.raw.PingContext(ctx)
}

// Exec prepares and executes a query that modifies data (INSERT,
// UPDATE, DELETE) using positional arguments. It returns the number
// of rows affected.
func (database *SQL) Exec(ctx context.Context, query string, args ...any) (int64, error) {
	q, err := database.db.PreparexContext(ctx, query)

	if err != nil {
		return 0, err
	}

	result, err := q.ExecContext(ctx, args...)

	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}

// ExecNamed prepares and executes a query that modifies data using
// a named parameter struct or map. It returns the number of rows affected.
func (database *SQL) ExecNamed(ctx context.Context, query string, arg any) (int64, error) {
	q, err := database.db.PrepareNamedContext(ctx, query)

	if err != nil {
		return 0, err
	}

	result, err := q.ExecContext(ctx, arg)

	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}

// Select prepares and executes a query that returns multiple rows,
// scanning the results into the dest slice using positional arguments.
func (database *SQL) Select(ctx context.Context, query string, dest any, args ...any) error {
	q, err := database.db.PreparexContext(ctx, query)

	if err != nil {
		return err
	}

	return q.Select(dest, args...)
}

// SelectNamed prepares and executes a query that returns multiple rows,
// scanning the results into the dest slice using a named parameter
// struct or map.
func (database *SQL) SelectNamed(ctx context.Context, query string, dest any, arg any) error {
	q, err := database.db.PrepareNamedContext(ctx, query)

	if err != nil {
		return err
	}

	return q.Select(dest, arg)
}

// Find prepares and executes a query expected to return a single row,
// scanning the result into dest. If no row is found, the returned
// error wraps both sql.ErrNoRows and contract.ErrDatabaseNoRows.
func (database *SQL) Find(ctx context.Context, query string, dest any, args ...any) error {
	q, err := database.db.PreparexContext(ctx, query)

	if err != nil {
		return err
	}

	if err := q.GetContext(ctx, dest, args...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.Join(err, contract.ErrDatabaseNoRows)
		}

		return err
	}

	return nil
}

// FindNamed prepares and executes a query expected to return a single
// row using a named parameter struct or map, scanning the result into
// dest.
func (database *SQL) FindNamed(ctx context.Context, query string, dest any, arg any) error {
	q, err := database.db.PrepareNamedContext(ctx, query)

	if err != nil {
		return err
	}

	return q.GetContext(ctx, dest, arg)
}

// WithTransaction executes fn inside a database transaction. If fn
// returns an error, the transaction is rolled back and both the
// original error and any rollback error are joined. If fn succeeds,
// the transaction is committed. Nested transactions are not supported
// and return contract.ErrDatabaseNestedTransaction.
func (database *SQL) WithTransaction(ctx context.Context, fn func(tx contract.Database) error) error {
	if _, ok := database.db.(*sqlx.Tx); ok {
		return contract.ErrDatabaseNestedTransaction
	}

	tx, err := database.raw.BeginTxx(ctx, &sql.TxOptions{})

	if err != nil {
		return err
	}

	txWrapper := &SQL{db: tx, raw: database.raw}

	if err := fn(txWrapper); err != nil {
		return errors.Join(err, tx.Rollback())
	}

	return tx.Commit()
}

// Close closes the underlying database connection and releases
// all associated resources.
func (database *SQL) Close() error {
	return database.raw.Close()
}
