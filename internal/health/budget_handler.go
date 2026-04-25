package health

import (
	"encoding/json"
	"net/http"
)

type budgetStatusResponse struct {
	Service    string  `json:"service"`
	Total      int     `json:"total"`
	Failures   int     `json:"failures"`
	ErrorRate  float64 `json:"error_rate"`
	Budget     float64 `json:"budget"`
	Exhausted  bool    `json:"exhausted"`
}

// NewBudgetStatusHandler returns an HTTP handler that reports the current
// error-budget state for every service tracked by bc.
func NewBudgetStatusHandler(bc *BudgetChecker) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		service := r.URL.Query().Get("service")
		if service == "" {
			http.Error(w, "missing service query param", http.StatusBadRequest)
			return
		}

		total, failures := bc.BudgetStats(service)
		var errorRate float64
		if total > 0 {
			errorRate = float64(failures) / float64(total)
		}

		resp := budgetStatusResponse{
			Service:   service,
			Total:     total,
			Failures:  failures,
			ErrorRate: errorRate,
			Budget:    bc.cfg.ErrorBudget,
			Exhausted: total >= bc.cfg.MinSamples && errorRate > bc.cfg.ErrorBudget,
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	})
}
