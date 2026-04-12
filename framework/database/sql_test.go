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
