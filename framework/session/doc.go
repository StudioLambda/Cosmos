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
//	driver := session.NewCacheDriver(*contract.NewCache(cache.NewMemory(5*time.Minute, 10*time.Minute)))
//	app.Use(session.Middleware(driver))
//
//	app.Post("/login", func(w http.ResponseWriter, r *http.Request) error {
//		sess := request.MustSession(r)
//		sess.Put("user_id", 42)
//		return sess.Regenerate()
//	})
package session
