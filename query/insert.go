package query

import (
	"slices"
	"strings"
)

// insertPair holds a column name and its associated value.
type insertPair struct {
	column string
	value  any
}

// InsertQuery represents an immutable INSERT query builder.
// Each method returns a new instance without modifying the original.
type InsertQuery struct {
	table string
	pairs []insertPair
}

// Insert creates a new [InsertQuery] for the given table.
func Insert(table string) InsertQuery {
	return InsertQuery{table: table}
}

// Set adds a column-value pair to the INSERT statement. The value can
// be a plain value (emits ?), [Raw] (emits literal SQL), or [Expr]
// (emits SQL with embedded placeholders).
func (query InsertQuery) Set(column string, value any) InsertQuery {
	query.pairs = append(slices.Clone(query.pairs), insertPair{column: column, value: value})

	return query
}

// Build produces the SQL query string and its positional arguments.
func (query InsertQuery) Build() (string, []any) {
	var builder strings.Builder
	var args []any

	builder.WriteString("INSERT INTO ")
	builder.WriteString(query.table)
	builder.WriteString(" (")

	for i, pair := range query.pairs {
		if i > 0 {
			builder.WriteString(", ")
		}

		builder.WriteString(pair.column)
	}

	builder.WriteString(") VALUES (")

	for i, pair := range query.pairs {
		if i > 0 {
			builder.WriteString(", ")
		}

		renderValue(&builder, &args, pair.value)
	}

	builder.WriteByte(')')

	return builder.String(), args
}
