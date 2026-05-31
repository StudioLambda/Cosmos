package query

import "strings"

// Raw represents a raw SQL fragment that is embedded directly into
// the query without a placeholder. Use this for SQL functions and
// expressions like NOW(), COUNT(*), etc.
type Raw string

// Expr represents a raw SQL fragment with embedded placeholders
// and their corresponding arguments. Use this for expressions that
// require parameterized values within raw SQL.
type Expr struct {
	SQL  string
	Args []any
}

// renderValue writes a single value to the builder. If the value is
// a [Raw], its SQL is written directly. If it is an [Expr], its SQL
// is written and its args are appended. Otherwise, a ? placeholder
// is written and the value is appended to args.
func renderValue(builder *strings.Builder, args *[]any, value any) {
	switch v := value.(type) {
	case Raw:
		builder.WriteString(string(v))
	case Expr:
		builder.WriteString(v.SQL)
		*args = append(*args, v.Args...)
	default:
		builder.WriteByte('?')
		*args = append(*args, value)
	}
}

// renderColumn writes a column reference to the builder. If the value
// is a [Raw], its SQL is written directly. If it is an [Expr], its SQL
// is written and its args are appended. Otherwise, the value is written
// as a plain string column name.
func renderColumn(builder *strings.Builder, args *[]any, column any) {
	switch v := column.(type) {
	case Raw:
		builder.WriteString(string(v))
	case Expr:
		builder.WriteString(v.SQL)
		*args = append(*args, v.Args...)
	case string:
		builder.WriteString(v)
	}
}
