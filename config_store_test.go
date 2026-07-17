package gyre_test

import (
	"context"
	"encoding/json"
	"github.com/yaop-labs/gyre"
	"testing"
)

func TestConfigStoreApplyWatchRollback(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var s gyre.ConfigStore
	watch := s.Watch(ctx)
	makeEnv := func(g uint64) gyre.Envelope {
		return gyre.Envelope{APIVersion: "gyre/v1", Kind: "Wisp", Generation: g, Spec: json.RawMessage(`{}`)}
	}
	if err := s.Apply(ctx, makeEnv(1), 0); err != nil {
		t.Fatal(err)
	}
	if err := s.Apply(ctx, makeEnv(2), 1); err != nil {
		t.Fatal(err)
	}
	select {
	case got := <-watch:
		if got.Generation != 1 {
			t.Fatalf("generation=%d", got.Generation)
		}
	case <-ctx.Done():
		t.Fatal("watch timeout")
	}
	if err := s.Rollback(ctx, 1); err != nil {
		t.Fatal(err)
	}
	if got, ok := s.Current(); !ok || got.Generation != 3 {
		t.Fatalf("current=%+v ok=%v", got, ok)
	}
}
