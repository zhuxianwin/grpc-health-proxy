package health

import (
	"encoding/json"
	"net/http"
)

type slopeStatusResponse struct {
	Window    int     `json:"window"`
	Threshold float64 `json:"threshold"`
}

// NewSlopeStatusHandler returns an HTTP handler that reports the current
// SlopeChecker configuration. Returns 400 if c is not a *slopeChecker.
func NewSlopeStatusHandler(c Checker) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sc, ok := c.(*slopeChecker)
		if !ok {
			http.Error(w, "not a slope checker", http.StatusBadRequest)
			return
		}
		sc.mu.Lock()
		resp := slopeStatusResponse{
			Window:    sc.cfg.Window,
			Threshold: sc.cfg.Threshold,
		}
		sc.mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp) //nolint:errcheck
	})
}
