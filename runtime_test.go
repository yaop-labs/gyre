package gyre_test

import (
	"context"
	"encoding/json"
	"github.com/yaop-labs/gyre"
	"testing"
)

type reloadComponent struct {
	component
	generation uint64
}

func (c *reloadComponent) Reload(_ context.Context, e gyre.Envelope) (gyre.ReloadResult, error) {
	c.generation = e.Generation
	return gyre.ReloadResult{Generation: e.Generation}, nil
}

func TestRuntimeReloadAndClose(t *testing.T) {
	c := &reloadComponent{component: component{ready: true}}
	var r gyre.Runtime
	if err := r.Add(c); err != nil {
		t.Fatal(err)
	}
	if err := r.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	e := gyre.Envelope{APIVersion: "gyre/v1", Kind: "Test", Generation: 2, Spec: json.RawMessage(`{}`)}
	result, err := r.Reload(context.Background(), "test", e)
	if err != nil || result.Generation != 2 {
		t.Fatalf("result=%+v err=%v", result, err)
	}
	if err := r.Close(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := r.Start(context.Background()); err == nil {
		t.Fatal("expected start-after-close error")
	}
}
