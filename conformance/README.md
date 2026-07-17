# Gyre conformance

Products adopting Gyre should run `gyre.ConformanceCheck` against their adapter
and add product-specific tests for startup, readiness failure/recovery, reload
generation, and bounded shutdown. The root package intentionally contains the
small helper for v0.5; this directory reserves the reusable test-kit surface
for the next module split.
