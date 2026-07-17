package gyre

import (
	"encoding/json"
	"net/http"
)

// HTTPHandler exposes the stable operational endpoints for a component.
// Products can mount it into an existing mux.
func HTTPHandler(component Component) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok\n"))
	})
	mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, r *http.Request) {
		if component == nil {
			http.Error(w, "component unavailable", http.StatusServiceUnavailable)
			return
		}
		if err := component.Ready(r.Context()); err != nil {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ready\n"))
	})
	mux.HandleFunc("GET /status", func(w http.ResponseWriter, _ *http.Request) {
		if component == nil {
			http.Error(w, "component unavailable", http.StatusServiceUnavailable)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(component.Status()); err != nil {
			return
		}
	})
	return mux
}
