# Gyre decisions

## D-001: contract before runtime

The first release contains stable types, interfaces, adapters, and conformance
tests. It does not require every product to adopt a central orchestrator.

## D-002: distinct liveness and readiness

`/healthz` is process liveness and must not fail merely because a dependency is
temporarily unavailable. `/readyz` reports whether the component can perform
its advertised work. A startup state is represented in status and readiness,
not by making liveness expensive.

## D-003: one shutdown vocabulary

The public lifecycle method is `Close(context.Context) error`. Adapters may
map existing `Shutdown` or database `Close` methods to it. Close is idempotent
and context-bounded.

## D-004: Reef remains the transport security layer

Gyre references Reef for TLS, mTLS, bearer credentials, rotation, and policy;
it does not duplicate those primitives.
