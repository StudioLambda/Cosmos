// Package query provides an immutable, copy-on-write SQL query builder
// that produces parameterized query strings with positional placeholders (?).
//
// The builder supports SELECT, INSERT, UPDATE, and DELETE statements
// with support for raw SQL expressions via [Raw] and [Expr].
//
// Each method returns a new value without mutating the original,
// making it safe to derive queries from shared base instances:
//
//	base := query.Select("users").Columns("id", "name").Where("active = ?", true)
//	admins := base.Where("role = ?", "admin")
//	editors := base.Where("role = ?", "editor")
//
// The produced queries use ? as the placeholder character. Use your
// database driver's rebind functionality if your database requires
// a different placeholder style (e.g., $1 for PostgreSQL).
package query
