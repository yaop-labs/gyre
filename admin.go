package gyre

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
)

// AdminHandler exposes local, authenticated-by-the-host admin operations. It
// does not implement network authentication; products must mount it behind
// Reef or an equivalent trusted boundary.
func AdminHandler(store *ConfigStore, runtime *Runtime) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /config", func(w http.ResponseWriter, _ *http.Request) {
		if store == nil {
			http.Error(w, "config store unavailable", http.StatusServiceUnavailable)
			return
		}
		cfg, ok := store.Current()
		if !ok {
			http.Error(w, "config not found", http.StatusNotFound)
			return
		}
		writeJSON(w, RedactEnvelope(cfg))
	})
	mux.HandleFunc("POST /config/apply", func(w http.ResponseWriter, r *http.Request) {
		if store == nil {
			http.Error(w, "config store unavailable", http.StatusServiceUnavailable)
			return
		}
		var request struct {
			Expected  uint64 `json:"expected_generation"`
			Component string `json:"component,omitempty"`
			Envelope
		}
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 8<<20)).Decode(&request); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if r.URL.Query().Get("dry_run") == "true" {
			if err := store.Validate(request.Envelope, request.Expected); err != nil {
				store.recordAudit("dry_run", request.Generation, err)
				writeError(w, err)
				return
			}
			store.recordAudit("dry_run", request.Generation, nil)
			w.WriteHeader(http.StatusNoContent)
			return
		}
		apply := runtimeApply(runtime, request.Component)
		if err := store.ApplyWith(r.Context(), request.Envelope, request.Expected, apply); err != nil {
			writeError(w, err)
			return
		}
		w.WriteHeader(http.StatusAccepted)
	})
	mux.HandleFunc("GET /config/watch", func(w http.ResponseWriter, r *http.Request) {
		if store == nil {
			http.Error(w, "config store unavailable", http.StatusServiceUnavailable)
			return
		}
		w.Header().Set("Content-Type", "application/x-ndjson")
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "streaming unsupported", http.StatusNotImplemented)
			return
		}
		for envelope := range store.Watch(r.Context()) {
			_ = json.NewEncoder(w).Encode(RedactEnvelope(envelope))
			flusher.Flush()
		}
	})
	mux.HandleFunc("POST /config/rollback", func(w http.ResponseWriter, r *http.Request) {
		if store == nil {
			http.Error(w, "config store unavailable", http.StatusServiceUnavailable)
			return
		}
		generation, err := strconv.ParseUint(r.URL.Query().Get("generation"), 10, 64)
		if err != nil || generation == 0 {
			http.Error(w, "generation must be positive", http.StatusBadRequest)
			return
		}
		if err := store.RollbackWith(r.Context(), generation, runtimeApply(runtime, r.URL.Query().Get("component"))); err != nil {
			writeError(w, err)
			return
		}
		w.WriteHeader(http.StatusAccepted)
	})
	mux.HandleFunc("GET /status", func(w http.ResponseWriter, _ *http.Request) {
		if runtime == nil {
			http.Error(w, "runtime unavailable", http.StatusServiceUnavailable)
			return
		}
		writeJSON(w, runtime.Status())
	})
	mux.HandleFunc("GET /audit", func(w http.ResponseWriter, _ *http.Request) {
		if store == nil {
			http.Error(w, "config store unavailable", http.StatusServiceUnavailable)
			return
		}
		writeJSON(w, store.Audit())
	})
	return mux
}

func runtimeApply(runtime *Runtime, component string) ConfigApplyFunc {
	if runtime == nil {
		return nil
	}
	return func(ctx context.Context, envelope Envelope) error {
		name := component
		if name == "" {
			name = strings.ToLower(envelope.Kind)
		}
		_, err := runtime.Reload(ctx, name, envelope)
		return err
	}
}

func writeJSON(w http.ResponseWriter, value any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(value)
}
func writeError(w http.ResponseWriter, err error) {
	status := http.StatusBadRequest
	var typed *Error
	if errors.As(err, &typed) && typed.Code == CodeUnavailable {
		status = http.StatusServiceUnavailable
	}
	http.Error(w, safeErrorMessage(err), status)
}
