package gyre

import (
	"context"
	"time"
)

type State string

const (
	StateStarting State = "starting"
	StateReady    State = "ready"
	StateDegraded State = "degraded"
	StateStopping State = "stopping"
	StateStopped  State = "stopped"
	StateFailed   State = "failed"
)

type Condition struct {
	Type           string    `json:"type"`
	Status         bool      `json:"status"`
	Reason         string    `json:"reason,omitempty"`
	Message        string    `json:"message,omitempty"`
	LastTransition time.Time `json:"last_transition"`
}

type Snapshot struct {
	Name       string      `json:"name"`
	Version    string      `json:"version"`
	State      State       `json:"state"`
	Generation uint64      `json:"generation"`
	Since      time.Time   `json:"since"`
	Conditions []Condition `json:"conditions,omitempty"`
}

type Component interface {
	Name() string
	Version() string
	Start(context.Context) error
	Ready(context.Context) error
	Status() Snapshot
	Close(context.Context) error
}

type Reloadable interface {
	Reload(context.Context, Envelope) (ReloadResult, error)
}

type ReloadResult struct {
	Generation      uint64
	Changed         []string
	RestartRequired []string
}
