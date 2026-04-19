package health

import (
	"encoding/json"
	"net/http"
)

// NewClampStatusHandler returns an HTTP handler that exposes the current
// clamp configuration as JSON.
func NewClampStatusHandler(c Checker) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cc, ok := c.(*clampChecker)
		if !ok {
			http.Error(w, "not a clamp checker", http.StatusBadRequest)
			return
		}
		payload := struct {
			MinStatus string `json:"min_status"`
			MaxStatus string `json:"max_status"`
		}{
			MinStatus: cc.cfg.MinStatus.String(),
			MaxStatus: cc.cfg.MaxStatus.String(),
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(payload)
	})
}
