package conformance_test

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/yaop-labs/gyre"
	"github.com/yaop-labs/gyre/conformance"
)

type adapter struct {
	mu         sync.Mutex
	ready      bool
	generation uint64
	closed     bool
}

func (*adapter) Name() string                { return "reference" }
func (*adapter) Version() string             { return "test" }
func (*adapter) Start(context.Context) error { return nil }
func (a *adapter) Ready(context.Context) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if !a.ready {
		return errors.New("dependency unavailable")
	}
	return nil
}
func (a *adapter) Status() gyre.Snapshot {
	return gyre.Snapshot{Name: a.Name(), Version: a.Version(), State: gyre.StateReady, Since: time.Now()}
}
func (a *adapter) Close(context.Context) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.closed = true
	return nil
}
func (a *adapter) Reload(_ context.Context, envelope gyre.Envelope) (gyre.ReloadResult, error) {
	if err := envelope.Validate(); err != nil {
		return gyre.ReloadResult{}, err
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	a.generation = envelope.Generation
	return gyre.ReloadResult{Generation: envelope.Generation}, nil
}

func TestCheck(t *testing.T) {
	valid := gyre.Envelope{APIVersion: "gyre/v1", Kind: "Reference", Generation: 2, Spec: json.RawMessage(`{}`)}
	invalid := gyre.Envelope{APIVersion: "gyre/v1", Kind: "Reference", Generation: 3}
	factory := func() conformance.Fixture {
		component := &adapter{}
		return conformance.Fixture{
			Component: component,
			SetReady: func(ready bool) {
				component.mu.Lock()
				defer component.mu.Unlock()
				component.ready = ready
			},
			ValidReload:   &valid,
			InvalidReload: &invalid,
			Generation: func() uint64 {
				component.mu.Lock()
				defer component.mu.Unlock()
				return component.generation
			},
		}
	}
	if err := conformance.Check(context.Background(), factory); err != nil {
		t.Fatal(err)
	}
}
