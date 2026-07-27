# framework/hash

Password hashing implementations for Cosmos contracts.

## Implementations

- `hash/argon2` (default recommendation)
- `hash/bcrypt` (compatibility-oriented)

## Security notes

- Use Argon2 for new systems unless interoperability constraints require bcrypt.
- Avoid logging raw passwords or hashes.
- Memory buffers holding sensitive data are explicitly zeroed where possible.
