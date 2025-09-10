package sql

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
	"github.com/studiolambda/cosmos/contract"
)

type prepare interface {
	PreparexContext(ctx context.Context, query string) (*sqlx.Stmt, error)
	PrepareNamedContext(ctx context.Context, query string) (*sqlx.NamedStmt, error)
}

type Database struct {
	db  prepare  // can be *sqlx.DB or *sqlx.Tx
	raw *sqlx.DB // needed for transactions
}

func New(driver string, dsn string) (*Database, error) {
	db, err := sqlx.Open(driver, dsn)

	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &Database{db: db, raw: db}, nil
}

func (db *Database) Exec(ctx context.Context, query string, args ...any) (int64, error) {
	q, err := db.db.PreparexContext(ctx, query)

	if err != nil {
		return 0, err
	}

	result, err := q.ExecContext(ctx, args...)

	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}

func (db *Database) ExecNamed(ctx context.Context, query string, arg any) (int64, error) {
	q, err := db.db.PrepareNamedContext(ctx, query)

	if err != nil {
		return 0, err
	}

	result, err := q.ExecContext(ctx, arg)

	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}

func (db *Database) Select(ctx context.Context, query string, dest any, args ...any) error {
	q, err := db.db.PreparexContext(ctx, query)

	if err != nil {
		return err
	}

	if err := q.Select(dest, args...); err != nil {
		return err
	}

	return nil
}

func (db *Database) SelectNamed(ctx context.Context, query string, dest any, arg any) error {
	q, err := db.db.PrepareNamedContext(ctx, query)

	if err != nil {
		return err
	}

	if err := q.Select(dest, arg); err != nil {
		return err
	}

	return nil
}

func (db *Database) Find(ctx context.Context, query string, dest any, args ...any) error {
	q, err := db.db.PreparexContext(ctx, query)

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

func (db *Database) FindNamed(ctx context.Context, query string, dest any, arg any) error {
	q, err := db.db.PrepareNamedContext(ctx, query)

	if err != nil {
		return err
	}

	if err := q.GetContext(ctx, dest, arg); err != nil {
		return err
	}

	return nil
}

func (db *Database) transaction(ctx context.Context, fn func(db *Database) error) error {
	if _, ok := db.db.(*sqlx.Tx); ok {
		return contract.ErrDatabaseNestedTransaction
	}

	tx, err := db.raw.BeginTxx(ctx, &sql.TxOptions{})

	if err != nil {
		return err
	}

	txWrapper := &Database{db: tx, raw: db.raw}

	if err := fn(txWrapper); err != nil {
		return errors.Join(err, tx.Rollback())
	}

	return tx.Commit()
}

func (db *Database) WithTransaction(ctx context.Context, fn func(tx contract.Database) error) error {
	return db.transaction(ctx, func(db *Database) error {
		return fn(db)
	})
}

func (db *Database) WithNamedTransaction(ctx context.Context, fn func(tx contract.DatabaseNamed) error) error {
	return db.transaction(ctx, func(db *Database) error {
		return fn(db)
	})
}
