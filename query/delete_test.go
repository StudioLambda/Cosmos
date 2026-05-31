package query_test

import (
	"testing"

	"github.com/studiolambda/cosmos/query"
)

func TestDeleteBasic(t *testing.T) {
	t.Parallel()

	sql, args := query.Delete("users").
		Where("id = ?", 1).
		Build()

	if sql != "DELETE FROM users WHERE id = ?" {
		t.Errorf("unexpected sql: %s", sql)
	}

	if len(args) != 1 || args[0] != 1 {
		t.Errorf("unexpected args: %v", args)
	}
}

func TestDeleteOrWhere(t *testing.T) {
	t.Parallel()

	sql, args := query.Delete("sessions").
		Where("expired = ?", true).
		OrWhere("revoked = ?", true).
		Build()

	expected := "DELETE FROM sessions WHERE expired = ? OR revoked = ?"

	if sql != expected {
		t.Errorf("unexpected sql: %s", sql)
	}

	if len(args) != 2 || args[0] != true || args[1] != true {
		t.Errorf("unexpected args: %v", args)
	}
}

func TestDeleteNoWhere(t *testing.T) {
	t.Parallel()

	sql, args := query.Delete("logs").Build()

	if sql != "DELETE FROM logs" {
		t.Errorf("unexpected sql: %s", sql)
	}

	if len(args) != 0 {
		t.Errorf("unexpected args: %v", args)
	}
}

func TestDeleteImmutability(t *testing.T) {
	t.Parallel()

	base := query.Delete("users")
	first := base.Where("id = ?", 1)
	second := base.Where("id = ?", 2)

	_, argsFirst := first.Build()
	_, argsSecond := second.Build()

	if argsFirst[0] != 1 {
		t.Errorf("unexpected first args: %v", argsFirst)
	}

	if argsSecond[0] != 2 {
		t.Errorf("unexpected second args: %v", argsSecond)
	}
}
