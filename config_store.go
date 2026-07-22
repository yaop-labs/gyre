package gyre

import (
	"context"
	"errors"
	"sync"
	"time"
)

type ConfigStore struct {
	applyMu sync.Mutex
	mu      sync.Mutex
	current Envelope
	history []Envelope
	subs    map[chan Envelope]struct{}
	audit   []AuditEvent
}

type AuditEvent struct {
	Action     string    `json:"action"`
	Generation uint64    `json:"generation"`
	Success    bool      `json:"success"`
	Error      string    `json:"error,omitempty"`
	Timestamp  time.Time `json:"timestamp"`
}

type ConfigApplyFunc func(context.Context, Envelope) error

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
	return s.validateLocked(envelope, expected)
}

func (s *ConfigStore) validateLocked(envelope Envelope, expected uint64) error {
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
	return cloneEnvelope(s.current), true
}

func (s *ConfigStore) Apply(ctx context.Context, envelope Envelope, expected uint64) error {
	return s.ApplyWith(ctx, envelope, expected, nil)
}

// ApplyWith commits envelope only after apply succeeds. The generation check,
// product apply, and store commit are serialized as one transaction from the
// perspective of other ConfigStore operations.
func (s *ConfigStore) ApplyWith(ctx context.Context, envelope Envelope, expected uint64, apply ConfigApplyFunc) error {
	return s.applyWithAction(ctx, "apply", envelope, expected, apply)
}

func (s *ConfigStore) applyWithAction(ctx context.Context, action string, envelope Envelope, expected uint64, apply ConfigApplyFunc) error {
	if err := envelope.Validate(); err != nil {
		typed := E(CodeConfigInvalid, envelope.Kind, "config."+action, false, err)
		s.recordAudit(action, envelope.Generation, typed)
		return typed
	}
	if err := ctx.Err(); err != nil {
		s.recordAudit(action, envelope.Generation, err)
		return err
	}
	s.applyMu.Lock()
	defer s.applyMu.Unlock()
	return s.applyWithActionLocked(ctx, action, envelope, expected, apply)
}

func (s *ConfigStore) applyWithActionLocked(ctx context.Context, action string, envelope Envelope, expected uint64, apply ConfigApplyFunc) error {
	s.mu.Lock()
	if err := s.validateLocked(envelope, expected); err != nil {
		typed := E(CodeConfigInvalid, envelope.Kind, "config."+action, false, err)
		s.auditLocked(action, envelope.Generation, typed)
		s.mu.Unlock()
		return typed
	}
	s.mu.Unlock()
	if apply != nil {
		if err := apply(ctx, cloneEnvelope(envelope)); err != nil {
			s.recordAudit(action, envelope.Generation, err)
			return err
		}
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.current.Generation != 0 {
		s.history = append(s.history, cloneEnvelope(s.current))
	}
	s.current = cloneEnvelope(envelope)
	s.auditLocked(action, envelope.Generation, nil)
	for sub := range s.subs {
		select {
		case sub <- cloneEnvelope(envelope):
		default:
		}
	}
	return nil
}

func (s *ConfigStore) Rollback(ctx context.Context, generation uint64) error {
	return s.RollbackWith(ctx, generation, nil)
}

// RollbackWith reapplies a historical config as a new monotonic generation and
// commits it only after apply succeeds.
func (s *ConfigStore) RollbackWith(ctx context.Context, generation uint64, apply ConfigApplyFunc) error {
	s.applyMu.Lock()
	defer s.applyMu.Unlock()
	s.mu.Lock()
	var target Envelope
	current := s.current.Generation
	for _, e := range s.history {
		if e.Generation == generation {
			target = cloneEnvelope(e)
		}
	}
	s.mu.Unlock()
	if target.Generation == 0 {
		err := errors.New("gyre: config generation not found")
		s.recordAudit("rollback", generation, err)
		return err
	}
	// Rollback is a new generation event; callers must provide a monotonic id.
	target.Generation = current + 1
	return s.applyWithActionLocked(ctx, "rollback", target, current, apply)
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

func (s *ConfigStore) recordAudit(action string, generation uint64, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.auditLocked(action, generation, err)
}

func (s *ConfigStore) auditLocked(action string, generation uint64, err error) {
	event := AuditEvent{
		Action:     action,
		Generation: generation,
		Success:    err == nil,
		Timestamp:  time.Now().UTC(),
	}
	if err != nil {
		event.Error = safeErrorMessage(err)
	}
	s.audit = append(s.audit, event)
}

func cloneEnvelope(envelope Envelope) Envelope {
	envelope.Spec = append(envelope.Spec[:0:0], envelope.Spec...)
	return envelope
}
