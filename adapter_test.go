package gyre_test

import (
	"context"
	"github.com/yaop-labs/gyre"
	"testing"
)

func TestFuncComponentAdapter(t *testing.T) {
	started, ready, closed := false, false, false
	c := &gyre.FuncComponent{
		ComponentName:    "legacy",
		ComponentVersion: "1",
		StartFunc:        func(context.Context) error { started = true; return nil },
		ReadyFunc:        func(context.Context) error { ready = true; return nil },
		CloseFunc:        func(context.Context) error { closed = true; return nil },
		StatusFunc: func() gyre.Snapshot {
			return gyre.Snapshot{Name: "legacy", Version: "1", State: gyre.StateReady}
		},
	}
	if err := c.Start(context.Background()); err != nil || !started {
		t.Fatal("start adapter failed")
	}
	if err := c.Close(context.Background()); err != nil || !closed {
		t.Fatal("close adapter failed")
	}
	if err := c.Ready(context.Background()); err != nil || !ready || c.Status().State != gyre.StateReady {
		t.Fatal("ready/status adapter failed")
	}
	if err := gyre.ConformanceCheck(context.Background(), c); err != nil {
		t.Fatal(err)
	}
}
