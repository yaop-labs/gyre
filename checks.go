package gyre

import (
	"context"
	"errors"
)

type CheckScope string

const (
	ScopeLiveness  CheckScope = "liveness"
	ScopeReadiness CheckScope = "readiness"
)

type Check struct {
	Name  string
	Scope CheckScope
	Check func(context.Context) error
}

type Checker struct{ checks []Check }

func (c *Checker) Add(check Check) error {
	if check.Name == "" || check.Check == nil {
		return errors.New("gyre: check name and function are required")
	}
	if check.Scope == "" {
		check.Scope = ScopeReadiness
	}
	c.checks = append(c.checks, check)
	return nil
}

func (c *Checker) Ready(ctx context.Context) error {
	for _, check := range c.checks {
		if check.Scope != ScopeReadiness {
			continue
		}
		if err := check.Check(ctx); err != nil {
			return E(CodeDependency, check.Name, "ready", true, err)
		}
	}
	return nil
}
