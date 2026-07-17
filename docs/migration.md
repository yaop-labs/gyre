# Adoption matrix

| Product | Adapter work | First milestone |
|---|---|---|
| Wisp | Map `Start/Shutdown/Reload`, register spool/exporter checks, mount standard endpoints | Gyre v0.1 reference |
| Coral | Wrap app/pipeline start and shutdown, normalize `/healthz` and `/readyz`, map resource attrs | v0.1 consumer |
| Amber | Adapter around DB `Status/IsReady/Close`; preserve rich DB status as conditions | v0.2 consumer |
| Fathom | Add explicit runtime owner and bounded readiness separate from heavy analysis readiness | v0.2 consumer |
| Manta | Use status/resource schemas for UI; no server lifecycle until backend exists | v0.3 consumer |

Migration is additive. Existing product APIs remain available during the
adapter phase. A product only claims Gyre conformance after its adapter and
contract tests pass.

## Compatibility

- Gyre API versions are explicit (`gyre/v1`).
- JSON status fields are additive within a major version.
- Config generations are monotonic per component.
- A rejected reload never partially applies.
- Reef owns TLS/auth semantics; Gyre only carries references and redacted
  status.
