# Gyre design research

## Scope

Gyre is the shared operational contract for YAOP components. It must unify
lifecycle, readiness, status, configuration metadata, resource identity,
errors, and telemetry without owning product business logic.

## Local findings

Wisp exposes `Start`, `Reload`, and `Shutdown`; Coral has `Start` and
`Shutdown`; Amber exposes `Close`, `Status`, and `IsReady`; Fathom has a
library-oriented API and no common runtime lifecycle; Manta is currently a UI
design surface. Health endpoints are `/healthz`, `/readyz`, `/health`, and
`/readiness` depending on the product.

The first migration target should be Wisp and Coral. Amber needs an adapter for
its database lifecycle, while Fathom and Manta should follow after the core
contract has stabilized.

## External constraints

Kubernetes separates startup, liveness, and readiness. Liveness must answer
whether the process can make progress; readiness may include required
dependencies and can temporarily remove a workload from traffic. Gyre adopts
the same distinction and keeps readiness checks cheap and bounded.

OpenTelemetry Resource semantics make `service.name` the logical service
identity. Gyre must preserve observed-resource identity separately from the
agent/distribution identity and use deterministic attribute merge precedence.

gRPC has a wire-compatible health model; Gyre should provide an adapter rather
than inventing a competing RPC protocol.

Graceful shutdown must be context-bounded and idempotent. Components must stop
accepting new work, drain admitted work, and then release resources.

## Decisions pending

- whether `Reload` accepts a typed envelope or a product-owned value;
- whether status snapshots include counters or expose them only through
  metrics;
- dependency graph ownership (Gyre runtime versus product application);
- exact resource attribute allowlist and merge precedence;
- whether the HTTP adapter is part of the core module or a subpackage.

## Non-goals

Gyre will not contain storage, pipelines, business workflows, UI models, or a
distributed control plane.
