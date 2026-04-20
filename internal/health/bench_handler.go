package health

import (
	"encoding/json"
	"net/http"
)

type benchStats struct {
	MeanMs int64 `json:"mean_ms"`
	MaxMs  int64 `json:"max_ms"`
	Window int   `json:"window_size"`
}

// NewBenchStatusHandler returns an HTTP handler that exposes latency
// statistics collected by a BenchChecker as JSON.
//
// If c is not a *BenchChecker the handler responds with 501 Not Implemented.
func NewBenchStatusHandler(c Checker) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bc, ok := c.(*BenchChecker)
		if !ok {
			http.Error(w, "not a BenchChecker", http.StatusNotImplemented)
			return
		}
		mean, max := bc.Stats()
		payload := benchStats{
			MeanMs: mean.Milliseconds(),
			MaxMs:  max.Milliseconds(),
			Window: bc.cfg.WindowSize,
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(payload)
	})
}
