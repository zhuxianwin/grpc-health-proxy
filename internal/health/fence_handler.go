package health

import (
	"encoding/json"
	"net/http"
)

// NewFenceStatusHandler returns an http.Handler that reports the current fence
// state and exposes POST /raise and POST /lower sub-paths to control it.
func NewFenceStatusHandler(c Checker) http.Handler {
	mux := http.NewServeMux()

	f, ok := c.(*FenceChecker)

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if !ok {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "not a FenceChecker"})
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]bool{"raised": f.IsRaised()})
	})

	mux.HandleFunc("/raise", func(w http.ResponseWriter, r *http.Request) {
		if !ok {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		f.Raise()
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]bool{"raised": true})
	})

	mux.HandleFunc("/lower", func(w http.ResponseWriter, r *http.Request) {
		if !ok {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		f.Lower()
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]bool{"raised": false})
	})

	return mux
}
