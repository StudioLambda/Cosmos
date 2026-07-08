# framework/crypto

Authenticated encryption adapters implementing `contract.Encrypter`.

## Implementations

- AES-GCM (`NewAES`)
- ChaCha20-Poly1305 (`NewChaCha20`)

## Security notes

- Prefer authenticated modes only (provided by this package).
- Use `AdditionalData` to bind ciphertexts to context when needed.
- Call `Close()` to zero in-memory key material after use.
