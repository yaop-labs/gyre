package gyre

import (
	"context"
	"time"
)

const DefaultShutdownTimeout = 30 * time.Second

// Run starts runtime and blocks until ctx is canceled, then performs a
// context-bounded close. Signal delivery belongs to the executable and should
// be converted to context cancellation with os/signal.NotifyContext.
func Run(ctx context.Context, runtime *Runtime) error {
	return RunWithShutdownTimeout(ctx, runtime, DefaultShutdownTimeout)
}

// RunWithShutdownTimeout is Run with an explicit shutdown bound. Shutdown gets
// a fresh context because ctx has already been canceled when cleanup begins.
func RunWithShutdownTimeout(ctx context.Context, runtime *Runtime, timeout time.Duration) error {
	if runtime == nil {
		return E(CodeInternal, "runtime", "run", false, nil)
	}
	if timeout <= 0 {
		return E(CodeConfigInvalid, "runtime", "run", false, nil)
	}
	if err := runtime.Start(ctx); err != nil {
		return err
	}
	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return runtime.Close(shutdownCtx)
}
