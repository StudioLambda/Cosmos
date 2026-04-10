package hash

// zeroBytes overwrites every byte in the slice with zero.
// This is used to clear sensitive data such as passwords
// from memory after hashing or verification is complete.
func zeroBytes(value []byte) {
	for i := range value {
		value[i] = 0
	}
}
