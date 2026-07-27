# framework/database

SQL database drivers implementing `contract.DatabaseDriver`.

## What this package provides

- `database/postgres` uses pure-Go pgx.
- `database/mysql` uses go-sql-driver/mysql.
- `database/sqlite` uses pure-Go modernc.org/sqlite.

## When to use it

Import one concrete driver package and wrap it with `contract.NewDatabase`.
