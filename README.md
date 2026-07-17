# Gyre

Gyre is the shared operational contract for YAOP components. It standardizes
lifecycle, readiness, status, reload generations, typed errors, and resource
identity while leaving product business logic in each product.

## v0.1

The first release is intentionally a small Go module:

- core `Component`, `Reloadable`, `Snapshot`, `Condition`, and `Envelope` types;
- typed operational errors with retryability;
- deterministic resource merging with explicit identity conflict rejection;
- HTTP adapter exposing `/healthz`, `/readyz`, and `/status`;
- named readiness checks and contract tests.

The HTTP adapter is composable: products mount `gyre.HTTPHandler(component)` in
their own mux and may add product-specific endpoints. `/healthz` is deliberately
cheap and does not execute dependency checks; `/readyz` executes bounded
readiness checks supplied by the component.

See `docs/` for the research, decisions, API draft, and migration matrix.
