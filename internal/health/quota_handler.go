package health

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

type quotaStatusResponse struct {
	MaxCalls int64  `json:"max_calls"`
	Window   string `json:"window"`
	Current  int64  `json:"current_count"`
	Exhausted bool  `json:"exhausted"`
}

// NewQuotaStatusHandler returns an HTTP handler that exposes quota state.
// It accepts a *quotaChecker cast from the provided Checker.
func NewQuotaStatusHandler(c Checker, logger *slog.Logger) http.Handler {
	if logger == nil {
		logger = slog.Default()
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		qc, ok := c.(*quotaChecker)
		if !ok {
			http.Error(w, "not a quota checker", http.StatusInternalServerError)
			return
		}
		current := qc.count.Load()
		resp := quotaStatusResponse{
			MaxCalls:  qc.cfg.MaxCalls,
			Window:    qc.cfg.Window.String(),
			Current:   current,
			Exhausted: current >= qc.cfg.MaxCalls,
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			logger.Error("failed to encode quota status", "err", err)
		}
	})
}
