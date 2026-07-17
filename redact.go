package gyre

import "strings"

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
