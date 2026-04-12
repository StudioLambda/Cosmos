//go:build cgo

package database_test

import (
	"context"
	"errors"
	"testing"

	"github.com/jmoiron/sqlx"
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

func setupTestDB(t *testing.T) *database.SQL {
	t.Helper()

	raw, err := sqlx.Connect("sqlite3", ":memory:")
	require.NoError(t, err)

	t.Cleanup(func() {
		raw.Close()
	})

	db := database.NewSQLFrom(raw)

	_, err = raw.Exec(`CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT NOT NULL)`)
	require.NoError(t, err)

	return db
}

func TestSelectPropagatesContext(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := setupTestDB(t)

	_, err := db.Exec(ctx, `INSERT INTO users (id, name) VALUES (?, ?)`, 1, "alice")
	require.NoError(t, err)

	_, err = db.Exec(ctx, `INSERT INTO users (id, name) VALUES (?, ?)`, 2, "bob")
	require.NoError(t, err)

	type user struct {
		ID   int    `db:"id"`
		Name string `db:"name"`
	}

	var users []user

	err = db.Select(ctx, `SELECT id, name FROM users ORDER BY id`, &users)

	require.NoError(t, err)
	require.Len(t, users, 2)
	require.Equal(t, "alice", users[0].Name)
	require.Equal(t, "bob", users[1].Name)
}

func TestFindNamedReturnsErrDatabaseNoRows(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := setupTestDB(t)

	type user struct {
		ID   int    `db:"id"`
		Name string `db:"name"`
	}

	var found user

	err := db.FindNamed(ctx, `SELECT id, name FROM users WHERE id = :id`, &found, map[string]any{"id": 999})

	require.Error(t, err)
	require.True(t, errors.Is(err, contract.ErrDatabaseNoRows))
}

func TestFindNamedReturnsResultWhenFound(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := setupTestDB(t)

	_, err := db.Exec(ctx, `INSERT INTO users (id, name) VALUES (?, ?)`, 1, "alice")
	require.NoError(t, err)

	type user struct {
		ID   int    `db:"id"`
		Name string `db:"name"`
	}

	var found user

	err = db.FindNamed(ctx, `SELECT id, name FROM users WHERE id = :id`, &found, map[string]any{"id": 1})

	require.NoError(t, err)
	require.Equal(t, 1, found.ID)
	require.Equal(t, "alice", found.Name)
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
