package query

import (
	"slices"
	"strconv"
	"strings"
)

// SelectQuery represents an immutable SELECT query builder.
// Each method returns a new instance without modifying the original.
type SelectQuery struct {
	table   string
	columns []any
	joins   []string
	where   whereClauses
	groupBy []string
	orderBy []string
	limit   int
	offset  int
}

// Select creates a new [SelectQuery] for the given table.
func Select(table string) SelectQuery {
	return SelectQuery{table: table}
}

// Columns sets the columns to select. Values can be strings or [Raw]/[Expr]
// for raw SQL expressions.
func (query SelectQuery) Columns(columns ...any) SelectQuery {
	query.columns = slices.Clone(columns)

	return query
}

// Column appends a single column to the select list. The value can be
// a string or [Raw]/[Expr] for raw SQL expressions.
func (query SelectQuery) Column(column any) SelectQuery {
	query.columns = append(slices.Clone(query.columns), column)

	return query
}

// Join appends a JOIN clause to the query. The clause should be a
// complete join expression (e.g., "posts ON posts.user_id = users.id").
func (query SelectQuery) Join(join string) SelectQuery {
	query.joins = append(slices.Clone(query.joins), join)

	return query
}

// Where appends an AND WHERE condition.
func (query SelectQuery) Where(sql string, args ...any) SelectQuery {
	query.where = query.where.add("AND", sql, args)

	return query
}

// OrWhere appends an OR WHERE condition.
func (query SelectQuery) OrWhere(sql string, args ...any) SelectQuery {
	query.where = query.where.add("OR", sql, args)

	return query
}

// GroupBy sets the GROUP BY columns.
func (query SelectQuery) GroupBy(columns ...string) SelectQuery {
	query.groupBy = slices.Clone(columns)

	return query
}

// OrderBy sets the ORDER BY clauses.
func (query SelectQuery) OrderBy(clauses ...string) SelectQuery {
	query.orderBy = slices.Clone(clauses)

	return query
}

// Limit sets the maximum number of rows to return.
// A value of 0 means no limit.
func (query SelectQuery) Limit(limit int) SelectQuery {
	query.limit = limit

	return query
}

// Offset sets the number of rows to skip.
// A value of 0 means no offset.
func (query SelectQuery) Offset(offset int) SelectQuery {
	query.offset = offset

	return query
}

// Build produces the SQL query string and its positional arguments.
func (query SelectQuery) Build() (string, []any) {
	var builder strings.Builder
	var args []any

	builder.WriteString("SELECT ")

	if len(query.columns) == 0 {
		builder.WriteByte('*')
	} else {
		for i, col := range query.columns {
			if i > 0 {
				builder.WriteString(", ")
			}

			renderColumn(&builder, &args, col)
		}
	}

	builder.WriteString(" FROM ")
	builder.WriteString(query.table)

	for _, join := range query.joins {
		builder.WriteString(" JOIN ")
		builder.WriteString(join)
	}

	query.where.render(&builder, &args)

	if len(query.groupBy) > 0 {
		builder.WriteString(" GROUP BY ")
		builder.WriteString(strings.Join(query.groupBy, ", "))
	}

	if len(query.orderBy) > 0 {
		builder.WriteString(" ORDER BY ")
		builder.WriteString(strings.Join(query.orderBy, ", "))
	}

	if query.limit > 0 {
		builder.WriteString(" LIMIT ")
		builder.WriteString(strconv.Itoa(query.limit))
	}

	if query.offset > 0 {
		builder.WriteString(" OFFSET ")
		builder.WriteString(strconv.Itoa(query.offset))
	}

	return builder.String(), args
}
