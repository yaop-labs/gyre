package gyre

import (
	"context"
	"errors"
	"sync"
)

// Runtime owns the lifecycle of registered components and coordinates reloads.
// Product code still owns dependency meaning and configuration decoding.
type Runtime struct {
	lifecycleMu sync.Mutex
	mu          sync.Mutex
	components  []Component
	reloadable  map[string]Reloadable
	reloadMu    sync.Mutex
	started     bool
	closed      bool
	deps        map[string][]string
}

func (r *Runtime) AddWithDependencies(component Component, dependencies ...string) error {
	r.lifecycleMu.Lock()
	defer r.lifecycleMu.Unlock()
	if err := r.add(component); err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.deps == nil {
		r.deps = map[string][]string{}
	}
	r.deps[component.Name()] = append([]string(nil), dependencies...)
	return nil
}

func (r *Runtime) Add(component Component) error {
	r.lifecycleMu.Lock()
	defer r.lifecycleMu.Unlock()
	return r.add(component)
}

func (r *Runtime) add(component Component) error {
	if component == nil || component.Name() == "" {
		return errors.New("gyre: component and name are required")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.started {
		return errors.New("gyre: cannot add component after runtime start")
	}
	if r.closed {
		return E(CodeShuttingDown, "runtime", "add", false, nil)
	}
	for _, existing := range r.components {
		if existing.Name() == component.Name() {
			return errors.New("gyre: duplicate component " + component.Name())
		}
	}
	r.components = append(r.components, component)
	if reloadable, ok := component.(Reloadable); ok {
		if r.reloadable == nil {
			r.reloadable = map[string]Reloadable{}
		}
		r.reloadable[component.Name()] = reloadable
	}
	return nil
}

func (r *Runtime) Start(ctx context.Context) error {
	r.lifecycleMu.Lock()
	defer r.lifecycleMu.Unlock()
	r.mu.Lock()
	if r.closed {
		r.mu.Unlock()
		return E(CodeShuttingDown, "runtime", "start", false, nil)
	}
	if r.started {
		r.mu.Unlock()
		return nil
	}
	components := append([]Component(nil), r.components...)
	r.mu.Unlock()
	ordered, err := r.order(components)
	if err != nil {
		return err
	}
	for i, component := range ordered {
		if err := component.Start(ctx); err != nil {
			cleanupCtx, cancel := context.WithTimeout(context.Background(), DefaultShutdownTimeout)
			defer cancel()
			for ; i >= 0; i-- {
				_ = ordered[i].Close(cleanupCtx)
			}
			return E(CodeInternal, component.Name(), "start", false, err)
		}
	}
	r.mu.Lock()
	r.started = true
	r.mu.Unlock()
	return nil
}

func (r *Runtime) order(components []Component) ([]Component, error) {
	r.mu.Lock()
	deps := map[string][]string{}
	for k, v := range r.deps {
		deps[k] = append([]string(nil), v...)
	}
	r.mu.Unlock()
	byName := map[string]Component{}
	for _, c := range components {
		byName[c.Name()] = c
	}
	state := map[string]uint8{}
	out := make([]Component, 0, len(components))
	var visit func(string) error
	visit = func(name string) error {
		if state[name] == 1 {
			return errors.New("gyre: dependency cycle at " + name)
		}
		if state[name] == 2 {
			return nil
		}
		state[name] = 1
		for _, dep := range deps[name] {
			if _, ok := byName[dep]; !ok {
				return errors.New("gyre: missing dependency " + dep)
			}
			if err := visit(dep); err != nil {
				return err
			}
		}
		state[name] = 2
		out = append(out, byName[name])
		return nil
	}
	for _, c := range components {
		if err := visit(c.Name()); err != nil {
			return nil, err
		}
	}
	return out, nil
}

func (r *Runtime) Ready(ctx context.Context) error {
	r.mu.Lock()
	components := append([]Component(nil), r.components...)
	r.mu.Unlock()
	for _, component := range components {
		if err := component.Ready(ctx); err != nil {
			return E(CodeDependency, component.Name(), "ready", true, err)
		}
	}
	return nil
}

func (r *Runtime) Reload(ctx context.Context, name string, envelope Envelope) (ReloadResult, error) {
	if err := envelope.Validate(); err != nil {
		return ReloadResult{}, E(CodeConfigInvalid, name, "reload", false, err)
	}
	r.lifecycleMu.Lock()
	defer r.lifecycleMu.Unlock()
	r.reloadMu.Lock()
	defer r.reloadMu.Unlock()
	r.mu.Lock()
	if r.closed {
		r.mu.Unlock()
		return ReloadResult{}, E(CodeShuttingDown, name, "reload", false, nil)
	}
	reloadable := r.reloadable[name]
	r.mu.Unlock()
	if reloadable == nil {
		return ReloadResult{}, E(CodeUnavailable, name, "reload", false, errors.New("component is not reloadable"))
	}
	return reloadable.Reload(ctx, envelope)
}

func (r *Runtime) Status() []Snapshot {
	r.mu.Lock()
	components := append([]Component(nil), r.components...)
	r.mu.Unlock()
	out := make([]Snapshot, 0, len(components))
	for _, component := range components {
		out = append(out, component.Status())
	}
	return out
}

func (r *Runtime) Close(ctx context.Context) error {
	r.lifecycleMu.Lock()
	defer r.lifecycleMu.Unlock()
	r.mu.Lock()
	if r.closed {
		r.mu.Unlock()
		return nil
	}
	r.closed = true
	components := append([]Component(nil), r.components...)
	r.mu.Unlock()
	if ordered, err := r.order(components); err == nil {
		components = ordered
	}
	var joined error
	for i := len(components) - 1; i >= 0; i-- {
		joined = errors.Join(joined, components[i].Close(ctx))
	}
	return joined
}
