package health

import (
	"encoding/json"
	"net/http"
	"time"
)

type warmupStatusResponse struct {
	WarmedUp  bool          `json:"warmed_up"`
	Elapsed   string        `json:"elapsed"`
	Successes int           `json:"successes"`
	MinChecks int           `json:"min_checks"`
	Duration  time.Duration `json:"warmup_duration_ms"`
}

// NewWarmupStatusHandler returns an HTTP handler that exposes warmup state.
// If the checker is not a *warmupChecker it returns 501.
func NewWarmupStatusHandler(c Checker) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wc, ok := c.(*warmupChecker)
		if !ok {
			http.Error(w, "not a warmup checker", http.StatusNotImplemented)
			return
		}

		elapsed := time.Since(wc.start)
		warmedUp := elapsed >= wc.cfg.Duration && wc.successes >= wc.cfg.MinChecks

		resp := warmupStatusResponse{
			WarmedUp:  warmedUp,
			Elapsed:   elapsed.Round(time.Millisecond).String(),
			Successes: wc.successes,
			MinChecks: wc.cfg.MinChecks,
			Duration:  wc.cfg.Duration / time.Millisecond,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp) //nolint:errcheck
	})
}
