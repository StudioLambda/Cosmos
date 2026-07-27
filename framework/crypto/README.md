# framework/crypto

Authenticated encryption adapters implementing `contract.Encrypter`.

## Implementations

- `crypto/aes`: AES-GCM.
- `crypto/chacha20`: ChaCha20-Poly1305.

## Security notes

- Prefer authenticated modes only (provided by this package).
- Use `AdditionalData` to bind ciphertexts to context when needed.
- Call `Close()` to zero in-memory key material after use.
