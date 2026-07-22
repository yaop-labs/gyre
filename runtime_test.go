package gyre_test

import (
	"context"
	"encoding/json"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/yaop-labs/gyre"
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
	if _, err := r.Reload(context.Background(), "test", e); err == nil {
		t.Fatal("expected reload-after-close error")
	}
}

type lifecycleComponent struct {
	starts  atomic.Int32
	closes  atomic.Int32
	entered chan struct{}
	release chan struct{}
}

func (c *lifecycleComponent) Name() string    { return "lifecycle" }
func (c *lifecycleComponent) Version() string { return "test" }
func (c *lifecycleComponent) Start(context.Context) error {
	c.starts.Add(1)
	if c.entered != nil {
		close(c.entered)
		<-c.release
	}
	return nil
}
func (c *lifecycleComponent) Ready(context.Context) error { return nil }
func (c *lifecycleComponent) Status() gyre.Snapshot {
	return gyre.Snapshot{Name: c.Name(), Version: c.Version(), State: gyre.StateReady}
}
func (c *lifecycleComponent) Close(context.Context) error {
	c.closes.Add(1)
	return nil
}

func TestRuntimeSerializesLifecycle(t *testing.T) {
	c := &lifecycleComponent{entered: make(chan struct{}), release: make(chan struct{})}
	var runtime gyre.Runtime
	if err := runtime.Add(c); err != nil {
		t.Fatal(err)
	}

	startErrors := make(chan error, 2)
	go func() { startErrors <- runtime.Start(context.Background()) }()
	<-c.entered
	go func() { startErrors <- runtime.Start(context.Background()) }()
	close(c.release)
	for range 2 {
		if err := <-startErrors; err != nil {
			t.Fatal(err)
		}
	}
	if err := runtime.Close(context.Background()); err != nil {
		t.Fatal(err)
	}
	if got := c.starts.Load(); got != 1 {
		t.Fatalf("starts=%d, want 1", got)
	}
	if got := c.closes.Load(); got != 1 {
		t.Fatalf("closes=%d, want 1", got)
	}

	blocking := &lifecycleComponent{entered: make(chan struct{}), release: make(chan struct{})}
	var second gyre.Runtime
	if err := second.Add(blocking); err != nil {
		t.Fatal(err)
	}
	startDone := make(chan error, 1)
	go func() { startDone <- second.Start(context.Background()) }()
	<-blocking.entered
	closeDone := make(chan error, 1)
	go func() { closeDone <- second.Close(context.Background()) }()
	select {
	case err := <-closeDone:
		t.Fatalf("close raced past start: %v", err)
	case <-time.After(10 * time.Millisecond):
	}
	close(blocking.release)
	if err := <-startDone; err != nil {
		t.Fatal(err)
	}
	if err := <-closeDone; err != nil {
		t.Fatal(err)
	}
}

func TestRuntimeRejectsRegistrationAfterStart(t *testing.T) {
	var runtime gyre.Runtime
	if err := runtime.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := runtime.Add(&component{}); err == nil {
		t.Fatal("expected registration-after-start error")
	}
}

func TestRuntimeDependencyOrdering(t *testing.T) {
	var mu sync.Mutex
	var events []string
	componentFor := func(name string) *gyre.FuncComponent {
		return &gyre.FuncComponent{
			ComponentName:    name,
			ComponentVersion: "test",
			StartFunc: func(context.Context) error {
				mu.Lock()
				defer mu.Unlock()
				events = append(events, "start:"+name)
				return nil
			},
			CloseFunc: func(context.Context) error {
				mu.Lock()
				defer mu.Unlock()
				events = append(events, "close:"+name)
				return nil
			},
		}
	}
	var runtime gyre.Runtime
	if err := runtime.AddWithDependencies(componentFor("api"), "db"); err != nil {
		t.Fatal(err)
	}
	if err := runtime.Add(componentFor("db")); err != nil {
		t.Fatal(err)
	}
	if err := runtime.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := runtime.Close(context.Background()); err != nil {
		t.Fatal(err)
	}
	want := []string{"start:db", "start:api", "close:api", "close:db"}
	if len(events) != len(want) {
		t.Fatalf("events=%v", events)
	}
	for i := range want {
		if events[i] != want[i] {
			t.Fatalf("events=%v, want %v", events, want)
		}
	}
}
