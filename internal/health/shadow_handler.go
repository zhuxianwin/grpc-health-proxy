package health

import (
	"encoding/json"
	"net/http"
)

type shadowStatusResponse struct {
	SampleRate  float64 `json:"sample_rate"`
	TimeoutSecs float64 `json:"timeout_seconds"`
}

// NewShadowStatusHandler returns an HTTP handler that reports the current
// shadow-checker configuration. It returns 404 if the provided Checker is not
// a *shadowChecker.
func NewShadowStatusHandler(c Checker) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sc, ok := c.(*shadowChecker)
		if !ok {
			http.Error(w, "not a shadow checker", http.StatusNotFound)
			return
		}
		resp := shadowStatusResponse{
			SampleRate:  sc.cfg.SampleRate,
			TimeoutSecs: sc.cfg.Timeout.Seconds(),
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	})
}
