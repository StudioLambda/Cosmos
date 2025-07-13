package sql

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
	"github.com/studiolambda/cosmos/contract"
)

type Database struct {
	db  sqlx.ExtContext // can be *sqlx.DB or *sqlx.Tx
	raw *sqlx.DB        // needed for transactions
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
	result, err := db.db.ExecContext(ctx, query, args...)

	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}

func (db *Database) Query(ctx context.Context, dest any, query string, args ...any) error {
	return sqlx.SelectContext(ctx, db.db, dest, query, args...)
}

func (db *Database) QueryOne(ctx context.Context, dest any, query string, args ...any) error {
	err := sqlx.GetContext(ctx, db.db, dest, query, args...)

	if errors.Is(err, sql.ErrNoRows) {
		return errors.Join(err, contract.ErrDatabaseNoRows)
	}

	return err
}

func (db *Database) WithTransaction(ctx context.Context, fn func(tx contract.Database) error) error {
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
