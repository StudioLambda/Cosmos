package query_test

import (
	"testing"

	"github.com/studiolambda/cosmos/query"
)

func TestInsertBasic(t *testing.T) {
	t.Parallel()

	sql, args := query.Insert("users").
		Set("name", "Erik").
		Set("email", "erik@example.com").
		Build()

	if sql != "INSERT INTO users (name, email) VALUES (?, ?)" {
		t.Errorf("unexpected sql: %s", sql)
	}

	if len(args) != 2 || args[0] != "Erik" || args[1] != "erik@example.com" {
		t.Errorf("unexpected args: %v", args)
	}
}

func TestInsertRaw(t *testing.T) {
	t.Parallel()

	sql, args := query.Insert("users").
		Set("name", "Erik").
		Set("created_at", query.Raw("NOW()")).
		Build()

	if sql != "INSERT INTO users (name, created_at) VALUES (?, NOW())" {
		t.Errorf("unexpected sql: %s", sql)
	}

	if len(args) != 1 || args[0] != "Erik" {
		t.Errorf("unexpected args: %v", args)
	}
}

func TestInsertExpr(t *testing.T) {
	t.Parallel()

	sql, args := query.Insert("events").
		Set("name", "deploy").
		Set("scheduled_at", query.Expr{SQL: "NOW() + INTERVAL ? HOUR", Args: []any{2}}).
		Build()

	if sql != "INSERT INTO events (name, scheduled_at) VALUES (?, NOW() + INTERVAL ? HOUR)" {
		t.Errorf("unexpected sql: %s", sql)
	}

	if len(args) != 2 || args[0] != "deploy" || args[1] != 2 {
		t.Errorf("unexpected args: %v", args)
	}
}

func TestInsertImmutability(t *testing.T) {
	t.Parallel()

	base := query.Insert("users").Set("role", "user")
	admin := base.Set("name", "Admin")
	editor := base.Set("name", "Editor")

	sqlAdmin, argsAdmin := admin.Build()
	sqlEditor, argsEditor := editor.Build()

	if argsAdmin[1] != "Admin" {
		t.Errorf("unexpected admin args: %v (sql: %s)", argsAdmin, sqlAdmin)
	}

	if argsEditor[1] != "Editor" {
		t.Errorf("unexpected editor args: %v (sql: %s)", argsEditor, sqlEditor)
	}
}
