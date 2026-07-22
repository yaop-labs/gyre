// Package conformance provides the reusable Gyre adapter contract suite.
package conformance

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/yaop-labs/gyre"
)

const defaultShutdownTimeout = time.Second

// Fixture describes one fresh product adapter and optional controls used to
// verify readiness transitions and reload generation behavior.
type Fixture struct {
	Component       gyre.Component
	SetReady        func(bool)
	ValidReload     *gyre.Envelope
	InvalidReload   *gyre.Envelope
	Generation      func() uint64
	ShutdownTimeout time.Duration
}

type Factory func() Fixture

// Check exercises close-before-start, identity, startup, readiness recovery,
// reload last-known-good behavior, bounded shutdown, and repeated close. A
// factory is used because close is terminal for many product adapters.
func Check(ctx context.Context, factory Factory) error {
	if factory == nil {
		return errors.New("gyre conformance: factory is required")
	}
	closed := factory()
	if err := gyre.ConformanceCheck(ctx, closed.Component); err != nil {
		return err
	}

	fixture := factory()
	component := fixture.Component
	if component == nil {
		return errors.New("gyre conformance: factory returned nil component")
	}
	if err := component.Start(ctx); err != nil {
		return fmt.Errorf("gyre conformance: start: %w", err)
	}
	snapshot := component.Status()
	if snapshot.Name != component.Name() || snapshot.Version != component.Version() {
		return errors.New("gyre conformance: status identity does not match component identity")
	}

	if fixture.SetReady != nil {
		fixture.SetReady(false)
		if err := component.Ready(ctx); err == nil {
			return errors.New("gyre conformance: readiness failure was not reported")
		}
		fixture.SetReady(true)
		if err := component.Ready(ctx); err != nil {
			return fmt.Errorf("gyre conformance: readiness did not recover: %w", err)
		}
	}

	if fixture.ValidReload != nil {
		reloadable, ok := component.(gyre.Reloadable)
		if !ok {
			return errors.New("gyre conformance: reload fixture is not reloadable")
		}
		result, err := reloadable.Reload(ctx, *fixture.ValidReload)
		if err != nil {
			return fmt.Errorf("gyre conformance: valid reload: %w", err)
		}
		if result.Generation != fixture.ValidReload.Generation {
			return errors.New("gyre conformance: reload returned the wrong generation")
		}
		accepted := result.Generation
		if fixture.Generation != nil {
			accepted = fixture.Generation()
			if accepted != fixture.ValidReload.Generation {
				return errors.New("gyre conformance: accepted generation was not published")
			}
		}
		if fixture.InvalidReload != nil {
			if _, err := reloadable.Reload(ctx, *fixture.InvalidReload); err == nil {
				return errors.New("gyre conformance: invalid reload was accepted")
			}
			if fixture.Generation != nil && fixture.Generation() != accepted {
				return errors.New("gyre conformance: invalid reload replaced last-known-good generation")
			}
		}
	}

	timeout := fixture.ShutdownTimeout
	if timeout <= 0 {
		timeout = defaultShutdownTimeout
	}
	closeCtx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	done := make(chan error, 1)
	go func() { done <- component.Close(closeCtx) }()
	select {
	case err := <-done:
		if err != nil {
			return fmt.Errorf("gyre conformance: close: %w", err)
		}
	case <-closeCtx.Done():
		return errors.New("gyre conformance: close exceeded its context bound")
	}
	if err := component.Close(context.Background()); err != nil {
		return fmt.Errorf("gyre conformance: repeated close: %w", err)
	}
	return nil
}
