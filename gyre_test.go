package gyre_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/yaop-labs/gyre"
)

func TestEnvelopeValidate(t *testing.T) {
	e := gyre.Envelope{APIVersion: "gyre/v1", Kind: "Wisp", Generation: 1, Spec: json.RawMessage(`{}`)}
	if err := e.Validate(); err != nil {
		t.Fatal(err)
	}
	for _, bad := range []gyre.Envelope{{Kind: "Wisp", Generation: 1, Spec: json.RawMessage(`{}`)}, e} {
		bad.APIVersion = ""
		if err := bad.Validate(); err == nil {
			t.Fatal("expected validation error")
		}
	}
}

func TestResourceMerge(t *testing.T) {
	r, err := gyre.MergeResource(gyre.Resource{"service.name": "wisp", "host.name": "node"}, gyre.Resource{"env": "prod"})
	if err != nil || r["env"] != "prod" {
		t.Fatalf("resource=%v err=%v", r, err)
	}
	if _, err := gyre.MergeResource(gyre.Resource{"service.name": "wisp"}, gyre.Resource{"service.name": "coral"}); err == nil {
		t.Fatal("expected identity conflict")
	}
}

func TestTypedError(t *testing.T) {
	cause := errors.New("temporary")
	err := gyre.E(gyre.CodeUnavailable, "wisp", "connect", true, cause)
	if !errors.Is(err, cause) || !err.Retryable || err.Code != gyre.CodeUnavailable {
		t.Fatalf("error=%+v", err)
	}
}
