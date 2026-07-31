# framework/hash

Password hashing drivers implementing `contract.HasherDriver`.

## Implementations

- `hash/argon2` (default recommendation)
- `hash/bcrypt` (compatibility-oriented)

## Security notes

- Use Argon2 for new systems unless interoperability constraints require bcrypt.
- Avoid logging raw passwords or hashes.
- The raw byte slice passed to driver `Hash` or `Check` is cleared; do not reuse it.
- Wrap a driver with `contract.NewHasher` for JSON-encoded typed values, or use
  `HashRaw` and `CheckRaw` for password bytes.
