// Package session provides session middleware and drivers for framework apps.
//
// It bridges HTTP cookies to contract.Session values and persists session data
// through a configurable driver (for example, cache-backed storage).
//
// # Lifecycle
//
// Middleware loads the session at request start, injects it into request
// context, and writes updates at response completion when state changes.
//
// Example
//
//	driver := session.NewCacheDriver(
//		contract.NewCache(cache.NewMemory(cache.MemoryConfig{Expiration: 5 * time.Minute, Cleanup: 10 * time.Minute})),
//		session.DefaultCacheDriverConfig,
//	)
//	app.Use(session.Middleware(driver, session.DefaultMiddlewareConfig))
//
//	app.Post("/login", func(w http.ResponseWriter, r *http.Request) error {
//		sess := request.MustSession(r)
//		sess.Put("user_id", 42)
//		return sess.Regenerate()
//	})
package session
