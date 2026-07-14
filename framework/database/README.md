# framework/database

SQL database adapter implementing `contract.Database`.

## What this package provides

- Query and exec helpers with positional and named parameters.
- Transaction helper with nested-transaction protection.
- Connection pool configuration hook.

## When to use it

Use this package when you want a contract-compatible DB abstraction over `sqlx`
with safer defaults for statement lifecycle and transaction boundaries.
