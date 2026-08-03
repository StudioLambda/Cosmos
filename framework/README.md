# Cosmos: Framework

`framework` combines Cosmos routing, problem details, contracts, and integrations into an error-returning HTTP application framework.

```bash
go get github.com/studiolambda/cosmos/framework
```

Requires Go 1.27 or later.

## Application

```go
app := framework.New()
app.Use(middleware.Recover())
app.Get("/users/{id}", func(w http.ResponseWriter, r *http.Request) error {
	return response.JSON(w, http.StatusOK, map[string]string{"id": r.PathValue("id")})
})

server := framework.NewServer(framework.ServerConfig{}, app)
```

`framework.Handler` returns an error. Errors implementing `framework.HTTPStatus` select a status; errors implementing `http.Handler`, including `problem.Details`, render themselves. A handler that neither writes a response nor returns an error receives `204 No Content`.

Use `framework.HTTP(standardHandler)` to adapt a standard handler for a route. Use `middleware.HTTP(standardMiddleware)` to adapt standard `func(http.Handler) http.Handler` middleware while retaining errors returned by Cosmos handlers.

## Middleware

Built-ins include recovery, structured logging, CORS, CSRF, secure headers, rate limiting, context injection, sessions, and correlation IDs. Put recovery and logging early in the chain. Use `framework.NewServer` rather than `http.ListenAndServe`; it applies secure timeout defaults.

## Sessions and OIDC

Create a cache facade and pass it to the cache session driver, then install the session middleware:

```go
cache := contract.NewCache(memory.NewMemory(memory.DefaultMemoryConfig))
driver := session.NewCacheDriver(cache, session.DefaultCacheDriverConfig)
app.Use(middleware.Session(driver, middleware.DefaultSessionConfig))
```

In handlers, `request.Session(r)` returns `(*contract.Session, bool)` and `request.MustSession(r)` panics if absent. Store values with `Put`, retrieve them with `Get[T]`, and call `Regenerate()` after authentication or a privilege change.

`framework/auth/oidc` supports Authorization Code with PKCE and bearer-token authentication. Construct a client with `oidc.New(ctx, oidc.Config{...})`; session middleware is required for `Login`, `Callback`, and `Logout`. Apply `client.RequireAuthentication()` and then `client.RequireScopes(...)` to protected routes.

## Configuration and Secrets

`framework/configuration` supplies map, filesystem, environment, and secret providers for the Koanf and Viper drivers. Later providers override earlier ones. Environment keys use `PREFIX__SECTION__KEY`; filesystem loading supports JSON, YAML, and YML in lexical order.

```go
driver, err := koanf.New(
	configuration.Map(map[string]any{"http.server.port": 8080}),
	configuration.Environment("COSMOS"),
)
config := contract.NewConfiguration(driver)
```

Use `configuration.JSONSecret`, `YAMLSecret`, or `RawSecret` with a `contract.SecretDriver` to add secret-backed values. Do not commit production credentials to configuration files.

## Crypto and Hashing

AES-GCM and ChaCha20-Poly1305 packages implement raw `contract.EncrypterDriver`; wrap one with `contract.NewEncrypter` for typed JSON encryption, or use `EncryptRaw` and `DecryptRaw` for bytes. Set `AdditionalData` before use to bind ciphertext to context, and call `Close` when finished.

Argon2id and bcrypt implement raw `contract.HasherDriver`; wrap one with `contract.NewHasher` for typed values, or use `HashRaw` and `CheckRaw` for passwords. Raw password byte slices are cleared by the drivers and must not be reused. Prefer Argon2id for new systems; bcrypt defaults to cost 12.

## Testing

Run from the workspace root:

```bash
go test ./framework/...
go test -race ./framework/...
```
