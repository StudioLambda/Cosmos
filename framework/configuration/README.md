# Configuration

Cosmos provides reusable configuration providers and interchangeable adapters
for the `contract.ConfigurationDriver` interface:

- `framework/configuration/koanf`
- `framework/configuration/viper`

Providers are resolved in order: later providers override earlier values. The
filesystem provider loads `.json`, `.yml`, and `.yaml` files from `config/` in lexical
order. Values in later files override earlier files.

## Environment

The environment provider overlays variables using double underscores for
nesting, preserving snake_case key names:

```text
COSMOS__HTTP__SERVER__PORT=9090
COSMOS__HTTP__SERVER__READ_TIMEOUT=45s
COSMOS__CACHE__REDIS__PASSWORD=secret
```

These map to `http.server.port`, `http.server.read_timeout`, and
`cache.redis.password`. Environment values have highest precedence, including
an explicitly empty value.

```go
driver, err := koanf.New(
	configuration.Map(map[string]any{"http.server.port": 8080}),
	configuration.Filesystem(configurationFS),
	configuration.Environment("COSMOS"),
)
```

Use a deployment secret manager or environment injection for credentials; do
not commit production secrets to YAML files.

## Extension

Configuration can be extended during application bootstrap with the same
providers used during construction. Extension must complete before concurrent
configuration reads begin.

```go
err := config.Extend(
    configuration.JSONSecret(ctx, secrets, "cosmos/api/production"),
    configuration.Environment("COSMOS"),
)
```

Use [JSONSecret] or [YAMLSecret] when a secret contains a complete
configuration object. Use [RawSecret] to assign a text secret to one explicit
configuration key.
