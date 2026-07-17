package gyre

import "context"

type ConfigValidator interface {
	Validate(context.Context, Envelope) error
}

type ConfigApplier interface {
	Apply(context.Context, Envelope) (ReloadResult, error)
}

// ReloadTransaction validates and applies a generation atomically from the
// caller's perspective: Apply is invoked only after validation succeeds.
func ReloadTransaction(ctx context.Context, envelope Envelope, validator ConfigValidator, applier ConfigApplier) (ReloadResult, error) {
	if err := envelope.Validate(); err != nil {
		return ReloadResult{}, E(CodeConfigInvalid, envelope.Kind, "reload", false, err)
	}
	if validator != nil {
		if err := validator.Validate(ctx, envelope); err != nil {
			return ReloadResult{}, E(CodeConfigInvalid, envelope.Kind, "validate", false, err)
		}
	}
	if applier == nil {
		return ReloadResult{}, E(CodeInternal, envelope.Kind, "apply", false, nil)
	}
	result, err := applier.Apply(ctx, envelope)
	if err != nil {
		return ReloadResult{}, err
	}
	if result.Generation != envelope.Generation {
		result.Generation = envelope.Generation
	}
	return result, nil
}
