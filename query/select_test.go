package query_test

import (
	"testing"

	"github.com/studiolambda/cosmos/query"
)

func TestSelectAll(t *testing.T) {
	t.Parallel()

	sql, args := query.Select("users").Build()

	if sql != "SELECT * FROM users" {
		t.Errorf("unexpected sql: %s", sql)
	}

	if len(args) != 0 {
		t.Errorf("unexpected args: %v", args)
	}
}

func TestSelectColumns(t *testing.T) {
	t.Parallel()

	sql, args := query.Select("users").
		Columns("id", "name", "email").
		Build()

	if sql != "SELECT id, name, email FROM users" {
		t.Errorf("unexpected sql: %s", sql)
	}

	if len(args) != 0 {
		t.Errorf("unexpected args: %v", args)
	}
}

func TestSelectColumnAppend(t *testing.T) {
	t.Parallel()

	sql, _ := query.Select("users").
		Columns("id").
		Column("name").
		Build()

	if sql != "SELECT id, name FROM users" {
		t.Errorf("unexpected sql: %s", sql)
	}
}

func TestSelectColumnRaw(t *testing.T) {
	t.Parallel()

	sql, args := query.Select("orders").
		Columns("user_id", query.Raw("COUNT(*) AS total")).
		Build()

	if sql != "SELECT user_id, COUNT(*) AS total FROM orders" {
		t.Errorf("unexpected sql: %s", sql)
	}

	if len(args) != 0 {
		t.Errorf("unexpected args: %v", args)
	}
}

func TestSelectColumnExpr(t *testing.T) {
	t.Parallel()

	sql, args := query.Select("orders").
		Column(query.Expr{SQL: "DATE_TRUNC(?, created_at) AS period", Args: []any{"month"}}).
		Build()

	if sql != "SELECT DATE_TRUNC(?, created_at) AS period FROM orders" {
		t.Errorf("unexpected sql: %s", sql)
	}

	if len(args) != 1 || args[0] != "month" {
		t.Errorf("unexpected args: %v", args)
	}
}

func TestSelectWhere(t *testing.T) {
	t.Parallel()

	sql, args := query.Select("users").
		Columns("id", "name").
		Where("active = ?", true).
		Where("age > ?", 18).
		Build()

	if sql != "SELECT id, name FROM users WHERE active = ? AND age > ?" {
		t.Errorf("unexpected sql: %s", sql)
	}

	if len(args) != 2 || args[0] != true || args[1] != 18 {
		t.Errorf("unexpected args: %v", args)
	}
}

func TestSelectOrWhere(t *testing.T) {
	t.Parallel()

	sql, args := query.Select("users").
		Where("role = ?", "admin").
		OrWhere("role = ?", "editor").
		Build()

	if sql != "SELECT * FROM users WHERE role = ? OR role = ?" {
		t.Errorf("unexpected sql: %s", sql)
	}

	if len(args) != 2 || args[0] != "admin" || args[1] != "editor" {
		t.Errorf("unexpected args: %v", args)
	}
}

func TestSelectJoin(t *testing.T) {
	t.Parallel()

	sql, _ := query.Select("users").
		Columns("users.id", "posts.title").
		Join("posts ON posts.user_id = users.id").
		Build()

	expected := "SELECT users.id, posts.title FROM users JOIN posts ON posts.user_id = users.id"

	if sql != expected {
		t.Errorf("unexpected sql: %s", sql)
	}
}

func TestSelectGroupBy(t *testing.T) {
	t.Parallel()

	sql, _ := query.Select("orders").
		Columns("user_id", query.Raw("SUM(amount) AS total")).
		GroupBy("user_id").
		Build()

	expected := "SELECT user_id, SUM(amount) AS total FROM orders GROUP BY user_id"

	if sql != expected {
		t.Errorf("unexpected sql: %s", sql)
	}
}

func TestSelectOrderBy(t *testing.T) {
	t.Parallel()

	sql, _ := query.Select("users").
		OrderBy("name ASC", "id DESC").
		Build()

	expected := "SELECT * FROM users ORDER BY name ASC, id DESC"

	if sql != expected {
		t.Errorf("unexpected sql: %s", sql)
	}
}

func TestSelectLimitOffset(t *testing.T) {
	t.Parallel()

	sql, _ := query.Select("users").
		Limit(10).
		Offset(20).
		Build()

	if sql != "SELECT * FROM users LIMIT 10 OFFSET 20" {
		t.Errorf("unexpected sql: %s", sql)
	}
}

func TestSelectImmutability(t *testing.T) {
	t.Parallel()

	base := query.Select("users").Columns("id", "name").Where("active = ?", true)
	admins := base.Where("role = ?", "admin")
	editors := base.Where("role = ?", "editor")

	sqlAdmins, argsAdmins := admins.Build()
	sqlEditors, argsEditors := editors.Build()

	if sqlAdmins != "SELECT id, name FROM users WHERE active = ? AND role = ?" {
		t.Errorf("unexpected admin sql: %s", sqlAdmins)
	}

	if sqlEditors != "SELECT id, name FROM users WHERE active = ? AND role = ?" {
		t.Errorf("unexpected editor sql: %s", sqlEditors)
	}

	if argsAdmins[1] != "admin" {
		t.Errorf("unexpected admin args: %v", argsAdmins)
	}

	if argsEditors[1] != "editor" {
		t.Errorf("unexpected editor args: %v", argsEditors)
	}
}

func TestSelectFull(t *testing.T) {
	t.Parallel()

	sql, args := query.Select("users").
		Columns("users.id", "users.name", query.Raw("COUNT(posts.id) AS post_count")).
		Join("posts ON posts.user_id = users.id").
		Where("users.active = ?", true).
		GroupBy("users.id", "users.name").
		OrderBy("post_count DESC").
		Limit(5).
		Build()

	expected := "SELECT users.id, users.name, COUNT(posts.id) AS post_count FROM users JOIN posts ON posts.user_id = users.id WHERE users.active = ? GROUP BY users.id, users.name ORDER BY post_count DESC LIMIT 5"

	if sql != expected {
		t.Errorf("unexpected sql: %s", sql)
	}

	if len(args) != 1 || args[0] != true {
		t.Errorf("unexpected args: %v", args)
	}
}
