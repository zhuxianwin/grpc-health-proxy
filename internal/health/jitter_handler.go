package health

import (
	"encoding/json"
	"net/http"
	"time"
)

type jitterStatusResponse struct {
	MaxJitter string `json:"max_jitter"`
}

// NewJitterStatusHandler returns an HTTP handler that reports the active
// jitter configuration for observability / debug endpoints.
func NewJitterStatusHandler(c Checker) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		jc, ok := c.(*jitterChecker)
		if !ok {
			http.Error(w, "not a jitter checker", http.StatusBadRequest)
			return
		}
		resp := jitterStatusResponse{
			MaxJitter: jc.cfg.MaxJitter.Round(time.Microsecond).String(),
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	})
}
