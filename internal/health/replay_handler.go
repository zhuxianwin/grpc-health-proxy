package health

import (
	"encoding/json"
	"net/http"
)

type replayStatusEntry struct {
	Service string   `json:"service"`
	History []string `json:"history"`
	Total   int      `json:"total"`
}

// NewReplayStatusHandler returns an HTTP handler that exposes the recorded
// result history for every service tracked by a ReplayChecker.
func NewReplayStatusHandler(rc *ReplayChecker, services []string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var entries []replayStatusEntry
		for _, svc := range services {
			history := rc.History(svc)
			strs := make([]string, len(history))
			for i, res := range history {
				strs[i] = res.Status.String()
			}
			entries = append(entries, replayStatusEntry{
				Service: svc,
				History: strs,
				Total:   len(history),
			})
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(entries)
	})
}
