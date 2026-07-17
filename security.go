package gyre

import "time"

// CredentialStatus is deliberately secret-free and can be backed by Reef's
// credential.Status without importing Reef into Gyre core.
type CredentialStatus struct {
	Name        string    `json:"name"`
	Kind        string    `json:"kind"`
	Generation  uint64    `json:"generation"`
	LastSuccess time.Time `json:"last_success,omitempty"`
	LastFailure time.Time `json:"last_failure,omitempty"`
	LastError   string    `json:"last_error,omitempty"`
	NotBefore   time.Time `json:"not_before,omitempty"`
	NotAfter    time.Time `json:"not_after,omitempty"`
}

type CredentialEvent struct {
	Status  CredentialStatus `json:"status"`
	Success bool             `json:"success"`
	Changed bool             `json:"changed"`
}

type CredentialObserver interface{ ObserveCredential(CredentialEvent) }
type CredentialObserverFunc func(CredentialEvent)

func (f CredentialObserverFunc) ObserveCredential(e CredentialEvent) { f(e) }
