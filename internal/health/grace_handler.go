package health

import (
	"encoding/json"
	"net/http"
	"time"
)

type graceStatusResponse struct {
	InGrace    bool          `json:"in_grace"`
	Elapsed    string        `json:"elapsed"`
	Window     string        `json:"window"`
	Checks     int64         `json:"checks"`
	MinChecks  int           `json:"min_checks"`
}

// NewGraceStatusHandler returns an HTTP handler that reports the current
// state of the grace period for a graceChecker.
func NewGraceStatusHandler(c Checker) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gc, ok := c.(*graceChecker)
		if !ok {
			http.Error(w, "checker is not a GraceChecker", http.StatusBadRequest)
			return
		}

		elapsed := time.Since(gc.start)
		checks := gc.count.Load()
		inGrace := elapsed < gc.cfg.Window || checks < int64(gc.cfg.MinChecks)

		resp := graceStatusResponse{
			InGrace:   inGrace,
			Elapsed:   elapsed.Round(time.Millisecond).String(),
			Window:    gc.cfg.Window.String(),
			Checks:    checks,
			MinChecks: gc.cfg.MinChecks,
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	})
}
