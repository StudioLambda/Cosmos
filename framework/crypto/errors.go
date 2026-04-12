package crypto

import "errors"

// ErrEncrypterClosed is returned when [AES.Encrypt], [AES.Decrypt],
// [ChaCha20.Encrypt], or [ChaCha20.Decrypt] is called after
// [AES.Close] or [ChaCha20.Close] has been called.
var ErrEncrypterClosed = errors.New("encrypter is closed")
