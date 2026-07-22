# Gyre

Gyre is the shared operational contract for YAOP components. It standardizes
lifecycle, dependency ordering, readiness, status, reload generations, typed
errors, configuration control, and resource identity while leaving product
business logic in each product.

The latest tagged release is `v0.6.0`.

## Packages

- `github.com/yaop-labs/gyre` contains the stable contracts, runtime, config
  store, HTTP adapters, and wire-neutral health types.
- `github.com/yaop-labs/gyre/conformance` contains the reusable product adapter
  contract suite.

Gyre has no third-party Go dependencies. Reef remains responsible for TLS,
mTLS, bearer credentials, rotation, and network policy.

## Runtime

Register every component before starting the runtime. Dependencies determine
startup order; shutdown runs in reverse dependency order. Lifecycle operations
are serialized, repeated `Start` and `Close` are idempotent, and registration
after startup is rejected.

```go
var runtime gyre.Runtime
_ = runtime.Add(database)
_ = runtime.AddWithDependencies(api, database.Name())

ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
defer stop()
if err := gyre.RunWithShutdownTimeout(ctx, &runtime, 20*time.Second); err != nil {
    return err
}
```

Products can mount `gyre.RuntimeHTTPHandler(&runtime)` for aggregate
`/healthz`, `/readyz`, and `/status` endpoints, or
`gyre.HTTPHandler(component)` for one component.

## Configuration

`ConfigStore.ApplyWith` and `RollbackWith` serialize the generation check,
product reload, and in-memory commit. A rejected reload never replaces the
last-known-good envelope. The admin handler uses this transaction automatically
when it is constructed with a runtime.

Admin config and watch responses recursively redact conventional secret fields.
The admin handler does not authenticate requests; it must be mounted behind
Reef or another trusted host boundary.

## Conformance

Adopting products should run `conformance.Check` with a factory returning a
fresh adapter. The suite verifies lifecycle identity and idempotence and can
also exercise readiness failure/recovery, reload generations, rejected reloads,
and bounded shutdown.

```sh
go test ./... -race -count=1
go vet ./...
```
