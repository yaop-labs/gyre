package gyre_test

import (
	"encoding/json"
	"github.com/yaop-labs/gyre"
	"testing"
)

func TestRedactConfig(t *testing.T) {
	r := gyre.RedactConfig(map[string]string{"endpoint": "collector:4317", "bearer_token": "secret"})
	if r["endpoint"] != "collector:4317" || r["bearer_token"] != "[REDACTED]" {
		t.Fatalf("redacted=%v", r)
	}
}

func TestRedactEnvelope(t *testing.T) {
	envelope := gyre.Envelope{
		APIVersion: "gyre/v1",
		Kind:       "Wisp",
		Generation: 1,
		Spec:       json.RawMessage(`{"password":"one","nested":{"api_key":"two"},"endpoint":"safe"}`),
	}
	redacted := gyre.RedactEnvelope(envelope)
	if string(redacted.Spec) != `{"endpoint":"safe","nested":{"api_key":"[REDACTED]"},"password":"[REDACTED]"}` {
		t.Fatalf("redacted=%s", redacted.Spec)
	}
	if string(envelope.Spec) == string(redacted.Spec) {
		t.Fatal("redaction mutated or reused original spec")
	}
}

func TestRedactEnvelopeOmitsInvalidJSON(t *testing.T) {
	envelope := gyre.Envelope{Spec: json.RawMessage(`{"token":"secret"`)}
	redacted := gyre.RedactEnvelope(envelope)
	if string(redacted.Spec) == string(envelope.Spec) {
		t.Fatal("invalid secret-bearing JSON was returned verbatim")
	}
}
