// Package framework provides Cosmos' top-level HTTP application primitives.
//
// It composes router, problem-details error handling, and contract abstractions
// into an error-returning handler model that remains compatible with net/http.
//
// # Core model
//
// Handlers return errors:
//
//	type Handler func(http.ResponseWriter, *http.Request) error
//
// This enables centralized error mapping while preserving familiar routing and
// middleware composition patterns.
//
// # Thread safety
//
// Framework routing and middleware setup is expected during startup. Runtime
// request processing is concurrent. Mutable per-request concerns should live in
// request context or dedicated synchronized components.
//
// # Initialization behavior
//
// Use [New] to create an application router and [NewServer] to run with secure
// timeout defaults.
//
// Example
//
//	app := framework.New()
//	app.Get("/health", func(w http.ResponseWriter, r *http.Request) error {
//		return response.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
//	})
//	server := framework.NewServer(":8080", app)
//	_ = server.ListenAndServe()
package framework
