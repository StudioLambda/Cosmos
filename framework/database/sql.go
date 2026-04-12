package database

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
	"github.com/studiolambda/cosmos/contract"
)

// SQL implements contract.Database using sqlx for query execution,
// named parameters, and struct scanning. It supports both direct
// connections and transactions through the shared [sqlx.ExtContext]
// interface.
type SQL struct {
	db  sqlx.ExtContext // can be *sqlx.DB or *sqlx.Tx
	raw *sqlx.DB        // needed for transactions
}

// sqlTx is a transaction wrapper that overrides [SQL.Close] to
// prevent accidentally closing the underlying connection pool
// from within a transaction.
type sqlTx struct {
	SQL
}

// Close is a no-op on transaction wrappers. Transactions are managed
// by [SQL.WithTransaction] which handles commit and rollback.
func (tx *sqlTx) Close() error {
	return nil
}

// NewSQL connects to the database using the given driver name and
// DSN, returning a ready-to-use SQL instance or an error if the
// connection cannot be established.
//
// WARNING: No default query timeout is applied. Long-running or
// runaway queries will block indefinitely unless the caller
// passes a context.Context with a deadline or timeout. For
// production use, configure statement timeouts at the database
// driver level (e.g., statement_timeout for PostgreSQL) or wrap
// all query contexts with context.WithTimeout.
//
// WARNING: The default connection pool has no limits on open connections.
// Use [SQL.Configure] to set appropriate pool limits for production:
//
//	db.Configure(func(raw *sql.DB) {
//	    raw.SetMaxOpenConns(25)
//	    raw.SetMaxIdleConns(5)
//	    raw.SetConnMaxLifetime(5 * time.Minute)
//	})
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
//
// WARNING: No default query timeout is applied. See [NewSQL] for
// recommendations on configuring query timeouts.
//
// WARNING: The default connection pool has no limits on open connections.
// Use [SQL.Configure] to set appropriate pool limits for production:
//
//	db.Configure(func(raw *sql.DB) {
//	    raw.SetMaxOpenConns(25)
//	    raw.SetMaxIdleConns(5)
//	    raw.SetConnMaxLifetime(5 * time.Minute)
//	})
func NewSQLFrom(db *sqlx.DB) *SQL {
	return &SQL{db: db, raw: db}
}

// Ping verifies that the database connection is still alive,
// returning an error if the check fails.
func (database *SQL) Ping(ctx context.Context) error {
	return database.raw.PingContext(ctx)
}

// Exec executes a query that modifies data (INSERT, UPDATE, DELETE)
// using positional arguments. It returns the number of rows affected.
func (database *SQL) Exec(ctx context.Context, query string, args ...any) (int64, error) {
	result, err := database.db.ExecContext(ctx, query, args...)

	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}

// ExecNamed executes a query that modifies data using a named
// parameter struct or map. It returns the number of rows affected.
func (database *SQL) ExecNamed(ctx context.Context, query string, arg any) (int64, error) {
	result, err := sqlx.NamedExecContext(ctx, database.db, query, arg)

	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}

// Select executes a query that returns multiple rows, scanning the
// results into the dest slice using positional arguments.
func (database *SQL) Select(ctx context.Context, query string, dest any, args ...any) error {
	return sqlx.SelectContext(ctx, database.db, dest, query, args...)
}

// SelectNamed executes a query that returns multiple rows, scanning
// the results into the dest slice using a named parameter struct or
// map.
func (database *SQL) SelectNamed(ctx context.Context, query string, dest any, arg any) error {
	rows, err := sqlx.NamedQueryContext(ctx, database.db, query, arg)

	if err != nil {
		return err
	}

	defer rows.Close()

	return sqlx.StructScan(rows, dest)
}

// Find executes a query expected to return a single row, scanning
// the result into dest. If no row is found, the returned error wraps
// both sql.ErrNoRows and contract.ErrDatabaseNoRows.
func (database *SQL) Find(ctx context.Context, query string, dest any, args ...any) error {
	if err := sqlx.GetContext(ctx, database.db, dest, query, args...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.Join(err, contract.ErrDatabaseNoRows)
		}

		return err
	}

	return nil
}

// FindNamed executes a query expected to return a single row using a
// named parameter struct or map, scanning the result into dest. If no
// row is found, the returned error wraps both sql.ErrNoRows and
// contract.ErrDatabaseNoRows.
func (database *SQL) FindNamed(ctx context.Context, query string, dest any, arg any) error {
	rows, err := sqlx.NamedQueryContext(ctx, database.db, query, arg)

	if err != nil {
		return err
	}

	defer rows.Close()

	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return err
		}

		return errors.Join(sql.ErrNoRows, contract.ErrDatabaseNoRows)
	}

	if err := rows.StructScan(dest); err != nil {
		return err
	}

	return rows.Err()
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

	txWrapper := &sqlTx{SQL{db: tx, raw: database.raw}}

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

// Configure exposes the underlying *sql.DB for connection pool
// tuning. The provided function receives the raw *sql.DB so
// callers can set pool parameters such as SetMaxOpenConns,
// SetMaxIdleConns, and SetConnMaxLifetime without importing
// database/sql directly.
//
// Example:
//
//	db.Configure(func(raw *sql.DB) {
//	    raw.SetMaxOpenConns(25)
//	    raw.SetMaxIdleConns(10)
//	    raw.SetConnMaxLifetime(5 * time.Minute)
//	})
func (database *SQL) Configure(fn func(*sql.DB)) {
	fn(database.raw.DB)
}
