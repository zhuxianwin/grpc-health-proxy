package health

import (
	"encoding/json"
	"net/http"
)

type burstStatusResponse struct {
	MaxBurst    int    `json:"max_burst"`
	WindowSecs  float64 `json:"window_seconds"`
	CurrentCount int   `json:"current_count"`
}

// NewBurstStatusHandler returns an HTTP handler that reports the current
// burst-checker configuration and live call count for the rolling window.
func NewBurstStatusHandler(c Checker) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bc, ok := c.(*burstChecker)
		if !ok {
			http.Error(w, "not a burst checker", http.StatusBadRequest)
			return
		}

		bc.mu.Lock()
		count := len(bc.times)
		bc.mu.Unlock()

		resp := burstStatusResponse{
			MaxBurst:    bc.cfg.MaxBurst,
			WindowSecs:  bc.cfg.Window.Seconds(),
			CurrentCount: count,
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	})
}
