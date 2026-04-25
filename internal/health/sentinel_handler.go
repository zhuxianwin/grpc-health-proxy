package health

import (
	"encoding/json"
	"net/http"
	"time"
)

type sentinelStatusResponse struct {
	Service   string    `json:"service"`
	Tripped   bool      `json:"tripped"`
	TrippedAt time.Time `json:"tripped_at,omitempty"`
	Consec    int       `json:"consecutive_unhealthy"`
}

// NewSentinelStatusHandler returns an HTTP handler that reports the internal
// state of a sentinelChecker. It responds with 200 and a JSON body.
func NewSentinelStatusHandler(c Checker) http.Handler {
	s, ok := c.(*sentinelChecker)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if !ok {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "not a sentinel checker"})
			return
		}

		s.mu.Lock()
		defer s.mu.Unlock()

		var out []sentinelStatusResponse
		for svc, consec := range s.consec {
			entry := sentinelStatusResponse{
				Service: svc,
				Consec:  consec,
			}
			if t, tripped := s.trippedAt[svc]; tripped {
				entry.Tripped = true
				entry.TrippedAt = t
			}
			out = append(out, entry)
		}
		if out == nil {
			out = []sentinelStatusResponse{}
		}
		_ = json.NewEncoder(w).Encode(out)
	})
}
