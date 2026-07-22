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

func TestAdminRedactsConfig(t *testing.T) {
	var store gyre.ConfigStore
	h := gyre.AdminHandler(&store, nil)
	body := []byte(`{"api_version":"gyre/v1","kind":"Wisp","generation":1,"spec":{"endpoint":"collector:4317","auth":{"bearer_token":"secret"}}}`)
	r := httptest.NewRequest(http.MethodPost, "/config/apply", bytes.NewReader(body))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != http.StatusAccepted {
		t.Fatalf("apply status=%d body=%s", w.Code, w.Body.String())
	}
	r = httptest.NewRequest(http.MethodGet, "/config", nil)
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != http.StatusOK || bytes.Contains(w.Body.Bytes(), []byte("secret")) {
		t.Fatalf("unsafe config response: status=%d body=%s", w.Code, w.Body.String())
	}
	if !bytes.Contains(w.Body.Bytes(), []byte("[REDACTED]")) {
		t.Fatalf("missing redaction marker: %s", w.Body.String())
	}
}

func TestAdminApplyCoordinatesRuntimeReload(t *testing.T) {
	component := &reloadComponent{component: component{ready: true}}
	var runtime gyre.Runtime
	if err := runtime.Add(component); err != nil {
		t.Fatal(err)
	}
	var store gyre.ConfigStore
	h := gyre.AdminHandler(&store, &runtime)
	body := []byte(`{"api_version":"gyre/v1","kind":"Test","generation":1,"spec":{}}`)
	r := httptest.NewRequest(http.MethodPost, "/config/apply", bytes.NewReader(body))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != http.StatusAccepted {
		t.Fatalf("apply status=%d body=%s", w.Code, w.Body.String())
	}
	if component.generation != 1 {
		t.Fatalf("component generation=%d", component.generation)
	}
	if current, ok := store.Current(); !ok || current.Generation != 1 {
		t.Fatalf("current=%+v ok=%v", current, ok)
	}
}

func TestAdminRejectedReloadIsSafeAndDoesNotCommit(t *testing.T) {
	var runtime gyre.Runtime
	if err := runtime.Add(&component{ready: true}); err != nil {
		t.Fatal(err)
	}
	var store gyre.ConfigStore
	h := gyre.AdminHandler(&store, &runtime)
	body := []byte(`{"api_version":"gyre/v1","kind":"Test","generation":1,"spec":{"token":"secret"}}`)
	r := httptest.NewRequest(http.MethodPost, "/config/apply", bytes.NewReader(body))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != http.StatusServiceUnavailable || bytes.Contains(w.Body.Bytes(), []byte("secret")) {
		t.Fatalf("unsafe rejection: status=%d body=%s", w.Code, w.Body.String())
	}
	if _, ok := store.Current(); ok {
		t.Fatal("rejected runtime reload committed config")
	}
	audit := store.Audit()
	if len(audit) != 1 || audit[0].Success || audit[0].Error != "gyre: unavailable" {
		t.Fatalf("audit=%+v", audit)
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
