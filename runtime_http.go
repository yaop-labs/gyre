package gyre

import (
	"encoding/json"
	"net/http"
)

// RuntimeHTTPHandler exposes aggregate operational endpoints.
func RuntimeHTTPHandler(runtime *Runtime) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok\n"))
	})
	mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, r *http.Request) {
		if runtime == nil {
			http.Error(w, "runtime unavailable", http.StatusServiceUnavailable)
			return
		}
		if err := runtime.Ready(r.Context()); err != nil {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ready\n"))
	})
	mux.HandleFunc("GET /status", func(w http.ResponseWriter, _ *http.Request) {
		if runtime == nil {
			http.Error(w, "runtime unavailable", http.StatusServiceUnavailable)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(runtime.Status())
	})
	return mux
}
