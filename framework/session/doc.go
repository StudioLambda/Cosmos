// Package session provides session drivers for framework apps.
//
// It bridges HTTP cookies to contract.Session values and persists session data
// through a configurable driver (for example, cache-backed storage).
//
// Example
//
//	driver := session.NewCacheDriver(
//		contract.NewCache(cache.NewMemory(cache.MemoryConfig{Expiration: 5 * time.Minute, Cleanup: 10 * time.Minute})),
//		session.DefaultCacheDriverConfig,
//	)
//	app.Use(middleware.Session(driver, middleware.DefaultSessionConfig))
//
//	app.Post("/login", func(w http.ResponseWriter, r *http.Request) error {
//		sess := request.MustSession(r)
//		sess.Put("user_id", 42)
//		sess.Regenerate()
//		return nil
//	})
package session
