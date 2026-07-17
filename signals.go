package gyre

import "context"

// Run starts runtime and blocks until ctx is canceled, then performs a
// context-bounded close. Signal delivery belongs to the executable and should
// be converted to context cancellation with os/signal.NotifyContext.
func Run(ctx context.Context, runtime *Runtime) error {
	if runtime == nil {
		return E(CodeInternal, "runtime", "run", false, nil)
	}
	if err := runtime.Start(ctx); err != nil {
		return err
	}
	<-ctx.Done()
	return runtime.Close(context.Background())
}
