package health

import (
	"encoding/json"
	"net/http"
	"time"
)

type stickyStatusResponse struct {
	Service    string  `json:"service"`
	Stuck      bool    `json:"stuck"`
	StuckUntil *string `json:"stuck_until,omitempty"`
}

// NewStickyStatusHandler returns an HTTP handler that reports sticky state
// for all tracked services on the given stickyChecker.
func NewStickyStatusHandler(c Checker) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sc, ok := c.(*stickyChecker)
		if !ok {
			http.Error(w, "not a sticky checker", http.StatusBadRequest)
			return
		}

		sc.mu.Lock()
		var rows []stickyStatusResponse
		now := time.Now()
		for svc, entry := range sc.entries {
			stuck := entry.result.Status == StatusUnhealthy && now.Before(entry.stuckUntil)
			row := stickyStatusResponse{Service: svc, Stuck: stuck}
			if stuck {
				t := entry.stuckUntil.Format(time.RFC3339)
				row.StuckUntil = &t
			}
			rows = append(rows, row)
		}
		sc.mu.Unlock()

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(rows)
	})
}
