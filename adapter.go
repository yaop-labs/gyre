package gyre

import (
	"context"
	"time"
)

// FuncComponent adapts an existing product lifecycle to Gyre. It is useful
// during migration and keeps compatibility code at the boundary.
type FuncComponent struct {
	ComponentName    string
	ComponentVersion string
	StartFunc        func(context.Context) error
	ReadyFunc        func(context.Context) error
	CloseFunc        func(context.Context) error
	StatusFunc       func() Snapshot
}

func (f *FuncComponent) Name() string    { return f.ComponentName }
func (f *FuncComponent) Version() string { return f.ComponentVersion }
func (f *FuncComponent) Start(ctx context.Context) error {
	if f.StartFunc == nil {
		return nil
	}
	return f.StartFunc(ctx)
}
func (f *FuncComponent) Ready(ctx context.Context) error {
	if f.ReadyFunc == nil {
		return nil
	}
	return f.ReadyFunc(ctx)
}
func (f *FuncComponent) Close(ctx context.Context) error {
	if f.CloseFunc == nil {
		return nil
	}
	return f.CloseFunc(ctx)
}
func (f *FuncComponent) Status() Snapshot {
	if f.StatusFunc != nil {
		return f.StatusFunc()
	}
	return Snapshot{Name: f.Name(), Version: f.Version(), State: StateStarting, Since: time.Now()}
}
