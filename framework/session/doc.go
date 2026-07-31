// Package session provides [contract.SessionDriver] implementations for Cosmos
// applications.
//
// Use middleware.Session to load sessions from cookies, attach them to requests,
// and persist changes through a driver from this package.
//
// Example
//
//	cache := contract.NewCache(memory.NewMemory(memory.MemoryConfig{
//		Expiration: 5 * time.Minute,
//		Cleanup:    10 * time.Minute,
//	}))
//	driver := session.NewCacheDriver(
//		cache,
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
