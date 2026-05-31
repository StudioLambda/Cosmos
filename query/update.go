package query

import (
	"slices"
	"strings"
)

// updatePair holds a column name and its associated value.
type updatePair struct {
	column string
	value  any
}

// UpdateQuery represents an immutable UPDATE query builder.
// Each method returns a new instance without modifying the original.
type UpdateQuery struct {
	table string
	pairs []updatePair
	where whereClauses
}

// Update creates a new [UpdateQuery] for the given table.
func Update(table string) UpdateQuery {
	return UpdateQuery{table: table}
}

// Set adds a column-value pair to the SET clause. The value can
// be a plain value (emits ?), [Raw] (emits literal SQL), or [Expr]
// (emits SQL with embedded placeholders).
func (query UpdateQuery) Set(column string, value any) UpdateQuery {
	query.pairs = append(slices.Clone(query.pairs), updatePair{column: column, value: value})

	return query
}

// Where appends an AND WHERE condition.
func (query UpdateQuery) Where(sql string, args ...any) UpdateQuery {
	query.where = query.where.add("AND", sql, args)

	return query
}

// OrWhere appends an OR WHERE condition.
func (query UpdateQuery) OrWhere(sql string, args ...any) UpdateQuery {
	query.where = query.where.add("OR", sql, args)

	return query
}

// Build produces the SQL query string and its positional arguments.
func (query UpdateQuery) Build() (string, []any) {
	var builder strings.Builder
	var args []any

	builder.WriteString("UPDATE ")
	builder.WriteString(query.table)
	builder.WriteString(" SET ")

	for i, pair := range query.pairs {
		if i > 0 {
			builder.WriteString(", ")
		}

		builder.WriteString(pair.column)
		builder.WriteString(" = ")
		renderValue(&builder, &args, pair.value)
	}

	query.where.render(&builder, &args)

	return builder.String(), args
}
