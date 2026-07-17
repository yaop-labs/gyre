package gyre

import (
	"context"
	"errors"
)

// Registry coordinates a set of components with deterministic startup and
// reverse-order shutdown. It intentionally does not infer dependencies.
type Registry struct {
	components []Component
	started    bool
}

func (r *Registry) Add(component Component) error {
	if component == nil || component.Name() == "" {
		return errors.New("gyre: component and name are required")
	}
	for _, existing := range r.components {
		if existing.Name() == component.Name() {
			return errors.New("gyre: duplicate component " + component.Name())
		}
	}
	r.components = append(r.components, component)
	return nil
}

func (r *Registry) Start(ctx context.Context) error {
	if r.started {
		return nil
	}
	for i, component := range r.components {
		if err := component.Start(ctx); err != nil {
			for ; i >= 0; i-- {
				_ = r.components[i].Close(context.Background())
			}
			return E(CodeInternal, component.Name(), "start", false, err)
		}
	}
	r.started = true
	return nil
}

func (r *Registry) Ready(ctx context.Context) error {
	for _, component := range r.components {
		if err := component.Ready(ctx); err != nil {
			return E(CodeDependency, component.Name(), "ready", true, err)
		}
	}
	return nil
}

func (r *Registry) Close(ctx context.Context) error {
	var joined error
	for i := len(r.components) - 1; i >= 0; i-- {
		joined = errors.Join(joined, r.components[i].Close(ctx))
	}
	r.started = false
	return joined
}

func (r *Registry) Status() []Snapshot {
	out := make([]Snapshot, 0, len(r.components))
	for _, component := range r.components {
		out = append(out, component.Status())
	}
	return out
}
