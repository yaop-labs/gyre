package gyre

import (
	"encoding/json"
	"fmt"
)

type Envelope struct {
	APIVersion string          `json:"api_version" yaml:"api_version"`
	Kind       string          `json:"kind" yaml:"kind"`
	Generation uint64          `json:"generation" yaml:"generation"`
	Spec       json.RawMessage `json:"spec" yaml:"spec"`
}

func (e Envelope) Validate() error {
	if e.APIVersion == "" || e.Kind == "" {
		return fmt.Errorf("gyre: api_version and kind are required")
	}
	if e.Generation == 0 {
		return fmt.Errorf("gyre: generation must be positive")
	}
	if len(e.Spec) == 0 {
		return fmt.Errorf("gyre: spec is required")
	}
	return nil
}
