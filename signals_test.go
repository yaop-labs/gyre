package gyre_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/yaop-labs/gyre"
)

func TestRunUsesBoundedShutdownContext(t *testing.T) {
	closeDeadline := make(chan time.Time, 1)
	component := &gyre.FuncComponent{
		ComponentName:    "bounded",
		ComponentVersion: "test",
		CloseFunc: func(ctx context.Context) error {
			deadline, ok := ctx.Deadline()
			if !ok {
				return errors.New("shutdown context has no deadline")
			}
			closeDeadline <- deadline
			return nil
		},
	}
	var runtime gyre.Runtime
	if err := runtime.Add(component); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := gyre.RunWithShutdownTimeout(ctx, &runtime, time.Second); err != nil {
		t.Fatal(err)
	}
	select {
	case deadline := <-closeDeadline:
		remaining := time.Until(deadline)
		if remaining <= 0 || remaining > time.Second {
			t.Fatalf("unexpected shutdown deadline: %v", remaining)
		}
	default:
		t.Fatal("component was not closed")
	}
}

func TestRunRejectsInvalidArguments(t *testing.T) {
	if err := gyre.RunWithShutdownTimeout(context.Background(), nil, time.Second); err == nil {
		t.Fatal("expected nil runtime error")
	}
	if err := gyre.RunWithShutdownTimeout(context.Background(), &gyre.Runtime{}, 0); err == nil {
		t.Fatal("expected invalid timeout error")
	}
}
