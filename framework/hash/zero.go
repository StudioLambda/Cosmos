package hash

import "runtime"

// zeroBytes overwrites every byte in the slice with zero.
// This is used to clear sensitive data such as passwords
// from memory after hashing or verification is complete.
// The runtime.KeepAlive call ensures the clear operation
// is not optimized away by the compiler.
func zeroBytes(value []byte) {
	clear(value)
	runtime.KeepAlive(&value)
}
