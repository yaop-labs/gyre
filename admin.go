package gyre

import (
	"encoding/json"
	"net/http"
	"strconv"
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
		writeJSON(w, cfg)
	})
	mux.HandleFunc("POST /config/apply", func(w http.ResponseWriter, r *http.Request) {
		if store == nil {
			http.Error(w, "config store unavailable", http.StatusServiceUnavailable)
			return
		}
		var request struct {
			Expected uint64 `json:"expected_generation"`
			Envelope
		}
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 8<<20)).Decode(&request); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if r.URL.Query().Get("dry_run") == "true" {
			if err := store.Validate(request.Envelope, request.Expected); err != nil {
				writeError(w, err)
				return
			}
			w.WriteHeader(http.StatusNoContent)
			return
		}
		if err := store.Apply(r.Context(), request.Envelope, request.Expected); err != nil {
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
			_ = json.NewEncoder(w).Encode(envelope)
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
		if err := store.Rollback(r.Context(), generation); err != nil {
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

func writeJSON(w http.ResponseWriter, value any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(value)
}
func writeError(w http.ResponseWriter, err error) {
	status := http.StatusBadRequest
	if e, ok := err.(*Error); ok && e.Code == CodeUnavailable {
		status = http.StatusServiceUnavailable
	}
	http.Error(w, err.Error(), status)
}
