package health

import (
	"encoding/json"
	"net/http"
)

type traceStatusResponse struct {
	Enabled bool   `json:"enabled"`
	Logger  string `json:"logger"`
}

// NewTraceStatusHandler returns an HTTP handler that reports whether trace
// logging is active on the given checker.
func NewTraceStatusHandler(c Checker) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := traceStatusResponse{Enabled: false, Logger: "none"}
		if tc, ok := c.(*traceChecker); ok {
			resp.Enabled = true
			if tc.cfg.Logger != nil {
				resp.Logger = "configured"
			}
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	})
}
