package query

import (
	"slices"
	"strings"
)

// clause represents a single WHERE condition with its connector (AND/OR).
type clause struct {
	connector string
	sql       string
	args      []any
}

// whereClauses is the shared WHERE clause state used by SELECT, UPDATE, and DELETE.
type whereClauses []clause

// clone returns a copy of the where clauses.
func (clauses whereClauses) clone() whereClauses {
	return slices.Clone(clauses)
}

// add appends a new clause with the given connector.
func (clauses whereClauses) add(connector string, sql string, args []any) whereClauses {
	result := clauses.clone()

	return append(result, clause{
		connector: connector,
		sql:       sql,
		args:      args,
	})
}

// render writes the WHERE clause to the builder and appends args.
func (clauses whereClauses) render(builder *strings.Builder, args *[]any) {
	if len(clauses) == 0 {
		return
	}

	builder.WriteString(" WHERE ")

	for i, c := range clauses {
		if i > 0 {
			builder.WriteByte(' ')
			builder.WriteString(c.connector)
			builder.WriteByte(' ')
		}

		builder.WriteString(c.sql)
		*args = append(*args, c.args...)
	}
}
