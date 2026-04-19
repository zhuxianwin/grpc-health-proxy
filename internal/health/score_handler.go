package health

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type scoreStatusResponse struct {
	Threshold float64            `json:"threshold"`
	Entries   []scoreEntryStatus `json:"entries"`
}

type scoreEntryStatus struct {
	Service string  `json:"service"`
	Weight  float64 `json:"weight"`
}

// NewScoreStatusHandler returns an HTTP handler that exposes the score
// checker's configuration as JSON.
func NewScoreStatusHandler(c Checker) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sc, ok := c.(*scoreChecker)
		if !ok {
			http.Error(w, "not a score checker", http.StatusBadRequest)
			return
		}

		sc.mu.Lock()
		entries := make([]scoreEntryStatus, len(sc.entries))
		for i, e := range sc.entries {
			entries[i] = scoreEntryStatus{Service: e.service, Weight: e.weight}
		}
		threshold := sc.cfg.Threshold
		sc.mu.Unlock()

		resp := scoreStatusResponse{
			Threshold: threshold,
			Entries:   entries,
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			http.Error(w, fmt.Sprintf("encode: %v", err), http.StatusInternalServerError)
		}
	})
}
