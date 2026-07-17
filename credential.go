package gyre

// CredentialSource is the narrow bridge for Reef or another provider. Gyre
// observes lifecycle only; loading, rotation, and policy stay provider-owned.
type CredentialSource interface {
	CredentialStatus() []CredentialStatus
	Close() error
}

type CredentialStatusProvider interface {
	CredentialSource
	SetObserver(CredentialObserver)
}
