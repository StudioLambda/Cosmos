// Package contract defines the core interfaces and shared types used across
// Cosmos modules.
//
// The package is dependency-light and focuses on stable abstraction boundaries:
// cache, database, session, cryptography, hashing, eventing, and response hooks.
// It is intended for dependency injection and adapter-style implementations.
//
// Architecture
//
//	contract (interfaces and helper types)
//	├── request   (HTTP request helpers)
//	├── response  (HTTP response helpers)
//	└── mock      (generated test doubles)
//
// # Thread safety
//
// Session values created with [NewSession] / [NewSessionFrom] are safe for
// concurrent access. The session implementation guards state with a mutex and
// returns defensive copies where needed (for example, [Session.All]).
//
// # Initialization behavior
//
// Most interfaces are implemented in sibling modules (for example,
// github.com/studiolambda/cosmos/framework). This package itself does not
// require global initialization.
//
// Example
//
//	func rememberProfile(ctx context.Context, cache *contract.Cache, userID string) (Profile, error) {
//		return cache.Remember(ctx, "user:"+userID, 5*time.Minute, func() (Profile, error) {
//			return loadProfile(userID)
//		})
//	}
package contract
