package health

import (
	"encoding/json"
	"net/http"
	"time"
)

type quotaWindowStatus struct {
	Service    string        `json:"service"`
	Window     time.Duration `json:"window_ns"`
	MaxCalls   int           `json:"max_calls"`
	CurrentUse int           `json:"current_use"`
	Exceeded   bool          `json:"exceeded"`
}

// NewQuotaWindowStatusHandler returns an HTTP handler that reports the
// sliding-window quota state for every tracked service.
func NewQuotaWindowStatusHandler(c Checker) http.Handler {
	qw, ok := c.(*quotaWindowChecker)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if !ok {
			w.WriteHeader(http.StatusNotImplemented)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "not a quota_window checker"})
			return
		}

		now := time.Now()
		cutoff := now.Add(-qw.cfg.Window)

		qw.mu.Lock()
		var statuses []quotaWindowStatus
		for svc, ts := range qw.times {
			count := 0
			for _, t := range ts {
				if t.After(cutoff) {
					count++
				}
			}
			statuses = append(statuses, quotaWindowStatus{
				Service:    svc,
				Window:     qw.cfg.Window,
				MaxCalls:   qw.cfg.MaxCalls,
				CurrentUse: count,
				Exceeded:   count >= qw.cfg.MaxCalls,
			})
		}
		qw.mu.Unlock()

		_ = json.NewEncoder(w).Encode(statuses)
	})
}
