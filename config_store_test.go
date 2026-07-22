package gyre_test

import (
	"context"
	"encoding/json"
	"errors"
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

func TestConfigStoreApplyWithIsAtomic(t *testing.T) {
	ctx := context.Background()
	var store gyre.ConfigStore
	makeEnv := func(g uint64) gyre.Envelope {
		return gyre.Envelope{APIVersion: "gyre/v1", Kind: "Wisp", Generation: g, Spec: json.RawMessage(`{"token":"secret"}`)}
	}
	first := makeEnv(1)
	if err := store.Apply(ctx, first, 0); err != nil {
		t.Fatal(err)
	}
	first.Spec[2] = 'X'
	if current, _ := store.Current(); string(current.Spec) != `{"token":"secret"}` {
		t.Fatalf("store retained caller-owned spec: %s", current.Spec)
	}
	if err := store.ApplyWith(ctx, makeEnv(2), 1, func(context.Context, gyre.Envelope) error {
		current, ok := store.Current()
		if !ok || current.Generation != 1 {
			t.Fatalf("callback current=%+v ok=%v", current, ok)
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}

	wantErr := errors.New("product rejected config")
	err := store.ApplyWith(ctx, makeEnv(3), 2, func(context.Context, gyre.Envelope) error {
		return wantErr
	})
	if !errors.Is(err, wantErr) {
		t.Fatalf("apply error=%v", err)
	}
	if current, _ := store.Current(); current.Generation != 2 {
		t.Fatalf("failed apply replaced current config: %+v", current)
	}
	audit := store.Audit()
	if len(audit) != 3 || audit[2].Success || audit[2].Error != "gyre: operation failed" {
		t.Fatalf("audit=%+v", audit)
	}
}

func TestConfigStoreRollbackWithIsAtomic(t *testing.T) {
	ctx := context.Background()
	var store gyre.ConfigStore
	makeEnv := func(g uint64) gyre.Envelope {
		return gyre.Envelope{APIVersion: "gyre/v1", Kind: "Wisp", Generation: g, Spec: json.RawMessage(`{}`)}
	}
	if err := store.Apply(ctx, makeEnv(1), 0); err != nil {
		t.Fatal(err)
	}
	if err := store.Apply(ctx, makeEnv(2), 1); err != nil {
		t.Fatal(err)
	}
	if err := store.RollbackWith(ctx, 1, func(context.Context, gyre.Envelope) error {
		return errors.New("reload failed")
	}); err == nil {
		t.Fatal("expected rollback failure")
	}
	if current, _ := store.Current(); current.Generation != 2 {
		t.Fatalf("failed rollback replaced current config: %+v", current)
	}
}
