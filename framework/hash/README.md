# framework/hash

Password hashing implementations for Cosmos contracts.

## Implementations

- Argon2 (default recommendation)
- Bcrypt (compatibility-oriented)

## Security notes

- Use Argon2 for new systems unless interoperability constraints require bcrypt.
- Avoid logging raw passwords or hashes.
- Memory buffers holding sensitive data are explicitly zeroed where possible.
