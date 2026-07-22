package gyre_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/yaop-labs/gyre"
)

type component struct{ ready bool }

func (c *component) Name() string                { return "test" }
func (c *component) Version() string             { return "0.1.0" }
func (c *component) Start(context.Context) error { return nil }
func (c *component) Ready(context.Context) error {
	if !c.ready {
		return errors.New("starting")
	}
	return nil
}
func (c *component) Status() gyre.Snapshot {
	return gyre.Snapshot{Name: "test", Version: "0.1.0", State: gyre.StateReady, Since: time.Unix(1, 0)}
}
func (c *component) Close(context.Context) error { return nil }

func TestHTTPHandler(t *testing.T) {
	c := &component{}
	h := gyre.HTTPHandler(c)
	for _, path := range []string{"/healthz", "/status"} {
		r := httptest.NewRequest(http.MethodGet, path, nil)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		if w.Code != http.StatusOK {
			t.Fatalf("%s: status=%d", path, w.Code)
		}
	}
	r := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("not ready status=%d", w.Code)
	}
	c.ready = true
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("ready status=%d", w.Code)
	}
}

func TestChecker(t *testing.T) {
	var c gyre.Checker
	if err := c.Add(gyre.Check{}); err == nil {
		t.Fatal("expected invalid check error")
	}
	if err := c.Add(gyre.Check{Name: "live", Scope: gyre.ScopeLiveness, Check: func(context.Context) error { return errors.New("ignored") }}); err != nil {
		t.Fatal(err)
	}
	if err := c.Add(gyre.Check{Name: "db", Scope: gyre.ScopeReadiness, Check: func(context.Context) error { return errors.New("down") }}); err != nil {
		t.Fatal(err)
	}
	if err := c.Ready(context.Background()); err == nil {
		t.Fatal("expected readiness error")
	}
}
