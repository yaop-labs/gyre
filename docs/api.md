# Gyre v0.1 API draft

This is a design contract, not implementation code. Public Go names are
deliberately small so adapters can be added without forcing a framework.

```go
type State string

const (
    StateStarting State = "starting"
    StateReady State = "ready"
    StateDegraded State = "degraded"
    StateStopping State = "stopping"
    StateStopped State = "stopped"
    StateFailed State = "failed"
)

type Component interface {
    Name() string
    Version() string
    Start(context.Context) error
    Ready(context.Context) error
    Status() Snapshot
    Close(context.Context) error
}

type Reloadable interface {
    Reload(context.Context, Envelope) (ReloadResult, error)
}
```

`Start` is not required to block until all dependencies are ready. `Ready` is
bounded and may return a typed dependency error. `Close` is safe before Start,
safe to call more than once, and must honor its context.

## Status JSON

```json
{
  "name": "wisp",
  "version": "0.3.0",
  "state": "ready",
  "generation": 7,
  "since": "2026-07-10T12:00:00Z",
  "conditions": [{
    "type": "exporter",
    "status": "true",
    "reason": "connected",
    "message": "",
    "last_transition": "2026-07-10T12:00:00Z"
  }]
}
```

Status output must be secret-free and stable enough for operators, but not a
long-term storage schema. Additive fields are allowed in v0.x; removing or
changing field meaning requires a new API version.

## Configuration envelope

```yaml
api_version: gyre/v1
kind: Wisp
generation: 7
spec: {}
```

`generation` identifies an accepted configuration, not an attempted one.
Rejected generations remain visible through a reload condition and error
metric, but never replace the active configuration.

## Conformance requirements

An adopting component must test:

- Start, Ready, Close ordering and cancellation;
- Close before Start and repeated Close;
- failed Start leaves no leaked listener or goroutine;
- readiness transitions when a required dependency fails and recovers;
- status generation changes only after successful reload;
- invalid reload preserves the last-known-good configuration;
- health endpoints never expose secrets;
- typed errors preserve retryability and operation context.
