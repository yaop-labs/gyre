package gyre

import (
	"context"
	"testing"
	"time"
)

type healthComponent struct {
	name  string
	state State
}

func (c healthComponent) Name() string                { return c.name }
func (c healthComponent) Version() string             { return "test" }
func (c healthComponent) Start(context.Context) error { return nil }
func (c healthComponent) Ready(context.Context) error {
	if c.state != StateReady {
		return context.Canceled
	}
	return nil
}
func (c healthComponent) Status() Snapshot {
	return Snapshot{Name: c.name, State: c.state, Since: time.Now()}
}
func (c healthComponent) Close(context.Context) error { return nil }

func TestGRPCHealthAdapter(t *testing.T) {
	r := &Runtime{}
	r.Add(healthComponent{name: "api", state: StateReady})
	a := GRPCHealthAdapter{Runtime: r}
	if got := a.Check(context.Background(), "api"); got != HealthServing {
		t.Fatalf("got %v", got)
	}
	if got := a.Check(context.Background(), "missing"); got != HealthUnknown {
		t.Fatalf("got %v", got)
	}
}
