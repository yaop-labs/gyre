package gyre_test

import (
	"bytes"
	"encoding/json"
	"github.com/yaop-labs/gyre"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAdminApplyAndCurrent(t *testing.T) {
	var store gyre.ConfigStore
	h := gyre.AdminHandler(&store, nil)
	body := []byte(`{"api_version":"gyre/v1","kind":"Wisp","generation":1,"spec":{}}`)
	r := httptest.NewRequest(http.MethodPost, "/config/apply", bytes.NewReader(body))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != http.StatusAccepted {
		t.Fatalf("apply status=%d body=%s", w.Code, w.Body.String())
	}
	r = httptest.NewRequest(http.MethodGet, "/config", nil)
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("config status=%d", w.Code)
	}
	var got gyre.Envelope
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil || got.Generation != 1 {
		t.Fatalf("config=%+v err=%v", got, err)
	}
}

func TestAdminDryRunAndAudit(t *testing.T) {
	var store gyre.ConfigStore
	h := gyre.AdminHandler(&store, nil)
	body := []byte(`{"api_version":"gyre/v1","kind":"Wisp","generation":1,"spec":{}}`)
	r := httptest.NewRequest(http.MethodPost, "/config/apply?dry_run=true", bytes.NewReader(body))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != http.StatusNoContent {
		t.Fatalf("dry-run status=%d", w.Code)
	}
	r = httptest.NewRequest(http.MethodGet, "/audit", nil)
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != http.StatusOK || w.Body.String() == "null\n" {
		t.Fatalf("audit status=%d body=%s", w.Code, w.Body.String())
	}
}
