package health

import (
	"encoding/json"
	"net/http"
	"time"
)

type checkpointEntry struct {
	Service    string    `json:"service"`
	Status     string    `json:"status"`
	RecordedAt time.Time `json:"recorded_at"`
	Err        string    `json:"error,omitempty"`
}

// NewCheckpointStatusHandler returns an HTTP handler that exposes the
// current checkpoint store as JSON.
func NewCheckpointStatusHandler(store *CheckpointStore) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		all := store.All()
		out := make([]checkpointEntry, 0, len(all))
		for _, e := range all {
			ce := checkpointEntry{
				Service:    e.Service,
				Status:     e.Result.Status.String(),
				RecordedAt: e.RecordedAt,
			}
			if e.Result.Err != nil {
				ce.Err = e.Result.Err.Error()
			}
			out = append(out, ce)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"checkpoints": out,
			"count":       len(out),
		})
	})
}
