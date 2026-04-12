//go:build cgo

package database_test

import (
	"context"
	"errors"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/require"
	"github.com/studiolambda/cosmos/contract"
	"github.com/studiolambda/cosmos/framework/database"
)

func newTestDB(t *testing.T) *database.SQL {
	t.Helper()

	db, err := database.NewSQL("sqlite3", ":memory:")

	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, db.Close())
	})

	return db
}

func TestWithTransactionPanicRollsBack(t *testing.T) {
	t.Parallel()

	db := newTestDB(t)
	ctx := context.Background()

	_, err := db.Exec(ctx, "CREATE TABLE items (id INTEGER PRIMARY KEY, name TEXT)")

	require.NoError(t, err)

	require.Panics(t, func() {
		_ = db.WithTransaction(ctx, func(tx contract.Database) error {
			_, err := tx.Exec(ctx, "INSERT INTO items (name) VALUES (?)", "should-be-rolled-back")

			if err != nil {
				return err
			}

			panic("unexpected failure")
		})
	})

	// The row must not exist because the transaction was rolled back.
	var count int
	err = db.Find(ctx, "SELECT COUNT(*) FROM items", &count)

	require.NoError(t, err)
	require.Equal(t, 0, count)
}

func TestWithTransactionNestedReturnsError(t *testing.T) {
	t.Parallel()

	db := newTestDB(t)
	ctx := context.Background()

	err := db.WithTransaction(ctx, func(tx contract.Database) error {
		return tx.WithTransaction(ctx, func(inner contract.Database) error {
			return nil
		})
	})

	require.ErrorIs(t, err, contract.ErrDatabaseNestedTransaction)
}

func TestWithTransactionCommitsOnSuccess(t *testing.T) {
	t.Parallel()

	db := newTestDB(t)
	ctx := context.Background()

	_, err := db.Exec(ctx, "CREATE TABLE items (id INTEGER PRIMARY KEY, name TEXT)")

	require.NoError(t, err)

	err = db.WithTransaction(ctx, func(tx contract.Database) error {
		_, err := tx.Exec(ctx, "INSERT INTO items (name) VALUES (?)", "committed")

		return err
	})

	require.NoError(t, err)

	var count int
	err = db.Find(ctx, "SELECT COUNT(*) FROM items", &count)

	require.NoError(t, err)
	require.Equal(t, 1, count)
}

func TestWithTransactionRollsBackOnError(t *testing.T) {
	t.Parallel()

	db := newTestDB(t)
	ctx := context.Background()

	_, err := db.Exec(ctx, "CREATE TABLE items (id INTEGER PRIMARY KEY, name TEXT)")

	require.NoError(t, err)

	callbackErr := errors.New("something went wrong")

	err = db.WithTransaction(ctx, func(tx contract.Database) error {
		_, err := tx.Exec(ctx, "INSERT INTO items (name) VALUES (?)", "should-be-rolled-back")

		if err != nil {
			return err
		}

		return callbackErr
	})

	require.ErrorIs(t, err, callbackErr)

	var count int
	err = db.Find(ctx, "SELECT COUNT(*) FROM items", &count)

	require.NoError(t, err)
	require.Equal(t, 0, count)
}

func TestWithTransactionPanicPreservesValue(t *testing.T) {
	t.Parallel()

	db := newTestDB(t)
	ctx := context.Background()

	var recovered any

	func() {
		defer func() {
			recovered = recover()
		}()

		_ = db.WithTransaction(ctx, func(tx contract.Database) error {
			panic("test-panic-value")
		})
	}()

	require.Equal(t, "test-panic-value", recovered)
}
