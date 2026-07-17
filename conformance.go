package gyre

import (
	"context"
	"errors"
)

// ConformanceCheck performs cheap, deterministic checks an adapter can run in
// its own test suite. It does not start network listeners or inspect product
// internals.
func ConformanceCheck(ctx context.Context, component Component) error {
	if component == nil || component.Name() == "" || component.Version() == "" {
		return errors.New("gyre: invalid component identity")
	}
	if err := component.Close(ctx); err != nil {
		return errors.New("gyre: close-before-start: " + err.Error())
	}
	return nil
}
