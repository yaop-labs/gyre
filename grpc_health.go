package gyre

import "context"

// HealthStatus is intentionally wire-neutral and maps directly to the
// grpc.health.v1 ServingStatus values in adapters.
type HealthStatus int

const (
	HealthUnknown HealthStatus = iota
	HealthServing
	HealthNotServing
)

// GRPCHealthAdapter provides the semantics required by grpc health services
// without forcing gyre to depend on a particular gRPC implementation.
// A transport bridge can translate HealthStatus to its native enum.
type GRPCHealthAdapter struct{ Runtime *Runtime }

func (a GRPCHealthAdapter) Check(ctx context.Context, service string) HealthStatus {
	if a.Runtime == nil {
		return HealthUnknown
	}
	if err := a.Runtime.Ready(ctx); err != nil {
		return HealthNotServing
	}
	if service == "" || service == "gyre" {
		return HealthServing
	}
	for _, s := range a.Runtime.Status() {
		if s.Name == service {
			if s.State == StateReady {
				return HealthServing
			}
			return HealthNotServing
		}
	}
	return HealthUnknown
}
