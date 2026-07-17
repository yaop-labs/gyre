package gyre

import (
	"context"
	"errors"
	"sync"
)

type ConfigStore struct {
	mu      sync.Mutex
	current Envelope
	history []Envelope
	subs    map[chan Envelope]struct{}
	audit   []AuditEvent
}

type AuditEvent struct {
	Action     string `json:"action"`
	Generation uint64 `json:"generation"`
	Success    bool   `json:"success"`
}

func (s *ConfigStore) Audit() []AuditEvent {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]AuditEvent{}, s.audit...)
}
func (s *ConfigStore) Validate(envelope Envelope, expected uint64) error {
	if err := envelope.Validate(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if expected != 0 && s.current.Generation != expected {
		return errors.New("gyre: generation conflict")
	}
	if s.current.Generation != 0 && envelope.Generation <= s.current.Generation {
		return errors.New("gyre: generation must increase")
	}
	return nil
}

func (s *ConfigStore) Current() (Envelope, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.current.Generation == 0 {
		return Envelope{}, false
	}
	return s.current, true
}

func (s *ConfigStore) Apply(ctx context.Context, envelope Envelope, expected uint64) error {
	if err := envelope.Validate(); err != nil {
		return E(CodeConfigInvalid, envelope.Kind, "config.apply", false, err)
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if expected != 0 && s.current.Generation != expected {
		return E(CodeConfigInvalid, envelope.Kind, "config.apply", false, errors.New("generation conflict"))
	}
	if s.current.Generation != 0 && envelope.Generation <= s.current.Generation {
		return E(CodeConfigInvalid, envelope.Kind, "config.apply", false, errors.New("generation must increase"))
	}
	if s.current.Generation != 0 {
		s.history = append(s.history, s.current)
	}
	s.current = envelope
	s.audit = append(s.audit, AuditEvent{Action: "apply", Generation: envelope.Generation, Success: true})
	for sub := range s.subs {
		select {
		case sub <- envelope:
		default:
		}
	}
	return nil
}

func (s *ConfigStore) Rollback(ctx context.Context, generation uint64) error {
	s.mu.Lock()
	var target Envelope
	current := s.current.Generation
	for _, e := range s.history {
		if e.Generation == generation {
			target = e
		}
	}
	s.mu.Unlock()
	if target.Generation == 0 {
		return errors.New("gyre: config generation not found")
	}
	// Rollback is a new generation event; callers must provide a monotonic id.
	target.Generation = current + 1
	return s.Apply(ctx, target, 0)
}

func (s *ConfigStore) Watch(ctx context.Context) <-chan Envelope {
	ch := make(chan Envelope, 1)
	s.mu.Lock()
	if s.subs == nil {
		s.subs = map[chan Envelope]struct{}{}
	}
	s.subs[ch] = struct{}{}
	s.mu.Unlock()
	go func() { <-ctx.Done(); s.mu.Lock(); delete(s.subs, ch); close(ch); s.mu.Unlock() }()
	return ch
}
