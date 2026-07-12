// Package cache contains framework cache-driver implementations.
//
// Drivers satisfy [contract.CacheDriver] and are intended to be wrapped by
// [contract.NewCache] when using typed cache operations.
//
// # Concurrency
//
// Memory and Redis drivers are safe for concurrent request workloads.
//
// Example
//
//	driver := cache.NewMemory(5*time.Minute, 10*time.Minute)
//	store := contract.NewCache(driver)
//	_ = store.Put(ctx, "health", "ok", time.Minute)
package cache
