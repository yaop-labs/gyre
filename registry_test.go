package gyre_test

import (
	"context"
	"github.com/yaop-labs/gyre"
	"testing"
)

func TestRegistryLifecycle(t *testing.T) {
	first, second := &component{ready: true}, &component{ready: true}
	// The test component names are identical by design; duplicate registration
	// is rejected before startup.
	var r gyre.Registry
	if err := r.Add(first); err != nil {
		t.Fatal(err)
	}
	if err := r.Add(second); err == nil {
		t.Fatal("expected duplicate component error")
	}
	if err := r.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := r.Ready(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := r.Close(context.Background()); err != nil {
		t.Fatal(err)
	}
}
