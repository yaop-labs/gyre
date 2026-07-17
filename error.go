package gyre

import "fmt"

type Code string

const (
	CodeConfigInvalid Code = "config_invalid"
	CodeUnavailable   Code = "unavailable"
	CodeOverloaded    Code = "overloaded"
	CodeUnauthorized  Code = "unauthorized"
	CodeDependency    Code = "dependency"
	CodeCorruption    Code = "corruption"
	CodeShuttingDown  Code = "shutting_down"
	CodeInternal      Code = "internal"
)

type Error struct {
	Code      Code
	Component string
	Operation string
	Retryable bool
	Cause     error
}

func (e *Error) Error() string {
	if e == nil {
		return "<nil>"
	}
	if e.Cause == nil {
		return fmt.Sprintf("gyre: %s", e.Code)
	}
	return fmt.Sprintf("gyre: %s: %v", e.Code, e.Cause)
}
func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Cause
}

func E(code Code, component, operation string, retryable bool, cause error) *Error {
	return &Error{Code: code, Component: component, Operation: operation, Retryable: retryable, Cause: cause}
}
