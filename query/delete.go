package query

import "strings"

// DeleteQuery represents an immutable DELETE query builder.
// Each method returns a new instance without modifying the original.
type DeleteQuery struct {
	table string
	where whereClauses
}

// Delete creates a new [DeleteQuery] for the given table.
func Delete(table string) DeleteQuery {
	return DeleteQuery{table: table}
}

// Where appends an AND WHERE condition.
func (query DeleteQuery) Where(sql string, args ...any) DeleteQuery {
	query.where = query.where.add("AND", sql, args)

	return query
}

// OrWhere appends an OR WHERE condition.
func (query DeleteQuery) OrWhere(sql string, args ...any) DeleteQuery {
	query.where = query.where.add("OR", sql, args)

	return query
}

// Build produces the SQL query string and its positional arguments.
func (query DeleteQuery) Build() (string, []any) {
	var builder strings.Builder
	var args []any

	builder.WriteString("DELETE FROM ")
	builder.WriteString(query.table)

	query.where.render(&builder, &args)

	return builder.String(), args
}
