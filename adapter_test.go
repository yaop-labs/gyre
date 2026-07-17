package gyre_test

import (
	"context"
	"github.com/yaop-labs/gyre"
	"testing"
)

func TestFuncComponentAdapter(t *testing.T) {
	started, closed := false, false
	c := &gyre.FuncComponent{ComponentName: "legacy", ComponentVersion: "1", StartFunc: func(context.Context) error { started = true; return nil }, CloseFunc: func(context.Context) error { closed = true; return nil }}
	if err := c.Start(context.Background()); err != nil || !started {
		t.Fatal("start adapter failed")
	}
	if err := c.Close(context.Background()); err != nil || !closed {
		t.Fatal("close adapter failed")
	}
	if err := gyre.ConformanceCheck(context.Background(), c); err != nil {
		t.Fatal(err)
	}
}
