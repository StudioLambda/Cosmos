// Package hash provides password hashing adapters.
//
// It includes Argon2 and bcrypt implementations that satisfy contract hashing
// abstractions and return verification-friendly errors.
//
// # Operational guidance
//
// Prefer Argon2 for modern deployments. Bcrypt remains available where legacy
// compatibility is needed.
//
// Example
//
//	hasher := hash.NewArgon2()
//	digest, err := hasher.Hash(ctx, []byte("correct horse battery staple"))
//	if err != nil {
//		return err
//	}
//
//	return hasher.Verify(ctx, []byte("correct horse battery staple"), digest)
package hash
