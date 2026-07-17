# Changelog

## v0.5.0

- runtime component registry with dependency ordering;
- lifecycle, readiness, status, and reverse shutdown orchestration;
- serialized versioned config reload;
- config store with apply, dry-run, watch, rollback, and audit history;
- admin HTTP endpoints for config, status, audit, and watch;
- HTTP and wire-neutral gRPC health adapters;
- metrics observer interface;
- typed operational errors and secret-safe config redaction;
- resource identity merge rules;
- Wisp reference adapter and conformance tests.

## Compatibility

The `gyre/v1` envelope and JSON status fields are additive within the 0.x
series. Products must reject unknown or invalid generations without replacing
last-known-good state. Breaking API changes require a new module major version.
