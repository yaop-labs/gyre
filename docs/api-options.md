# Gyre API options research

## Reload

### Typed product config

`Reload(context.Context, Config)` keeps the core generic but pushes decoding and
validation into every component. It is flexible, but mistakes become runtime
type assertions and the contract is hard to expose over HTTP.

### Versioned envelope

`Reload(context.Context, Envelope)` carries `api_version`, `kind`,
`generation`, and redacted metadata. Products decode the payload into their
configuration. This is the best fit for file, remote, and administrative
reloads, and allows rejected generations to remain observable.

### Product-owned callback

Gyre only exposes a `ReloadFunc`. This is simple but gives up standard error,
generation, and status semantics.

**Recommendation:** use a versioned envelope in the adapter layer and a typed
`Apply` callback internally. Return a result containing accepted generation,
changed fields, and restart-required fields.

## Status and metrics

Counters should not be embedded in snapshots: snapshots are point-in-time
state and counters have metric cardinality/lifetime semantics. Status may carry
small counters such as generation and reload failures, while detailed rates
remain in metrics.

**Recommendation:** `Snapshot` contains state, conditions, generation, and
timestamps. Gyre's telemetry adapter emits standard counters/gauges.

## Dependencies

The product owns its dependency graph because only the product knows whether a
dependency is required for its advertised work. Gyre provides a registry and
aggregation helper, but does not infer dependency criticality.

**Recommendation:** register named checks with `RequiredFor: Liveness|Readiness`
and a bounded `Check(ctx)` function. Never run unbounded dependency checks from
`/healthz`.

## Resource merge precedence

Use explicit layers, from weakest to strongest:

1. detected host/runtime defaults;
2. environment/orchestrator metadata;
3. component configuration;
4. per-signal or per-request identity.

Reject conflicting required identity (`service.name`) rather than silently
choosing between two explicit values. Keep observed-resource attributes
separate from agent identity.

## HTTP adapter boundary

The HTTP adapter should be a subpackage, not part of core. Core must work for
CLI-only libraries, gRPC-only services, and embedded databases. The adapter
mounts `/healthz`, `/readyz`, `/status`, and optionally `/metrics`; products may
compose it into an existing mux.

## Lifecycle precedent

OpenTelemetry Collector components use a start/shutdown contract and have
learned that shutdown must wait for run cleanup to complete. Gyre should make
that guarantee explicit, support Close before Start, and make Close idempotent.
