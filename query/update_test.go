package query_test

import (
	"testing"

	"github.com/studiolambda/cosmos/query"
)

func TestUpdateBasic(t *testing.T) {
	t.Parallel()

	sql, args := query.Update("users").
		Set("name", "Erik").
		Where("id = ?", 1).
		Build()

	if sql != "UPDATE users SET name = ? WHERE id = ?" {
		t.Errorf("unexpected sql: %s", sql)
	}

	if len(args) != 2 || args[0] != "Erik" || args[1] != 1 {
		t.Errorf("unexpected args: %v", args)
	}
}

func TestUpdateRaw(t *testing.T) {
	t.Parallel()

	sql, args := query.Update("users").
		Set("name", "Erik").
		Set("updated_at", query.Raw("NOW()")).
		Where("id = ?", 1).
		Build()

	if sql != "UPDATE users SET name = ?, updated_at = NOW() WHERE id = ?" {
		t.Errorf("unexpected sql: %s", sql)
	}

	if len(args) != 2 || args[0] != "Erik" || args[1] != 1 {
		t.Errorf("unexpected args: %v", args)
	}
}

func TestUpdateOrWhere(t *testing.T) {
	t.Parallel()

	sql, args := query.Update("users").
		Set("active", false).
		Where("banned = ?", true).
		OrWhere("expired = ?", true).
		Build()

	expected := "UPDATE users SET active = ? WHERE banned = ? OR expired = ?"

	if sql != expected {
		t.Errorf("unexpected sql: %s", sql)
	}

	if len(args) != 3 || args[0] != false || args[1] != true || args[2] != true {
		t.Errorf("unexpected args: %v", args)
	}
}

func TestUpdateImmutability(t *testing.T) {
	t.Parallel()

	base := query.Update("users").Set("updated_at", query.Raw("NOW()"))
	first := base.Set("name", "A").Where("id = ?", 1)
	second := base.Set("name", "B").Where("id = ?", 2)

	_, argsFirst := first.Build()
	_, argsSecond := second.Build()

	if argsFirst[0] != "A" || argsFirst[1] != 1 {
		t.Errorf("unexpected first args: %v", argsFirst)
	}

	if argsSecond[0] != "B" || argsSecond[1] != 2 {
		t.Errorf("unexpected second args: %v", argsSecond)
	}
}
