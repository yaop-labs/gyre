package gyre

import (
	"encoding/json"
	"strings"
)

// RedactConfig removes values for fields conventionally carrying secrets.
// It is intended for status/admin output, never for applying configuration.
func RedactConfig(values map[string]string) map[string]string {
	out := make(map[string]string, len(values))
	for key, value := range values {
		lower := strings.ToLower(key)
		if strings.Contains(lower, "token") || strings.Contains(lower, "password") || strings.Contains(lower, "secret") || strings.Contains(lower, "private_key") {
			out[key] = "[REDACTED]"
			continue
		}
		out[key] = value
	}
	return out
}

// RedactEnvelope recursively removes conventionally secret values from a JSON
// config envelope. Invalid JSON is omitted rather than returned verbatim.
func RedactEnvelope(envelope Envelope) Envelope {
	var value any
	if err := json.Unmarshal(envelope.Spec, &value); err != nil {
		envelope.Spec = json.RawMessage(`{"_redacted":"invalid configuration omitted"}`)
		return envelope
	}
	redactValue(value)
	redacted, err := json.Marshal(value)
	if err != nil {
		envelope.Spec = json.RawMessage(`{"_redacted":"configuration omitted"}`)
		return envelope
	}
	envelope.Spec = redacted
	return envelope
}

func redactValue(value any) {
	switch value := value.(type) {
	case map[string]any:
		for key, child := range value {
			if sensitiveKey(key) {
				value[key] = "[REDACTED]"
				continue
			}
			redactValue(child)
		}
	case []any:
		for _, child := range value {
			redactValue(child)
		}
	}
}

func sensitiveKey(key string) bool {
	lower := strings.ToLower(key)
	return strings.Contains(lower, "token") ||
		strings.Contains(lower, "password") ||
		strings.Contains(lower, "secret") ||
		strings.Contains(lower, "private_key") ||
		strings.Contains(lower, "api_key") ||
		strings.Contains(lower, "authorization")
}
