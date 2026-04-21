package health

import (
	"encoding/json"
	"net/http"
)

// trendSnapshot is the JSON shape returned by the trend status handler.
type trendSnapshot struct {
	Service string            `json:"service"`
	Trend   string            `json:"trend"`
	Meta    map[string]string `json:"meta,omitempty"`
}

// NewTrendStatusHandler returns an http.Handler that reports the latest trend
// annotation for a named service by performing a live check.
//
// The service name is read from the "service" query parameter; if absent the
// handler returns 400.
func NewTrendStatusHandler(tc *TrendChecker) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		svc := r.URL.Query().Get("service")
		if svc == "" {
			http.Error(w, "missing service parameter", http.StatusBadRequest)
			return
		}

		res := tc.Check(r.Context(), svc)

		trendVal := "unknown"
		if res.Meta != nil {
			if v, ok := res.Meta["trend"]; ok {
				trendVal = v
			}
		}

		snap := trendSnapshot{
			Service: svc,
			Trend:   trendVal,
			Meta:    res.Meta,
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(snap)
	})
}
