package gyre_test

import (
	"github.com/yaop-labs/gyre"
	"testing"
)

func TestRedactConfig(t *testing.T) {
	r := gyre.RedactConfig(map[string]string{"endpoint": "collector:4317", "bearer_token": "secret"})
	if r["endpoint"] != "collector:4317" || r["bearer_token"] != "[REDACTED]" {
		t.Fatalf("redacted=%v", r)
	}
}
