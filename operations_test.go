package gyre_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/yaop-labs/gyre"
)

type validatorFunc func(context.Context, gyre.Envelope) error

func (f validatorFunc) Validate(ctx context.Context, envelope gyre.Envelope) error {
	return f(ctx, envelope)
}

type applierFunc func(context.Context, gyre.Envelope) (gyre.ReloadResult, error)

func (f applierFunc) Apply(ctx context.Context, envelope gyre.Envelope) (gyre.ReloadResult, error) {
	return f(ctx, envelope)
}

func TestReloadTransaction(t *testing.T) {
	envelope := gyre.Envelope{APIVersion: "gyre/v1", Kind: "Test", Generation: 7, Spec: json.RawMessage(`{}`)}
	validated := false
	result, err := gyre.ReloadTransaction(
		context.Background(),
		envelope,
		validatorFunc(func(context.Context, gyre.Envelope) error {
			validated = true
			return nil
		}),
		applierFunc(func(context.Context, gyre.Envelope) (gyre.ReloadResult, error) {
			return gyre.ReloadResult{Changed: []string{"endpoint"}}, nil
		}),
	)
	if err != nil || !validated || result.Generation != envelope.Generation {
		t.Fatalf("result=%+v validated=%v err=%v", result, validated, err)
	}

	rejected := errors.New("invalid product config")
	_, err = gyre.ReloadTransaction(context.Background(), envelope, validatorFunc(func(context.Context, gyre.Envelope) error {
		return rejected
	}), applierFunc(func(context.Context, gyre.Envelope) (gyre.ReloadResult, error) {
		t.Fatal("apply called after validation failure")
		return gyre.ReloadResult{}, nil
	}))
	var typed *gyre.Error
	if !errors.As(err, &typed) || typed.Code != gyre.CodeConfigInvalid {
		t.Fatalf("error=%v", err)
	}
	if _, err := gyre.ReloadTransaction(context.Background(), envelope, nil, nil); err == nil {
		t.Fatal("expected missing applier error")
	}
}

func TestRuntimeHTTPHandler(t *testing.T) {
	component := &component{}
	var runtime gyre.Runtime
	if err := runtime.Add(component); err != nil {
		t.Fatal(err)
	}
	handler := gyre.RuntimeHTTPHandler(&runtime)

	request := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("not-ready status=%d", response.Code)
	}
	component.ready = true
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("ready status=%d", response.Code)
	}
	for _, path := range []string{"/healthz", "/status"} {
		response = httptest.NewRecorder()
		handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))
		if response.Code != http.StatusOK {
			t.Fatalf("%s status=%d", path, response.Code)
		}
	}

	nilResponse := httptest.NewRecorder()
	gyre.RuntimeHTTPHandler(nil).ServeHTTP(nilResponse, request)
	if nilResponse.Code != http.StatusServiceUnavailable {
		t.Fatalf("nil runtime status=%d", nilResponse.Code)
	}
}

type metricsSink struct {
	name      string
	operation string
	err       error
}

func (s *metricsSink) Observe(name, operation string, _ time.Duration, err error) {
	s.name, s.operation, s.err = name, operation, err
}

func TestObserversAndHelpers(t *testing.T) {
	sink := &metricsSink{}
	wantErr := errors.New("failed")
	gyre.RuntimeMetrics{Sink: sink}.Observe("wisp", "reload", time.Now(), wantErr)
	if sink.name != "wisp" || sink.operation != "reload" || !errors.Is(sink.err, wantErr) {
		t.Fatalf("metrics observation=%+v", sink)
	}
	gyre.RuntimeMetrics{}.Observe("ignored", "ignored", time.Now(), nil)

	var observed gyre.CredentialEvent
	observer := gyre.CredentialObserverFunc(func(event gyre.CredentialEvent) { observed = event })
	observer.ObserveCredential(gyre.CredentialEvent{Success: true, Changed: true})
	if !observed.Success || !observed.Changed {
		t.Fatalf("credential event=%+v", observed)
	}

	resource := gyre.Resource{"service.name": "wisp"}
	clone := resource.Clone()
	clone["service.name"] = "changed"
	if resource["service.name"] != "wisp" {
		t.Fatal("resource clone aliases the original")
	}
}

func TestRunWrapper(t *testing.T) {
	var runtime gyre.Runtime
	if err := runtime.Add(&component{ready: true}); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := gyre.Run(ctx, &runtime); err != nil {
		t.Fatal(err)
	}
}
