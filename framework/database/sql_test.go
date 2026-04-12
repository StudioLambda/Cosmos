//go:build cgo

package database_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/require"
	"github.com/studiolambda/cosmos/contract"
	"github.com/studiolambda/cosmos/framework/database"
)

func openTestDB(t *testing.T) *database.SQL {
	t.Helper()

	raw, err := sqlx.Connect("sqlite3", ":memory:")
	require.NoError(t, err)

	db := database.NewSQLFrom(raw)
	t.Cleanup(func() { db.Close() })

	_, err = raw.Exec(`CREATE TABLE users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		email TEXT NOT NULL
	)`)
	require.NoError(t, err)

	return db
}

func TestPingSucceeds(t *testing.T) {
	t.Parallel()

	db := openTestDB(t)

	err := db.Ping(context.Background())

	require.NoError(t, err)
}

func TestExecInsertsRow(t *testing.T) {
	t.Parallel()

	db := openTestDB(t)
	ctx := context.Background()

	affected, err := db.Exec(ctx, "INSERT INTO users (name, email) VALUES (?, ?)", "alice", "alice@example.com")

	require.NoError(t, err)
	require.Equal(t, int64(1), affected)
}

func TestExecNamedInsertsRow(t *testing.T) {
	t.Parallel()

	db := openTestDB(t)
	ctx := context.Background()

	arg := map[string]any{"name": "bob", "email": "bob@example.com"}
	affected, err := db.ExecNamed(ctx, "INSERT INTO users (name, email) VALUES (:name, :email)", arg)

	require.NoError(t, err)
	require.Equal(t, int64(1), affected)
}

func TestSelectReturnsMultipleRows(t *testing.T) {
	t.Parallel()

	db := openTestDB(t)
	ctx := context.Background()

	_, err := db.Exec(ctx, "INSERT INTO users (name, email) VALUES (?, ?)", "alice", "a@x.com")
	require.NoError(t, err)

	_, err = db.Exec(ctx, "INSERT INTO users (name, email) VALUES (?, ?)", "bob", "b@x.com")
	require.NoError(t, err)

	type user struct {
		ID    int    `db:"id"`
		Name  string `db:"name"`
		Email string `db:"email"`
	}

	var users []user
	err = db.Select(ctx, "SELECT id, name, email FROM users ORDER BY id", &users)

	require.NoError(t, err)
	require.Len(t, users, 2)
	require.Equal(t, "alice", users[0].Name)
	require.Equal(t, "bob", users[1].Name)
}

func TestSelectNamedReturnsMultipleRows(t *testing.T) {
	t.Parallel()

	db := openTestDB(t)
	ctx := context.Background()

	_, err := db.Exec(ctx, "INSERT INTO users (name, email) VALUES (?, ?)", "alice", "a@x.com")
	require.NoError(t, err)

	_, err = db.Exec(ctx, "INSERT INTO users (name, email) VALUES (?, ?)", "bob", "b@x.com")
	require.NoError(t, err)

	type user struct {
		ID    int    `db:"id"`
		Name  string `db:"name"`
		Email string `db:"email"`
	}

	var users []user
	arg := map[string]any{"domain": "%x.com"}
	err = db.SelectNamed(ctx, "SELECT id, name, email FROM users WHERE email LIKE :domain ORDER BY id", &users, arg)

	require.NoError(t, err)
	require.Len(t, users, 2)
	require.Equal(t, "alice", users[0].Name)
}

func TestFindReturnsSingleRow(t *testing.T) {
	t.Parallel()

	db := openTestDB(t)
	ctx := context.Background()

	_, err := db.Exec(ctx, "INSERT INTO users (name, email) VALUES (?, ?)", "alice", "a@x.com")
	require.NoError(t, err)

	type user struct {
		ID    int    `db:"id"`
		Name  string `db:"name"`
		Email string `db:"email"`
	}

	var found user
	err = db.Find(ctx, "SELECT id, name, email FROM users WHERE name = ?", &found, "alice")

	require.NoError(t, err)
	require.Equal(t, "alice", found.Name)
}

func TestFindWrapsErrNoRows(t *testing.T) {
	t.Parallel()

	db := openTestDB(t)
	ctx := context.Background()

	type user struct {
		ID   int    `db:"id"`
		Name string `db:"name"`
	}

	var found user
	err := db.Find(ctx, "SELECT id, name FROM users WHERE name = ?", &found, "ghost")

	require.Error(t, err)
	require.True(t, errors.Is(err, sql.ErrNoRows))
	require.True(t, errors.Is(err, contract.ErrDatabaseNoRows))
}

func TestFindNamedReturnsSingleRow(t *testing.T) {
	t.Parallel()

	db := openTestDB(t)
	ctx := context.Background()

	_, err := db.Exec(ctx, "INSERT INTO users (name, email) VALUES (?, ?)", "alice", "a@x.com")
	require.NoError(t, err)

	type user struct {
		ID    int    `db:"id"`
		Name  string `db:"name"`
		Email string `db:"email"`
	}

	var found user
	arg := map[string]any{"name": "alice"}
	err = db.FindNamed(ctx, "SELECT id, name, email FROM users WHERE name = :name", &found, arg)

	require.NoError(t, err)
	require.Equal(t, "alice", found.Name)
}

func TestFindNamedWrapsErrNoRows(t *testing.T) {
	t.Parallel()

	db := openTestDB(t)
	ctx := context.Background()

	type user struct {
		ID   int    `db:"id"`
		Name string `db:"name"`
	}

	var found user
	arg := map[string]any{"name": "ghost"}
	err := db.FindNamed(ctx, "SELECT id, name FROM users WHERE name = :name", &found, arg)

	require.Error(t, err)
	require.True(t, errors.Is(err, sql.ErrNoRows))
	require.True(t, errors.Is(err, contract.ErrDatabaseNoRows))
}

func TestWithTransactionCommitsOnSuccess(t *testing.T) {
	t.Parallel()

	db := openTestDB(t)
	ctx := context.Background()

	err := db.WithTransaction(ctx, func(tx contract.Database) error {
		_, err := tx.Exec(ctx, "INSERT INTO users (name, email) VALUES (?, ?)", "alice", "a@x.com")
		return err
	})
	require.NoError(t, err)

	type user struct {
		ID   int    `db:"id"`
		Name string `db:"name"`
	}

	var found user
	err = db.Find(ctx, "SELECT id, name FROM users WHERE name = ?", &found, "alice")

	require.NoError(t, err)
	require.Equal(t, "alice", found.Name)
}

func TestWithTransactionRollsBackOnError(t *testing.T) {
	t.Parallel()

	db := openTestDB(t)
	ctx := context.Background()

	sentinel := errors.New("rollback me")

	err := db.WithTransaction(ctx, func(tx contract.Database) error {
		_, err := tx.Exec(ctx, "INSERT INTO users (name, email) VALUES (?, ?)", "alice", "a@x.com")
		require.NoError(t, err)

		return sentinel
	})
	require.True(t, errors.Is(err, sentinel))

	type user struct {
		ID   int    `db:"id"`
		Name string `db:"name"`
	}

	var users []user
	err = db.Select(ctx, "SELECT id, name FROM users", &users)

	require.NoError(t, err)
	require.Empty(t, users)
}

func TestWithTransactionRejectsNesting(t *testing.T) {
	t.Parallel()

	db := openTestDB(t)
	ctx := context.Background()

	err := db.WithTransaction(ctx, func(tx contract.Database) error {
		return tx.WithTransaction(ctx, func(inner contract.Database) error {
			return nil
		})
	})

	require.True(t, errors.Is(err, contract.ErrDatabaseNestedTransaction))
}

func TestCloseOnTransactionWrapperIsNoOp(t *testing.T) {
	t.Parallel()

	db := openTestDB(t)
	ctx := context.Background()

	err := db.WithTransaction(ctx, func(tx contract.Database) error {
		// Closing the transaction wrapper should be a no-op and
		// must not close the underlying connection pool.
		err := tx.Close()
		require.NoError(t, err)

		// The transaction should still be usable after Close.
		_, err = tx.Exec(ctx, "INSERT INTO users (name, email) VALUES (?, ?)", "alice", "a@x.com")
		return err
	})
	require.NoError(t, err)

	// The pool should still be functional after the transaction completes.
	err = db.Ping(ctx)
	require.NoError(t, err)

	type user struct {
		ID   int    `db:"id"`
		Name string `db:"name"`
	}

	var found user
	err = db.Find(ctx, "SELECT id, name FROM users WHERE name = ?", &found, "alice")

	require.NoError(t, err)
	require.Equal(t, "alice", found.Name)
}

func TestConfigureSetsPoolLimits(t *testing.T) {
	t.Parallel()

	db := openTestDB(t)

	db.Configure(func(raw *sql.DB) {
		raw.SetMaxOpenConns(10)
		raw.SetMaxIdleConns(5)
	})

	err := db.Ping(context.Background())
	require.NoError(t, err)
}
