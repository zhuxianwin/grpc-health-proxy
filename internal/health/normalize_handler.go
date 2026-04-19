package health

import (
	"encoding/json"
	"net/http"
)

// NewNormalizeStatusHandler returns an HTTP handler that reports the active
// normalisation configuration as JSON.
func NewNormalizeStatusHandler(c Checker) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nc, ok := c.(*normalizeChecker)
		if !ok {
			http.Error(w, "not a normalize checker", http.StatusBadRequest)
			return
		}
		payload := struct {
			TrimSpace bool `json:"trim_space"`
			ToLower   bool `json:"to_lower"`
		}{
			TrimSpace: nc.cfg.TrimSpace,
			ToLower:   nc.cfg.ToLower,
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(payload)
	})
}
